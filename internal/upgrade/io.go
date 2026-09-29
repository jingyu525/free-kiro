package upgrade

import (
	"archive/tar"
	"compress/gzip"
	"io"
)

// gzipNewReader + tarNewReader wrap the stdlib constructors behind
// interface variables so tests can substitute fakes if needed. Kept
// as plain wrappers today (no test stubs needed yet — the SHA256
// verification + atomic-rename paths are the real risk surfaces).
var (
	gzipNewReader = func(b []byte) (io.ReadCloser, error) {
		return gzip.NewReader(readerOf(b))
	}
	tarNewReader = tar.NewReader
)

// readerOf returns an io.Reader over a byte slice.
func readerOf(b []byte) io.Reader {
	return &sliceReader{b: b}
}

type sliceReader struct {
	b []byte
	i int
}

func (r *sliceReader) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}