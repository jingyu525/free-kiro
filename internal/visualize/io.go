package visualize

import (
	"io"
	"os"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

// openWriter returns a file writer for `path`. Empty path or `-`
// means stdout (used for CLI piping).
func openWriter(path string) (io.WriteCloser, error) {
	if path == "" || path == "-" {
		return nopWriteCloser{os.Stdout}, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, ferrors.Wrap("visualize.write", err, "create "+path)
	}
	return f, nil
}

// nopWriteCloser wraps an io.Writer as io.WriteCloser (no-op Close).
type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }
