package lint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSpec_BaselineFlagSet covers AC-6 (positive case): when an
// Issue.Code is listed in .baseline.json IgnoredCodes, Spec() must
// return an Issue with Baseline == true (L3 fix — explicit field, not
// a Message prefix heuristic).
func TestSpec_BaselineFlagSet(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Base(dir)
	writeBaseline(t, dir, fmt.Sprintf(`{
  "schema_version": 1,
  "spec_name": %q,
  "ignored_issues": ["ears-etc-list"]
}`, name))
	// requirements.md contains "etc" → triggers ears-etc-list ERROR.
	require := filepath.Join(dir, "requirements.md")
	if err := os.WriteFile(require, []byte(
		"WHEN foo etc THE SYSTEM SHALL bar.\n",
	), 0o644); err != nil {
		t.Fatal(err)
	}

	var found *Issue
	for i := range Spec(dir) {
		if Spec(dir)[i].Code == "ears-etc-list" {
			f := Spec(dir)[i]
			found = &f
			break
		}
	}
	// Re-run once for the captured copy above.
	issues := Spec(dir)
	for i := range issues {
		if issues[i].Code == "ears-etc-list" {
			found = &issues[i]
		}
	}
	if found == nil {
		t.Fatal("expected ears-etc-list issue; got none")
	}
	if !found.Baseline {
		t.Errorf("expected Issue.Baseline == true (Code is in baseline); got false. Message=%q", found.Message)
	}
}

// TestSpec_BaselineFlagUnset covers AC-6 (negative case): when the
// Code is NOT in .baseline.json, Issue.Baseline must be false so
// Gate() includes the issue.
func TestSpec_BaselineFlagUnset(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Base(dir)
	writeBaseline(t, dir, fmt.Sprintf(`{
  "schema_version": 1,
  "spec_name": %q,
  "ignored_issues": []
}`, name))
	require := filepath.Join(dir, "requirements.md")
	if err := os.WriteFile(require, []byte(
		"WHEN foo etc THE SYSTEM SHALL bar.\n",
	), 0o644); err != nil {
		t.Fatal(err)
	}

	issues := Spec(dir)
	var found *Issue
	for i := range issues {
		if issues[i].Code == "ears-etc-list" {
			f := issues[i]
			found = &f
		}
	}
	if found == nil {
		t.Fatal("expected ears-etc-list issue")
	}
	if found.Baseline {
		t.Errorf("expected Issue.Baseline == false (Code not in baseline); got true")
	}
}

// TestGate_ExcludesByFieldNotMessagePrefix is the L3 regression
// guard. The pre-fix implementation matched `strings.HasPrefix(i.Message,
// "[baseline] ")` to decide whether to exclude an Issue from Gate().
// A user-authored Message starting with that string would silently
// bypass the gate. The fix reads Issue.Baseline directly. We verify
// the new behaviour end-to-end: with no baseline entry for the code,
// Gate() includes it; with a baseline entry, Gate() excludes it.
func TestGate_ExcludesByFieldNotMessagePrefix(t *testing.T) {
	dir := t.TempDir()
	require := filepath.Join(dir, "requirements.md")
	if err := os.WriteFile(require, []byte(
		"WHEN foo etc THE SYSTEM SHALL bar.\n",
	), 0o644); err != nil {
		t.Fatal(err)
	}

	// Without baseline → Gate includes ears-etc-list.
	if !gateHasCode(t, dir, "ears-etc-list") {
		t.Fatal("precondition: without baseline, Gate must include ears-etc-list")
	}

	// With baseline entry for ears-etc-list → Gate excludes it.
	name := filepath.Base(dir)
	writeBaseline(t, dir, fmt.Sprintf(`{
  "schema_version": 1,
  "spec_name": %q,
  "ignored_issues": ["ears-etc-list"]
}`, name))
	if gateHasCode(t, dir, "ears-etc-list") {
		t.Error("with baseline entry, Gate must exclude ears-etc-list")
	}
}

func gateHasCode(t *testing.T, dir, code string) bool {
	t.Helper()
	for _, i := range Gate(dir) {
		if i.Code == code {
			return true
		}
	}
	return false
}

// TestSpec_RequirementsReadError_Wrapped covers AC-4 (F1): when
// requirements.md is unreadable for a non-ErrNotExist reason
// (EACCES via chmod 000), Spec() must emit a `lint-requirements-read-error`
// Issue that wraps the underlying error — NOT a misleading
// `missing-requirements` Issue.
func TestSpec_RequirementsReadError_Wrapped(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses file mode bits; cannot simulate permission denied")
	}
	dir := t.TempDir()
	name := filepath.Base(dir)
	writeBaseline(t, dir, fmt.Sprintf(`{
  "schema_version": 1,
  "spec_name": %q,
  "ignored_issues": []
}`, name))
	req := filepath.Join(dir, "requirements.md")
	if err := os.WriteFile(req, []byte("WHEN x THE SYSTEM SHALL y.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(req, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(req, 0o644) })

	issues := Spec(dir)
	var readErr *Issue
	for i := range issues {
		switch issues[i].Code {
		case "lint-requirements-read-error":
			f := issues[i]
			readErr = &f
		case "missing-requirements":
			t.Errorf("EACCES must not be reported as missing-requirements; got Message=%q", issues[i].Message)
		}
	}
	if readErr == nil {
		t.Fatal("expected lint-requirements-read-error issue")
	}
	if !strings.Contains(strings.ToLower(readErr.Message), "permission denied") &&
		!strings.Contains(strings.ToLower(readErr.Message), "read "+req) {
		t.Errorf("expected wrapped permission-denied or read path; got %q", readErr.Message)
	}
	if readErr.Severity != SeverityError {
		t.Errorf("expected ERROR severity for read failure; got %q", readErr.Severity)
	}
}

// TestSpec_DesignStatError_Wrapped covers AC-4 for design.md: the
// Stat failure path uses errors.Is(err, fs.ErrNotExist) to distinguish
// "no design.md" (missing-design WARNING) from "can't stat" (lint-
// design-stat-error WARNING). We mock by pointing SpecDir at a
// directory tree where design.md exists but is unreadable.
func TestSpec_DesignStatError_Wrapped(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses file mode bits")
	}
	dir := t.TempDir()
	design := filepath.Join(dir, "design.md")
	if err := os.WriteFile(design, []byte("# design\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Write a minimal requirements.md so Spec() doesn't bail out earlier.
	require := filepath.Join(dir, "requirements.md")
	if err := os.WriteFile(require, []byte("WHEN x THE SYSTEM SHALL y.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(design, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(design, 0o644) })

	issues := Spec(dir)
	for _, i := range issues {
		switch i.Code {
		case "missing-design":
			t.Errorf("EACCES must not be reported as missing-design; got %q", i.Message)
		case "lint-design-stat-error":
			if !strings.Contains(i.Message, "stat ") {
				t.Errorf("expected wrapped stat error; got %q", i.Message)
			}
		}
	}
}

// TestBaselineJSONShape is a lightweight regression guard: changing
// the JSON tag (e.g. from `ignored_issues` to `ignored_codes`) breaks
// historical .baseline.json files. The test pins the public JSON
// schema so any rename is caught here.
func TestBaselineJSONShape(t *testing.T) {
	b := Baseline{SchemaVersion: BaselineSchemaVersion, IgnoredCodes: []string{"a", "b"}}
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"ignored_issues":["a","b"]`) {
		t.Errorf("baseline JSON shape changed; got %s", data)
	}
}