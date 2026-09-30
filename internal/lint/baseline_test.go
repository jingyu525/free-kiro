package lint

import (
	"os"
	"path/filepath"
	"testing"
)

// helper: write a .baseline.json with given content
func writeBaseline(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, BaselineFileName), []byte(content), 0o644); err != nil {
		t.Fatalf("write baseline: %v", err)
	}
}

// Case 1: file does not exist → zero-value Baseline, no error
func TestLoadBaseline_MissingFile(t *testing.T) {
	dir := t.TempDir()
	b, err := LoadBaseline(dir)
	if err != nil {
		t.Fatalf("expected no error for missing file; got %v", err)
	}
	if b.SchemaVersion != 0 || b.SpecName != "" || len(b.IgnoredCodes) != 0 {
		t.Errorf("missing file should return zero-value baseline; got %+v", b)
	}
}

// Case 2: empty file → error (parse failure)
func TestLoadBaseline_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	writeBaseline(t, dir, "")
	if _, err := LoadBaseline(dir); err == nil {
		t.Error("empty baseline file should error on parse")
	}
}

// Case 3: valid baseline — load + ShouldIgnore + Empty
func TestLoadBaseline_Valid(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Base(dir)
	writeBaseline(t, dir, `{
  "schema_version": 1,
  "spec_name": "`+name+`",
  "ignored_issues": ["ears-few-ac", "ears-low-template-diversity"]
}`)
	b, err := LoadBaseline(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.SchemaVersion != 1 {
		t.Errorf("schema_version: got %d want 1", b.SchemaVersion)
	}
	if b.SpecName != name {
		t.Errorf("spec_name: got %q want %q", b.SpecName, name)
	}
	if !b.ShouldIgnore("ears-few-ac") {
		t.Error("ShouldIgnore(ears-few-ac) should be true")
	}
	if !b.ShouldIgnore("ears-low-template-diversity") {
		t.Error("ShouldIgnore(ears-low-template-diversity) should be true")
	}
	if b.ShouldIgnore("no-ears") {
		t.Error("ShouldIgnore(no-ears) should be false (not in baseline)")
	}
}

// Case 4: schema_version mismatch → error
func TestLoadBaseline_WrongSchemaVersion(t *testing.T) {
	dir := t.TempDir()
	writeBaseline(t, dir, `{
  "schema_version": 99,
  "ignored_issues": ["x"]
}`)
	_, err := LoadBaseline(dir)
	if err == nil {
		t.Fatal("expected schema_version mismatch error")
	}
	if got, want := err.Error(), "schema_version 99"; !contains(got, want) {
		t.Errorf("error should mention %q; got %q", want, got)
	}
}

// Case 5: spec_name mismatch → error
func TestLoadBaseline_SpecNameMismatch(t *testing.T) {
	dir := t.TempDir()
	writeBaseline(t, dir, `{
  "schema_version": 1,
  "spec_name": "some-other-spec",
  "ignored_issues": []
}`)
	_, err := LoadBaseline(dir)
	if err == nil {
		t.Fatal("expected spec_name mismatch error")
	}
	if got, want := err.Error(), "some-other-spec"; !contains(got, want) {
		t.Errorf("error should mention the wrong name %q; got %q", want, got)
	}
}

// Case 6: spec_name omitted → accepted (uses directory basename implicitly)
func TestLoadBaseline_SpecNameOmitted(t *testing.T) {
	dir := t.TempDir()
	writeBaseline(t, dir, `{
  "schema_version": 1,
  "ignored_issues": ["ears-few-ac"]
}`)
	b, err := LoadBaseline(dir)
	if err != nil {
		t.Fatalf("omitted spec_name should be accepted; got %v", err)
	}
	if !b.ShouldIgnore("ears-few-ac") {
		t.Error("ShouldIgnore(ears-few-ac) should be true")
	}
}

// Case 7: malformed JSON → error
func TestLoadBaseline_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	writeBaseline(t, dir, `{not valid json`)
	if _, err := LoadBaseline(dir); err == nil {
		t.Error("malformed JSON should error")
	}
}

// Case 8: ShouldIgnore on zero-value baseline → false for everything
func TestBaseline_ZeroValueShouldIgnore(t *testing.T) {
	var b Baseline
	if b.ShouldIgnore("anything") {
		t.Error("zero-value Baseline should ignore nothing")
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
