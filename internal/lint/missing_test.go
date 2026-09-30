package lint

import (
	"path/filepath"
	"strings"
	"testing"
)

func hasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// --- Requirements() paths ---

func TestLintRequirements_PlaceholderAC(t *testing.T) {
	// Doc has EARS-shaped content but it contains a <TODO:...> placeholder,
	// so the EARS regex matches yet the placeholder is the real problem.
	doc := `# My Feature

## User Stories
As a user I want X so that Y.

## Acceptance Criteria
WHEN user clicks <TODO:button> THE SYSTEM SHALL <TODO:response>.
`
	issues := Requirements(doc)
	var found bool
	for _, i := range issues {
		if i.Code == "placeholder-ac" {
			found = true
			if i.Severity != SeverityError {
				t.Errorf("placeholder-ac should be ERROR; got %q", i.Severity)
			}
		}
	}
	if !found {
		t.Errorf("expected placeholder-ac ERROR; got %v", issues)
	}
}

// --- Bugfix() paths ---

func TestLintBugfix_MissingCurrent(t *testing.T) {
	doc := `## Expected Behavior
WHEN user submits valid creds THE SYSTEM SHALL redirect to /dashboard.

## Unchanged Behavior
WHEN auth fails THE SYSTEM SHALL CONTINUE TO show error.
`
	issues := Bugfix(doc)
	var found bool
	for _, i := range issues {
		if i.Code == "missing-current" {
			found = true
			if i.Severity != SeverityWarning {
				t.Errorf("missing-current should be WARNING; got %q", i.Severity)
			}
		}
	}
	if !found {
		t.Errorf("expected missing-current WARNING; got %v", issues)
	}
}

// --- Spec() document-level paths ---
//
// Spec() emits the `missing-<stem>` findings when files are absent. Each
// test creates a minimal fake spec dir and asserts which codes appear.
// Uses the writeFile helper from tasks_test.go (takes a full path).

func TestSpec_MissingRequirements(t *testing.T) {
	dir := t.TempDir()
	// no requirements.md, only tasks.md (so requirements is the only miss)
	writeFile(t, filepath.Join(dir, "tasks.md"), "- [ ] #1 task")
	issues := Spec(dir)
	assertHasCode(t, issues, "missing-requirements", SeverityError)
}

func TestSpec_MissingBugfix(t *testing.T) {
	dir := t.TempDir()
	// fake a bugfix spec meta
	writeFile(t, filepath.Join(dir, ".meta.json"), `{
  "spec_type": "bugfix",
  "phase": "draft"
}`)
	// no bugfix.md, only tasks.md
	writeFile(t, filepath.Join(dir, "tasks.md"), "- [ ] #1 task")
	issues := Spec(dir)
	assertHasCode(t, issues, "missing-bugfix", SeverityError)
}

func TestSpec_MissingTasks(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "requirements.md"), "# Feature\n\nWHEN foo THE SYSTEM SHALL bar.\n")
	// no tasks.md, no design.md
	issues := Spec(dir)
	assertHasCode(t, issues, "missing-tasks", SeverityWarning)
	assertHasCode(t, issues, "missing-design", SeverityWarning)
}

func TestSpec_MissingDesign(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "requirements.md"), "# Feature\n\nWHEN foo THE SYSTEM SHALL bar.\n")
	writeFile(t, filepath.Join(dir, "tasks.md"), "- [ ] #1 task")
	// no design.md
	issues := Spec(dir)
	assertHasCode(t, issues, "missing-design", SeverityWarning)
	// tasks.md exists, so missing-tasks should NOT appear
	assertNoCode(t, issues, "missing-tasks")
}

// --- Gate() behavior on missing-* ---

func TestGate_ExcludesMissingCodes(t *testing.T) {
	dir := t.TempDir()
	// Only tasks.md present → requirements.md and design.md missing.
	// Gate() must not include missing-* as blocking findings, otherwise
	// the very first `generate` (nothing written) could never advance.
	writeFile(t, filepath.Join(dir, "tasks.md"), "- [ ] #1 task")
	gate := Gate(dir)
	for _, i := range gate {
		if isMissingCode(i.Code) {
			t.Errorf("Gate() must exclude missing-* findings; got %s", i)
		}
	}
}

// --- Baseline integration with Spec() / Gate() / SpecStrict() ---

