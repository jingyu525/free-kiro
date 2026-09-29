// Package watch provides a debounced file-system watcher that runs a
// caller-supplied command whenever files under `.kiro/` change.
//
// Powers `free-kiro watch` — the IDE-style "always show lint errors"
// experience without leaving the terminal.
//
// Why debounce: editors (vim, vscode, idea) often write files in two
// steps (truncate + write) and produce several fs events per logical
// save. We coalesce them into one command invocation per quiet period.
package watch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

// Options configures a Watcher.
type Options struct {
	// Roots to watch. Defaults to .kiro/ when empty.
	Roots []string
	// Command is the shell snippet executed on every quiet-period fire.
	// Required.
	Command string
	// Debounce is the quiet period after the last change before
	// triggering the command. Default 200ms.
	Debounce time.Duration
	// Verbose logs every event the watcher sees (useful for debugging
	// path filters). Default false.
	Verbose bool
}

// Watcher runs the file-watcher loop. Call Run to block until ctx is
// done or the watcher fails irrecoverably.
type Watcher struct {
	opts Options
}

// New constructs a Watcher with defaults applied.
func New(opts Options) (*Watcher, error) {
	if opts.Command == "" {
		return nil, ferrors.New("watch.new", "Options.Command is required")
	}
	if len(opts.Roots) == 0 {
		return nil, ferrors.New("watch.new", "Options.Roots must include at least one directory")
	}
	if opts.Debounce == 0 {
		opts.Debounce = 200 * time.Millisecond
	}
	for i, r := range opts.Roots {
		abs, err := filepath.Abs(r)
		if err != nil {
			return nil, ferrors.Wrap("watch.new", err, "resolve "+r)
		}
		opts.Roots[i] = abs
	}
	return &Watcher{opts: opts}, nil
}

// Run blocks until ctx is done. For every quiet period (no events for
// Debounce) the command is exec'd. Returns ctx.Err() on cancellation,
// or the first irrecoverable watcher error otherwise.
func (w *Watcher) Run(ctx context.Context) error {
	fs, err := fsnotify.NewWatcher()
	if err != nil {
		return ferrors.Wrap("watch.run", err, "create watcher")
	}
	defer fs.Close()

	// Add every root + all its subdirectories (fsnotify doesn't
	// recurse on its own).
	for _, root := range w.opts.Roots {
		if err := addTree(fs, root); err != nil {
			return ferrors.Wrap("watch.run", err, "watch "+root)
		}
	}
	if w.opts.Verbose {
		fmt.Fprintf(stderr(), "[watch] watching %s (debounce=%s)\n", strings.Join(w.opts.Roots, ", "), w.opts.Debounce)
	}

	var (
		timer    *time.Timer
		timerCh  <-chan time.Time
		lastPath string
	)

	for {
		select {
		case <-ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return ctx.Err()
		case ev, ok := <-fs.Events:
			if !ok {
				return ferrors.New("watch.run", "fsnotify events channel closed")
			}
			if !shouldReact(ev) {
				continue
			}
			lastPath = ev.Name
			if w.opts.Verbose {
				fmt.Fprintf(stderr(), "[watch] %s %s\n", ev.Op, ev.Name)
			}
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(w.opts.Debounce)
			timerCh = timer.C
		case <-timerCh:
			timerCh = nil
			if err := w.runCommand(ctx, lastPath); err != nil {
				fmt.Fprintf(stderr(), "[watch] command error: %v\n", err)
			}
		case err, ok := <-fs.Errors:
			if !ok {
				return ferrors.New("watch.run", "fsnotify errors channel closed")
			}
			fmt.Fprintf(stderr(), "[watch] fsnotify error: %v\n", err)
		}
	}
}

// runCommand executes the configured shell snippet with the changed
// file's path exported as $FILE. Returns an error if the command
// itself failed (non-zero exit) — but the watcher keeps running for
// future events.
func (w *Watcher) runCommand(ctx context.Context, file string) error {
	fmt.Fprintf(stderr(), "[watch] change → %s\n", runLabel(file))
	cmd := exec.CommandContext(ctx, "sh", "-c", w.opts.Command)
	cmd.Env = append(cmd.Environ(), "FILE="+file)
	cmd.Stdout = stderr() // echo output so the user sees it inline
	cmd.Stderr = stderr()
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ferrors.New("watch.cmd",
				fmt.Sprintf("command exited with code %d", ee.ExitCode()))
		}
		return err
	}
	return nil
}

// shouldReact filters events: ignore chmod / non-write events and
// anything outside .kiro/ (defensive — we only added .kiro/ to the
// watcher, but symlinks could surprise us).
func shouldReact(ev fsnotify.Event) bool {
	switch ev.Op {
	case fsnotify.Write, fsnotify.Create, fsnotify.Remove, fsnotify.Rename:
		return true
	}
	return false
}

// addTree adds root + every subdirectory to the watcher.
func addTree(fs *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		// Skip noisy editor / VCS directories.
		name := filepath.Base(path)
		if name == ".git" || name == "node_modules" || (strings.HasPrefix(name, ".") && path != root) {
			return filepath.SkipDir
		}
		return fs.Add(path)
	})
}

// runLabel returns a short "name (op)" string for log output.
func runLabel(path string) string {
	return filepath.Base(path)
}

// stderr returns os.Stderr. Indirected so tests can swap if needed.
var stderr = func() io.Writer { return os.Stderr }