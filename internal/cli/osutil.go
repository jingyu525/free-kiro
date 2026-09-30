package cli

import (
	"os"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

// osStat / osWriteFile are the real implementations of stat / writeFile.
// They live in their own file so tests can swap them via the package-level
// variables declared in init.go.
func osStat(p string) (any, error) {
	info, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func osWriteFile(p, content string) error {
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		return ferrors.Wrap("cli.write", err, "write "+p)
	}
	return nil
}
