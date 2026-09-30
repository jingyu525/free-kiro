package lint

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// BaselineFileName is the filename inside a spec directory that holds the
// baseline JSON. Authors place this alongside requirements.md / tasks.md
// to declare "this spec knowingly accepts the listed lint codes — do not
// block advance/approve on them." The file is optional; absence is the
// zero-value baseline (ignore nothing).
const BaselineFileName = ".baseline.json"

// BaselineSchemaVersion is the schema version this build of free-kiro
// reads / writes. Bump this when the JSON shape changes in a way that
// cannot be auto-detected.
const BaselineSchemaVersion = 1

// Baseline captures the per-spec lint whitelist. It is loaded from
// `<specDir>/.baseline.json` and consulted by Spec() / Gate() to decide
// which issues still count toward the advance/approve block.
//
// Fields:
//   - SchemaVersion: integer; mismatch → error (no silent fallthrough).
//   - SpecName:      string; must match the directory basename when set.
//                     Mismatch → error (catches copy-paste mistakes).
//   - IgnoredCodes:  list of Issue.Code strings to skip in Gate() while
//                     still emitting them in `free-kiro lint` output
//                     (prefixed with `[baseline]`).
//
// Empty IgnoredCodes is a valid baseline — it lets the file exist so
// historical specs don't have to invent ignore entries just to opt into
// the baseline mechanism.
type Baseline struct {
	SchemaVersion int      `json:"schema_version"`
	SpecName      string   `json:"spec_name,omitempty"`
	IgnoredCodes  []string `json:"ignored_issues"`
}

// LoadBaseline reads `<specDir>/.baseline.json`. Missing file → zero-value
// Baseline (no error). Malformed file, wrong schema_version, or mismatched
// spec_name → descriptive error. Callers (Spec / Gate) translate the
// error into a `baseline-parse-error` Issue so the user sees the
// underlying message.
func LoadBaseline(specDir string) (Baseline, error) {
	path := filepath.Join(specDir, BaselineFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Baseline{}, nil
		}
		return Baseline{}, fmt.Errorf("read %s: %w", path, err)
	}
	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return Baseline{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if b.SchemaVersion != BaselineSchemaVersion {
		return Baseline{}, fmt.Errorf("%s schema_version %d not supported (want %d)",
			BaselineFileName, b.SchemaVersion, BaselineSchemaVersion)
	}
	if b.SpecName != "" {
		expected := filepath.Base(specDir)
		if b.SpecName != expected {
			return Baseline{}, fmt.Errorf("%s spec_name %q does not match directory %q",
				BaselineFileName, b.SpecName, expected)
		}
	}
	return b, nil
}

// ShouldIgnore reports whether the given Issue.Code is in this baseline's
// IgnoredCodes set. Unknown codes are silently ignored (forward
// compatibility: a future free-kier version can add new Issue.Codes
// without invalidating older baseline files).
func (b Baseline) ShouldIgnore(code string) bool {
	return slices.Contains(b.IgnoredCodes, code)
}