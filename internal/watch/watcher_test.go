package watch

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestNew_RequiresCommand(t *testing.T) {
	_, err := New(Options{Roots: []string{"."}})
	if err == nil {
		t.Fatal("expected error when Command is empty")
	}
}

func TestNew_RequiresRoots(t *testing.T) {
	_, err := New(Options{Command: "echo"})
	if err == nil {
		t.Fatal("expected error when Roots is empty")
	}
}

func TestNew_AppliesDefaultDebounce(t *testing.T) {
	w, err := New(Options{Command: "echo", Roots: []string{"."}})
	if err != nil {
		t.Fatal(err)
	}
	if w.opts.Debounce != 200*time.Millisecond {
		t.Errorf("default debounce should be 200ms; got %s", w.opts.Debounce)
	}
}

func TestNew_ResolvesAbsolute(t *testing.T) {
	dir := t.TempDir()
	w, err := New(Options{Command: "echo", Roots: []string{dir}})
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(w.opts.Roots[0]) {
		t.Errorf("root should be absolute; got %q", w.opts.Roots[0])
	}
}

func TestShouldReact(t *testing.T) {
	cases := []struct {
		op   fsnotify.Op
		want bool
	}{
		{fsnotify.Write, true},
		{fsnotify.Create, true},
		{fsnotify.Remove, true},
		{fsnotify.Rename, true},
		{fsnotify.Chmod, false},
	}
	for _, c := range cases {
		if got := shouldReact(fsnotify.Event{Op: c.op}); got != c.want {
			t.Errorf("shouldReact(%v) = %v, want %v", c.op, got, c.want)
		}
	}
}

func TestRun_CommandFiresOnChange(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "fired")
	w, err := New(Options{
		Command:  "touch " + marker,
		Roots:    []string{dir},
		Debounce: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Write a file in the watched dir from another goroutine so the
	// watcher picks up the event.
	go func() {
		time.Sleep(150 * time.Millisecond)
		_ = os.WriteFile(filepath.Join(dir, "trigger.txt"), []byte("x"), 0o644)
	}()
	_ = w.Run(ctx)
	if _, err := os.Stat(marker); err != nil {
		t.Errorf("command did not fire: %v", err)
	}
}