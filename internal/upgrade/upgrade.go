// Package upgrade implements `free-kiro upgrade`: detect the latest
// GitHub release, download the matching binary, verify its SHA256,
// atomically replace the running executable, and re-exec.
//
// Zero external dependencies — uses stdlib (net/http, os/exec, sha256)
// plus the same GitHub Releases API install.sh uses.
package upgrade

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

const (
	// GitHubRepo is the source for version checks + downloads. Must match
	// the repo used by install.sh (and GoReleaser).
	GitHubRepo = "jingyu525/free-kiro"

	// httpTimeout caps how long any single network call may take.
	httpTimeout = 30 * time.Second
)

// Plan describes what an upgrade would do — used by --check to preview
// without touching the filesystem.
type Plan struct {
	Current  string // version of the running binary (empty if unknown)
	Latest   string // latest released tag (e.g. "0.4.1")
	Same     bool   // true when current == latest
	Download string // full URL of the binary tarball
	Checksum string // full URL of the SHA256SUMS file
	Binary   string // basename of the tarball inside the archive
	Target   string // path the new binary would be written to
}

// Check fetches the latest release info and returns a Plan describing
// what `upgrade` (or `upgrade --check`) would do. No filesystem writes
// happen here.
func Check(ctx context.Context, currentVersion string) (*Plan, error) {
	release, err := FetchLatestRelease(ctx)
	if err != nil {
		return nil, err
	}
	if release.TagName == "" {
		return nil, ferrors.New("upgrade.check", "latest release has no tag_name")
	}
	latest := strings.TrimPrefix(release.TagName, "v")
	tarball, sums, binary := findAssets(release.TagName)
	exe, err := os.Executable()
	if err != nil {
		return nil, ferrors.Wrap("upgrade.check", err, "locate current executable")
	}
	// Resolve any symlinks so the plan points at the real file path.
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	return &Plan{
		Current:  currentVersion,
		Latest:   latest,
		Same:     currentVersion != "" && currentVersion == latest,
		Download: tarball,
		Checksum: sums,
		Binary:   binary,
		Target:   exe,
	}, nil
}

// Apply downloads the new tarball, verifies SHA256, extracts the
// matching binary, atomically replaces `target`, and re-execs the
// current process so the user runs the new version immediately.
//
// Pass force=true to skip the same-version check (useful for re-
// installing after a botched upgrade or recovering from corruption).
//
// On platforms that don't support process re-exec (Windows), Apply
// returns without re-spawning; the user is expected to re-run manually.
func Apply(ctx context.Context, p *Plan, force bool) error {
	if !force && p.Same {
		return ferrors.New("upgrade.apply",
			fmt.Sprintf("already on v%s; use --force to reinstall", p.Latest))
	}
	// Download tarball + SHA256SUMS.
	tarball, err := Download(ctx, p.Download)
	if err != nil {
		return err
	}
	sumsFile, err := Download(ctx, p.Checksum)
	if err != nil {
		return err
	}
	expected, err := LookupSHA256(string(sumsFile), p.Binary)
	if err != nil {
		return err
	}
	got := sha256.Sum256(tarball)
	if hex.EncodeToString(got[:]) != expected {
		return ferrors.New("upgrade.apply",
			"SHA256 mismatch — refusing to install (downloaded "+p.Download+")")
	}

	// Extract the binary to a temp path next to the target.
	dir := filepath.Dir(p.Target)
	tmp, err := extractBinary(tarball, p.Binary, dir)
	if err != nil {
		return err
	}

	// Atomic rename (same filesystem → atomic on POSIX).
	if err := os.Rename(tmp, p.Target); err != nil {
		_ = os.Remove(tmp)
		return ferrors.Wrap("upgrade.apply", err, "rename "+tmp+" → "+p.Target)
	}
	// Preserve executable bit (rename does on POSIX but be defensive).
	_ = os.Chmod(p.Target, 0o755)

	// Re-exec — replaces the current process with the new binary.
	return reexec(p.Target)
}

// reexec replaces the current process with the given executable.
// Returns a non-nil error on platforms where this isn't supported
// (Windows); the caller falls back to instructing the user to
// restart manually.
func reexec(bin string) error {
	if runtime.GOOS == "windows" {
		return ferrors.New("upgrade.reexec",
			"re-exec not supported on Windows; please restart manually")
	}
	cmd := exec.Command(bin, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		// exec returned — child finished (or errored). Propagate the
		// child's exit status so the shell sees the right code.
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return ferrors.Wrap("upgrade.reexec", err, "re-exec "+bin)
	}
	// Child exited 0 — exit the parent cleanly.
	os.Exit(0)
	return nil
}

// --- GitHub API + tarball extraction ---

