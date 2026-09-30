package lint

import "os"

// osWriteFile is a tiny helper used by the task-lint tests. Keeping it in
// the test package avoids importing os everywhere.
func osWriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
