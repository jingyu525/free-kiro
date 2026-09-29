package skill

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/upgrade"
)

// GitHubRepo is the upstream source for skill bundles. Alias for the
// upgrade package's constant; kept here for callers that already import
// internal/skill.
const GitHubRepo = upgrade.GitHubRepo

// AssetName is the released zip name. GoReleaser builds it from the
// `skill` archive id; we mirror that template here so updates stay in
// sync.
func AssetName(version string) string {
	v := strings.TrimPrefix(version, "v")
	return "free-kiro-skill_" + v + ".zip"
}

// releaseFileURL returns the GitHub Releases download URL for a file
// pinned to `version` (e.g. "0.7.0" or "v0.7.0"). Both AssetURL and
// SHA256SUMSURL build on top of this so the URL template lives in one
// place.
func releaseFileURL(version, name string) string {
	v := strings.TrimPrefix(version, "v")
	return fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s",
		upgrade.GitHubRepo, v, name)
}

// AssetURL returns the GitHub Releases URL for the skill bundle zip.
func AssetURL(version string) string {
	return releaseFileURL(version, AssetName(version))
}

// SHA256SUMSURL returns the GitHub Releases URL for the SHA256SUMS file.
func SHA256SUMSURL(version string) string {
	v := strings.TrimPrefix(version, "v")
	return releaseFileURL(version, fmt.Sprintf("free-kiro_%s_SHA256SUMS", v))
}

// LatestRelease returns the latest published tag (no "v" prefix).
func LatestRelease(ctx context.Context) (string, error) {
	rel, err := upgrade.FetchLatestRelease(ctx)
	if err != nil {
		return "", ferrors.Wrap("skill.release", err, "fetch latest")
	}
	if rel.TagName == "" {
		return "", ferrors.New("skill.release", "latest release has no tag_name")
	}
	return strings.TrimPrefix(rel.TagName, "v"), nil
}

// DownloadAndExtract downloads the skill zip for `version` into a temp
// directory, verifies its sha256 against SHA256SUMS, then extracts it
// into `dest`. The zip + SHA256SUMS fetches run in parallel to save one
// RTT on cold cache.
//
// If `version` is empty, resolves to the latest release first.
func DownloadAndExtract(ctx context.Context, version, dest string) error {
	if version == "" {
		v, err := LatestRelease(ctx)
		if err != nil {
			return err
		}
		version = v
	}

	zipURL := AssetURL(version)
	sumsURL := SHA256SUMSURL(version)

	// Parallel fetch: SHA256SUMS is ~few hundred bytes and is fetched
	// during the zip transfer, saving one full RTT.
	type result struct {
		data []byte
		err  error
	}
	zipCh := make(chan result, 1)
	sumsCh := make(chan result, 1)
	go func() { d, e := upgrade.Download(ctx, zipURL); zipCh <- result{d, e} }()
	go func() { d, e := upgrade.Download(ctx, sumsURL); sumsCh <- result{d, e} }()
	zr := <-zipCh
	sr := <-sumsCh
	if zr.err != nil {
		return zr.err
	}
	if sr.err != nil {
		return sr.err
	}
	want, err := upgrade.LookupSHA256(string(sr.data), AssetName(version))
	if err != nil {
		return ferrors.New("skill.release",
			"sha256 entry missing in SHA256SUMS — bundle "+AssetName(version)+" not published for v"+version)
	}
	if got := sha256Hex(zr.data); got != want {
		return ferrors.New("skill.release",
			fmt.Sprintf("sha256 mismatch for %s: got %s, want %s", AssetName(version), got, want))
	}

	// Extract to dest.
	if err := os.RemoveAll(dest); err != nil {
		return ferrors.Wrap("skill.release", err, "clean "+dest)
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return ferrors.Wrap("skill.release", err, "mkdir "+dest)
	}
	if err := unzip(zr.data, dest); err != nil {
		return ferrors.Wrap("skill.release", err, "extract "+zipURL)
	}
	return nil
}

// sha256Hex returns the hex sha256 of data.
func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func unzip(data []byte, dest string) error {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	for _, f := range r.File {
		// Files are written under dest/<name>. The GoReleaser archive
		// prefixes with "free-kiro/" — strip that single prefix so the
		// final layout is dest/SKILL.md, dest/references/*.md, etc.
		name := f.Name
		name = strings.TrimPrefix(name, "free-kiro/")
		name = strings.TrimPrefix(name, "./")
		if name == "" || strings.Contains(name, "..") {
			continue // safety: skip empty / path-traversal entries
		}
		target := filepath.Join(dest, name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		src, err := f.Open()
		if err != nil {
			return err
		}
		dst, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			src.Close()
			return err
		}
		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			dst.Close()
			return err
		}
		src.Close()
		dst.Close()
	}
	return nil
}