// Release is the public subset of the GitHub release JSON we care about.
// Exported so other packages (e.g. internal/skill) can reuse the fetch
// without re-parsing the same struct.
type Release struct {
	TagName string         `json:"tag_name"`
	Assets  []ReleaseAsset `json:"assets"`
}

// ReleaseAsset is one entry in the release's `assets` list.
type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// FetchLatestRelease queries GitHub for the latest published tag.
func FetchLatestRelease(ctx context.Context) (*Release, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://api.github.com/repos/"+GitHubRepo+"/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, ferrors.Wrap("upgrade.check", err, "GET releases/latest")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, ferrors.New("upgrade.check",
			fmt.Sprintf("GitHub API returned HTTP %d (rate limit?)", resp.StatusCode))
	}
	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, ferrors.Wrap("upgrade.check", err, "decode release JSON")
	}
	return &rel, nil
}

// findAssets picks the tarball + SHA256SUMS URL for the current
// platform/arch from the release's assets list.
func findAssets(tag string) (tarball, sums, binary string) {
	platform := runtime.GOOS + "_" + runtime.GOARCH
	tarball = fmt.Sprintf("https://github.com/%s/releases/download/v%s/free-kiro_%s_%s.tar.gz",
		GitHubRepo, strings.TrimPrefix(tag, "v"), strings.TrimPrefix(tag, "v"), platform)
	sums = fmt.Sprintf("https://github.com/%s/releases/download/v%s/free-kiro_%s_SHA256SUMS",
		GitHubRepo, strings.TrimPrefix(tag, "v"), strings.TrimPrefix(tag, "v"))
	binary = "free-kiro"
	return
}

// Download fetches a URL into memory (returns bytes). Tarballs are
// ~2-4 MB so a 32 MB cap is comfortable; SHA256SUMS is tiny. Exported
// for reuse by internal/skill.
func Download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, ferrors.Wrap("upgrade.download", err, url)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, ferrors.New("upgrade.download",
			fmt.Sprintf("%s returned HTTP %d", url, resp.StatusCode))
	}
	limited := io.LimitReader(resp.Body, 32<<20) // 32 MB cap
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, ferrors.Wrap("upgrade.download", err, url)
	}
	return data, nil
}

// LookupSHA256 finds the expected hash for `tarball` in the SHA256SUMS
// file format used by GoReleaser: each line is `<hex>  <filename>`.
// Matches by exact filename; the caller can pre-resolve the canonical
// name. Falls back to "free-kiro" for the binary inside a tarball.
func LookupSHA256(sums, tarball string) (string, error) {
	for line := range strings.SplitSeq(sums, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		hash, name := parts[0], parts[1]
		// Match either the full tarball name or the binary inside it.
		if name == tarball || name == "free-kiro" {
			return hash, nil
		}
	}
	return "", ferrors.New("upgrade.apply",
		fmt.Sprintf("no SHA256 entry found for %s in SHA256SUMS", tarball))
}

// extractBinary uncompresses `tarball` (tar.gz) and writes the
// `binary` member to a temp file in `dir` next to the target. Returns
// the temp file's path; the caller is expected to os.Rename it over
// the target.
func extractBinary(tarball []byte, binary, dir string) (string, error) {
	// Hand-rolled tar.gz extraction with stdlib only. Go's archive/tar
	// handles tar; gzip.Reader handles gzip.
	gz, err := gzipNewReader(tarball)
	if err != nil {
		return "", ferrors.Wrap("upgrade.extract", err, "open gzip")
	}
	defer gz.Close()
	tr := tarNewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", ferrors.Wrap("upgrade.extract", err, "tar next")
		}
		if hdr.Typeflag != tarTypeReg && hdr.Typeflag != tarTypeRegA {
			continue
		}
		// Match by basename (the tarball includes the binary at the
		// archive root).
		if filepath.Base(hdr.Name) != binary {
			continue
		}
		tmp := filepath.Join(dir, "."+binary+".new")
		f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return "", ferrors.Wrap("upgrade.extract", err, "open "+tmp)
		}
		if _, err := io.Copy(f, tr); err != nil {
			_ = f.Close()
			_ = os.Remove(tmp)
			return "", ferrors.Wrap("upgrade.extract", err, "write "+tmp)
		}
		if err := f.Close(); err != nil {
			_ = os.Remove(tmp)
			return "", ferrors.Wrap("upgrade.extract", err, "close "+tmp)
		}
		return tmp, nil
	}
	return "", ferrors.New("upgrade.extract",
		"binary "+binary+" not found in tarball")
}

// tar header typeflag constants (mirrored to avoid the archive/tar import).
const (
	tarTypeReg  = '0'
	tarTypeRegA = '\x00' // legacy "regular file" alias used by some tar writers
)