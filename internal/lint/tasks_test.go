package lint

import "testing"

func TestLintTasks_Empty(t *testing.T) {
	issues := LintTasks("# Tasks\n\nNo tasks yet.\n")
	if len(issues) != 1 || issues[0].Code != "empty-tasks" {
		t.Fatalf("expected empty-tasks WARNING; got %v", issues)
	}
	if issues[0].Severity != SeverityWarning {
		t.Errorf("expected WARNING; got %q", issues[0].Severity)
	}
}

func TestLintTasks_SelfDep(t *testing.T) {
	doc := "- [ ] #1 Bad task [deps: #1]\n"
	issues := LintTasks(doc)
	var found bool
	for _, i := range issues {
		if i.Code == "self-dep" {
			found = true
			if i.Severity != SeverityError {
				t.Errorf("self-dep should be ERROR; got %q", i.Severity)
			}
		}
	}
	if !found {
		t.Errorf("expected self-dep ERROR; got %v", issues)
	}
}

func TestLintTasks_DanglingDep(t *testing.T) {
	doc := "- [ ] #1 Real task\n- [ ] #2 Bad task [deps: #99]\n"
	issues := LintTasks(doc)
	var found bool
	for _, i := range issues {
		if i.Code == "dangling-dep" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected dangling-dep ERROR; got %v", issues)
	}
}

func TestLintTasks_Cycle(t *testing.T) {
	doc := `- [ ] #1 a [deps: #2]
- [ ] #2 b [deps: #1]
`
	issues := LintTasks(doc)
	var found bool
	for _, i := range issues {
		if i.Code == "cycle" {
			found = true
			if i.Severity != SeverityError {
				t.Errorf("cycle should be ERROR; got %q", i.Severity)
			}
		}
	}
	if !found {
		t.Errorf("expected cycle ERROR; got %v", issues)
	}
}

func TestLintTasks_GoldenPath(t *testing.T) {
	doc := `- [ ] #1 Set up module layout
- [ ] #2 Implement core [deps: #1]
- [ ] #3 Implement persistence [deps: #1]
- [x] #4 Polish [deps: #2,#3]
`
	issues := LintTasks(doc)
	for _, i := range issues {
		if i.Severity == SeverityError {
			t.Errorf("golden path should have no ERRORs; got %s", i)
		}
	}
}

func TestLintSpec_MissingAll(t *testing.T) {
	dir := t.TempDir()
	issues := LintSpec(dir)
	var hasReq, hasDesign, hasTasks bool
	for _, i := range issues {
		switch i.Code {
		case "missing-requirements":
			hasReq = true
		case "missing-design":
			hasDesign = true
		case "missing-tasks":
			hasTasks = true
		}
	}
	if !hasReq {
		t.Error("expected missing-requirements ERROR for empty spec")
	}
	if !hasDesign {
		t.Error("expected missing-design WARNING for empty spec")
	}
	if !hasTasks {
		t.Error("expected missing-tasks WARNING for empty spec")
	}
}

func TestLintSpec_GoldenPath(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir+"/requirements.md", `# Feature

## User Stories
As a user I want it.

## Acceptance Criteria
WHEN foo THE SYSTEM SHALL bar.
`)
	writeFile(t, dir+"/design.md", "# Design\n\nArchitecture notes.\n")
	writeFile(t, dir+"/tasks.md", "- [ ] #1 First task\n- [ ] #2 Second task [deps: #1]\n")
	// No .meta.json → defaults to feature.
	issues := LintSpec(dir)
	for _, i := range issues {
		if i.Severity == SeverityError {
			t.Errorf("golden spec should have no ERRORs; got %s", i)
		}
	}
}

func TestLintSpec_BugfixSpec(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir+"/.meta.json", `{"name":"x","phase":"design","spec_type":"bugfix"}`)
	writeFile(t, dir+"/bugfix.md", `## Current Behavior
broken thing happens.

## Expected Behavior
WHEN fixed THE SYSTEM SHALL work.

## Unchanged Behavior
WHEN foo THE SYSTEM SHALL CONTINUE TO bar.
`)
	writeFile(t, dir+"/design.md", "# Design\n\nRoot cause: line 42.\n")
	issues := LintSpec(dir)
	for _, i := range issues {
		if i.Severity == SeverityError {
			t.Errorf("golden bugfix should have no ERRORs; got %s", i)
		}
	}
}

func TestLintGate_ExcludesMissing(t *testing.T) {
	dir := t.TempDir()
	gate := LintGate(dir)
	// Empty spec → missing-* findings are EXCLUDED from the gate.
	for _, i := range gate {
		if i.Code == "missing-requirements" || i.Code == "missing-tasks" || i.Code == "missing-design" {
			t.Errorf("missing-* should not be in gate; got %s", i)
		}
	}
}

func TestLintGate_IncludesRealErrors(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir+"/requirements.md", `# Feature

Some prose, no EARS, no user stories.
`)
	gate := LintGate(dir)
	var foundNoEars bool
	for _, i := range gate {
		if i.Code == "no-ears" {
			foundNoEars = true
		}
	}
	if !foundNoEars {
		t.Error("no-ears should be in gate when requirements lacks EARS")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := osWriteFile(path, content); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}