func TestSpec_AppliesBaseline(t *testing.T) {
	dir := t.TempDir()
	// Doc with no User Stories → no-user-stories ERROR
	writeFile(t, filepath.Join(dir, "requirements.md"), "# F\n\n## Acceptance Criteria\nWHEN foo THE SYSTEM SHALL bar.\n")
	// Baseline whitelists no-user-stories + the etc-list warning
	name := filepath.Base(dir)
	writeFile(t, filepath.Join(dir, BaselineFileName), `{
  "schema_version": 1,
  "spec_name": "`+name+`",
  "ignored_issues": ["no-user-stories"]
}`)
	issues := Spec(dir)
	// no-user-stories is in baseline → prefixed
	var foundPrefixed bool
	for _, i := range issues {
		if i.Code == "no-user-stories" && hasPrefix(i.Message, "[baseline]") {
			foundPrefixed = true
		}
	}
	if !foundPrefixed {
		t.Errorf("expected no-user-stories to be prefixed with [baseline]; got %v", issues)
	}
	// Gate() must NOT include it
	gate := Gate(dir)
	for _, i := range gate {
		if i.Code == "no-user-stories" {
			t.Errorf("Gate() must exclude baseline-ignored issues; got %s", i)
		}
	}
}

func TestSpecStrict_IgnoresBaseline(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "requirements.md"), "# F\n\n## Acceptance Criteria\nWHEN foo THE SYSTEM SHALL bar.\n")
	name := filepath.Base(dir)
	writeFile(t, filepath.Join(dir, BaselineFileName), `{
  "schema_version": 1,
  "spec_name": "`+name+`",
  "ignored_issues": ["no-user-stories"]
}`)
	issues := SpecStrict(dir)
	// SpecStrict does NOT apply baseline → no-user-stories has no [baseline] prefix
	var foundUnprefixed bool
	for _, i := range issues {
		if i.Code == "no-user-stories" && !hasPrefix(i.Message, "[baseline]") {
			foundUnprefixed = true
		}
	}
	if !foundUnprefixed {
		t.Errorf("SpecStrict should not apply baseline; got %v", issues)
	}
	// Gate() includes it (ERROR severity, not filtered)
	gate := Gate(dir) // uses Spec which applies baseline → still filtered
	// Compare: SpecStrict's no-user-stories is an ERROR
	gateStrict := gateFromStrict(dir)
	if len(gateStrict) == 0 {
		t.Error("Gate via SpecStrict should block on no-user-stories ERROR")
	}
	_ = gate
}

func TestSpec_BaselineParseError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "requirements.md"), "# F\n\n## Acceptance Criteria\nWHEN foo THE SYSTEM SHALL bar.\n")
	// Malformed JSON
	writeFile(t, filepath.Join(dir, BaselineFileName), `{not valid`)
	issues := Spec(dir)
	var found bool
	for _, i := range issues {
		if i.Code == "baseline-parse-error" {
			found = true
			if i.Severity != SeverityError {
				t.Errorf("baseline-parse-error should be ERROR; got %q", i.Severity)
			}
		}
	}
	if !found {
		t.Errorf("expected baseline-parse-error; got %v", issues)
	}
}

// gateFromStrict is a helper that runs Gate() equivalent but via
// SpecStrict — used by tests that need to verify the "baseline not
// applied" path through Gate().
func gateFromStrict(specDir string) []Issue {
	var gate []Issue
	for _, i := range SpecStrict(specDir) {
		if i.Severity != SeverityError {
			continue
		}
		if isMissingCode(i.Code) {
			continue
		}
		gate = append(gate, i)
	}
	return gate
}

// --- helpers ---

func assertHasCode(t *testing.T, issues []Issue, code, wantSeverity string) {
	t.Helper()
	for _, i := range issues {
		if i.Code == code {
			if i.Severity != wantSeverity {
				t.Errorf("issue %q: severity %q, want %q", code, i.Severity, wantSeverity)
			}
			return
		}
	}
	t.Errorf("expected code %q in issues; got %d issues", code, len(issues))
}

func assertNoCode(t *testing.T, issues []Issue, code string) {
	t.Helper()
	for _, i := range issues {
		if i.Code == code {
			t.Errorf("did not expect code %q; got %s", code, i)
		}
	}
}