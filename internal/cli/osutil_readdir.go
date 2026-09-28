package cli

import (
	"io"
	"os"
)

// readDir is a tiny indirection over os.ReadDir so tests can stub it.
func readDir(name string) ([]os.DirEntry, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.ReadDir(-1)
}

// ensure io import isn't pruned (test stubs may use it).
var _ = io.EOF