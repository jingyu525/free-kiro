package upgrade

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
)

// gzipNewReader + tarNewReader wrap the stdlib constructors behind
// interface variables so tests can substitute fakes if needed. Kept
// as plain wrappers today (no test stubs needed yet — the SHA256
// verification + atomic-rename paths are the real risk surfaces).
var (
	gzipNewReader = func(b []byte) (io.ReadCloser, error) {
		return gzip.NewReader(bytes.NewReader(b))
	}
	tarNewReader = tar.NewReader
)