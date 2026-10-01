package skill

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/upgrade"
)

// InstallOptions configures one InstallOne invocation.
type InstallOptions struct {
	App     App    // target app
	Home    string // user home (empty = use $FREE_KIRO_HOME or os.UserHomeDir)
	Subdir  string // bundle name inside skills/ (default "free-kiro")
	Source  string // path or URL of bundle (empty = latest from GitHub)
	Version string // explicit version (e.g. "0.7.0"); used when Source is empty
	DryRun  bool   // print plan, write nothing
	Force   bool   // overwrite existing install without prompting
}

// InstallResult is the per-app outcome of an install (or dry-run).
type InstallResult struct {
	App          App
	SkillsDir    string
	Version      string
	FilesWritten int
	SHA256OK     bool
	Status       string // "installed" | "updated" | "already-current" | "dry-run" | "skipped"
	Err          error
}

// InstallOne resolves the bundle (from Source or latest release), verifies
// sha256, and writes files into `~/.{app}/skills/<subdir>/`. The write
// is non-destructive unless Force is set or the destination doesn't yet
// exist; otherwise the existing version is compared and the result
// surfaces in `result.Status`.
func InstallOne(ctx context.Context, opts InstallOptions) InstallResult {
	res := InstallResult{App: opts.App}
	opts.Subdir = resolveSubdir(opts.Subdir)
	h, err := resolveHome(opts.Home)
	if err != nil {
		res.Err = err
		return res
	}
	opts.Home = h
	res.SkillsDir = SkillsDir(opts.App, opts.Home, opts.Subdir)

	// Detect existing install.
	existingVersion := ""
	if _, err = os.Stat(filepath.Join(res.SkillsDir, "skill.json")); err == nil {
		// Existing install — read its version.
		var m *Manifest
		if m, err = LoadManifest(res.SkillsDir); err == nil {
			existingVersion = m.Version
		}
	}
	res.Version = opts.Version

	if opts.DryRun {
		res.Status = "dry-run"
		res.FilesWritten = 0
		return res
	}

	if existingVersion != "" && existingVersion == opts.Version && !opts.Force {
		res.Status = "already-current"
		return res
	}

	// Resolve the bundle.
	bundleDir, bundleOwned, err := resolveBundle(ctx, opts)
	if err != nil {
		res.Err = err
		return res
	}
	if bundleOwned {
		// Only remove dirs we created (temp extraction dirs). When
		// Source is a local path, resolveBundle returns it as-is — it
		// belongs to the user and must not be deleted.
		defer func() { _ = os.RemoveAll(bundleDir) }()
	}

	// Verify sha256 if manifest has them populated.
	m, err := LoadManifest(bundleDir)
	if err != nil {
		res.Err = err
		return res
	}
	if failures, _ := m.VerifyBundle(bundleDir); len(failures) > 0 {
		res.Err = ferrors.New("skill.install",
			"sha256 verification failed for: "+fmt.Sprint(failures))
		return res
	}
	res.SHA256OK = true

	// Copy bundle files into the target dir.
	if err = os.MkdirAll(res.SkillsDir, 0o755); err != nil {
		res.Err = ferrors.Wrap("skill.install", err, "mkdir "+res.SkillsDir)
		return res
	}
	count, err := copyTree(bundleDir, res.SkillsDir)
	if err != nil {
		res.Err = ferrors.Wrap("skill.install", err, "copy bundle → "+res.SkillsDir)
		return res
	}
	res.FilesWritten = count
	if existingVersion == "" {
		res.Status = "installed"
	} else {
		res.Status = "updated"
	}
	return res
}

// UninstallOne removes the installed bundle directory for `app`. Returns
// the list of files removed (best-effort).
func UninstallOne(app App, home, subdir string) (removed []string, err error) {
	subdir = resolveSubdir(subdir)
	home, err = resolveHome(home)
	if err != nil {
		return nil, err
	}
	dir := SkillsDir(app, home, subdir)
	if dir == "" {
		return nil, ferrors.New("skill.uninstall", "unsupported app: "+string(app))
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil // not installed → idempotent no-op
	}
	if err := os.RemoveAll(dir); err != nil {
		return nil, ferrors.Wrap("skill.uninstall", err, "remove "+dir)
	}
	return []string{dir}, nil
}

// resolveBundle returns the bundle's extracted directory and an `owned`
// flag indicating whether the caller may remove it. When opts.Source is
// empty or a zip URL, the bundle is extracted into a fresh temp dir
// (owned=true, safe to RemoveAll). When opts.Source is a local path,
// that path is returned as-is (owned=false) — deleting it would destroy
// the user's source files.
func resolveBundle(ctx context.Context, opts InstallOptions) (string, bool, error) {
	if opts.Source == "" {
		// Latest release path.
		version := opts.Version
		if version == "" {
			v, err := LatestRelease(ctx)
			if err != nil {
				return "", false, err
			}
			version = v
		}
		opts.Version = version
		dest, err := os.MkdirTemp("", "free-kiro-skill-*")
		if err != nil {
			return "", false, err
		}
		if err := DownloadAndExtract(ctx, version, dest); err != nil {
			_ = os.RemoveAll(dest)
			return "", false, err
		}
		return dest, true, nil
	}
	// Source given: treat as a local path or zip URL.
	if isZipURL(opts.Source) {
		dest, err := os.MkdirTemp("", "free-kiro-skill-*")
		if err != nil {
			return "", false, err
		}
		body, err := upgrade.Download(ctx, opts.Source)
		if err != nil {
			_ = os.RemoveAll(dest)
			return "", false, err
		}
		if err := unzip(body, dest); err != nil {
			_ = os.RemoveAll(dest)
			return "", false, ferrors.Wrap("skill.install", err, "extract "+opts.Source)
		}
		return dest, true, nil
	}
	// Local path: must already be an extracted directory.
	if st, err := os.Stat(opts.Source); err == nil && st.IsDir() {
		return opts.Source, false, nil
	}
	return "", false, ferrors.New("skill.install",
		"unsupported --from value: "+opts.Source+" (must be local dir or http(s) zip URL)")
}

func isZipURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// copyTree recursively copies files from src to dst. Returns the count
// of files copied. dst is created if missing.
func copyTree(src, dst string) (int, error) {
	count := 0
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := copyFile(path, target); err != nil {
			return err
		}
		count++
		return nil
	})
	return count, err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	_, err = io.Copy(out, in)
	return err
}
