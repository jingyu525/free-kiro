package lint

import (
	"regexp"
	"testing"
)

func TestEARSRe_AllTemplates(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{"event-driven WHEN", "WHEN user logs in THE SYSTEM SHALL redirect to dashboard", true},
		{"state-driven WHILE", "WHILE session is active THE SYSTEM SHALL refresh token", true},
		{"optional WHERE", "WHERE feature flag X is on THE SYSTEM SHALL show banner", true},
		{"exception UNLESS", "UNLESS user is admin THE SYSTEM SHALL hide button", true},
		{"complex IF-THEN", "IF rate limit exceeded THEN THE SYSTEM SHALL return 429", true},
		{"ubiquitous THE SYSTEM SHALL", "THE SYSTEM SHALL persist user preferences", true},
		{"no-ears plain prose", "users should be able to login", false},
		{"lowercase shall", "when foo the system shall bar", true}, // IGNORECASE
		{"mixed case", "When foo THE SYSTEM Shall do X", true},
		{"SHALL alone (ubiquitous)", "the system shall require strong passwords", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := EARSRe.MatchString(c.input)
			if got != c.want {
				t.Errorf("EARSRe.MatchString(%q) = %v, want %v", c.input, got, c.want)
			}
		})
	}
}

func TestLintRequirements_Empty(t *testing.T) {
	issues := Requirements("")
	if len(issues) != 1 || issues[0].Code != "empty-requirements" {
		t.Fatalf("expected empty-requirements ERROR; got %v", issues)
	}
	if issues[0].Severity != SeverityError {
		t.Errorf("expected ERROR severity; got %q", issues[0].Severity)
	}
}

func TestLintRequirements_NoEars(t *testing.T) {
	issues := Requirements("# Feature\n\nSome prose without acceptance criteria.\n")
	if len(issues) == 0 {
		t.Fatal("expected at least one issue")
	}
	var foundEars, foundUserStory bool
	for _, i := range issues {
		if i.Code == "no-ears" {
			foundEars = true
			if i.Severity != SeverityError {
				t.Errorf("no-ears should be ERROR; got %q", i.Severity)
			}
		}
		if i.Code == "no-user-stories" {
			foundUserStory = true
		}
	}
	if !foundEars {
		t.Error("expected no-ears ERROR")
	}
	if !foundUserStory {
		t.Error("expected no-user-stories WARNING")
	}
}

func TestLintRequirements_GoldenPath(t *testing.T) {
	doc := `# User Auth

## User Stories
As a user I want to log in so that I can access my dashboard.

## Acceptance Criteria
WHEN the user submits valid credentials THE SYSTEM SHALL redirect to /dashboard.
WHILE the user is authenticated THE SYSTEM SHALL keep the session active.
`
	issues := Requirements(doc)
	for _, i := range issues {
		if i.Severity == SeverityError {
			t.Errorf("golden path should have no ERRORs; got %s", i)
		}
	}
}

func TestLintBugfix_Empty(t *testing.T) {
	issues := Bugfix("")
	if len(issues) != 1 || issues[0].Code != "empty-bugfix" {
		t.Fatalf("expected empty-bugfix; got %v", issues)
	}
}

func TestLintBugfix_GoldenPath(t *testing.T) {
	doc := `# Login redirect bug

## Current Behavior
WHEN a user logs in with valid credentials the system redirects to /home instead of /dashboard.

## Expected Behavior
WHEN the user submits valid credentials THE SYSTEM SHALL redirect to /dashboard.

## Unchanged Behavior
WHEN the user submits invalid credentials THE SYSTEM SHALL CONTINUE TO show the error page.
`
	issues := Bugfix(doc)
	for _, i := range issues {
		if i.Severity == SeverityError {
			t.Errorf("golden path should have no ERRORs; got %s", i)
		}
	}
}

func TestLintBugfix_DefectUsesShall(t *testing.T) {
	doc := `## Current Behavior
WHEN a user logs in THE SYSTEM SHALL redirect to /home. (wrong)

## Expected Behavior
WHEN the user logs in THE SYSTEM SHALL redirect to /dashboard.

## Unchanged Behavior
WHEN the session expires THE SYSTEM SHALL CONTINUE TO log the user out.
`
	issues := Bugfix(doc)
	var found bool
	for _, i := range issues {
		if i.Code == "defect-uses-shall" {
			found = true
			if i.Severity != SeverityError {
				t.Errorf("defect-uses-shall should be ERROR; got %q", i.Severity)
			}
		}
	}
	if !found {
		t.Error("expected defect-uses-shall ERROR")
	}
}

func TestLintBugfix_MissingExpected(t *testing.T) {
	doc := `## Current Behavior
wrong thing happens.

## Unchanged Behavior
WHEN foo THE SYSTEM SHALL CONTINUE TO bar.
`
	issues := Bugfix(doc)
	var found bool
	for _, i := range issues {
		if i.Code == "missing-expected" {
			found = true
		}
	}
	if !found {
		t.Error("expected missing-expected ERROR")
	}
}

func TestLintBugfix_MissingUnchanged(t *testing.T) {
	doc := `## Current Behavior
broken.

## Expected Behavior
WHEN fixed THE SYSTEM SHALL work.
`
	issues := Bugfix(doc)
	var found bool
	for _, i := range issues {
		if i.Code == "missing-unchanged" {
			found = true
		}
	}
	if !found {
		t.Error("expected missing-unchanged ERROR")
	}
}

func TestLintBugfix_ExpectedNoShall(t *testing.T) {
	doc := `## Expected Behavior
This should fix the redirect. (no SHALL here)
## Unchanged Behavior
WHEN foo THE SYSTEM SHALL CONTINUE TO bar.
`
	issues := Bugfix(doc)
	var found bool
	for _, i := range issues {
		if i.Code == "no-ears-expected" {
			found = true
		}
	}
	if !found {
		t.Error("expected no-ears-expected ERROR")
	}
}

func TestLintBugfix_NoShallContinue(t *testing.T) {
	doc := `## Expected Behavior
WHEN fixed THE SYSTEM SHALL work.
## Unchanged Behavior
everything else stays the same.
`
	issues := Bugfix(doc)
	var found bool
	for _, i := range issues {
		if i.Code == "no-shall-continue" {
			found = true
			if i.Severity != SeverityWarning {
				t.Errorf("no-shall-continue should be WARNING; got %q", i.Severity)
			}
		}
	}
	if !found {
		t.Error("expected no-shall-continue WARNING")
	}
}

func TestEARSRe_PerTemplateRegex(t *testing.T) {
	cases := []struct {
		name  string
		re    *regexp.Regexp
		input string
		want  bool
	}{
		{"WHEN matches its own template", WHENRe, "WHEN user logs in THE SYSTEM SHALL redirect", true},
		{"WHEN does not match WHILE template", WHENRe, "WHILE session active THE SYSTEM SHALL refresh", false},
		{"WHILE matches its own template", WHILERe, "WHILE session active THE SYSTEM SHALL refresh", true},
		{"WHILE does not match WHEN template", WHILERe, "WHEN user logs in THE SYSTEM SHALL redirect", false},
		{"WHERE matches its own template", WHERERe, "WHERE flag X is on THE SYSTEM SHALL show banner", true},
		{"UNLESS matches its own template", UNLESSRe, "UNLESS user is admin THE SYSTEM SHALL hide button", true},
		{"case-insensitive WHEN", WHENRe, "when foo the system shall bar", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.re.MatchString(c.input)
			if got != c.want {
				t.Errorf("%s.MatchString(%q) = %v, want %v", c.re, c.input, got, c.want)
			}
		})
	}
}

func TestUsesIFTHEN(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{"IF THEN SHALL on one line", "IF rate limit exceeded THEN THE SYSTEM SHALL return 429", true},
		{"IF without THEN", "IF user is admin THE SYSTEM SHALL show", false},
		{"THEN without IF", "WHEN foo THEN THE SYSTEM SHALL bar", false},
		{"lowercase", "if foo then the system shall bar", true},
		{"case-insensitive IF", "If foo THEN the system shall bar", true},
		{"multi-word condition", "IF user submits form 3 times within 1 minute THEN rate limit THE SYSTEM SHALL return 429", true},
		{"plain prose", "users should be able to login", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := UsesIFTHEN(c.input); got != c.want {
				t.Errorf("UsesIFTHEN(%q) = %v, want %v", c.input, got, c.want)
			}
		})
	}
}

func TestSplitSections_StripsParentheses(t *testing.T) {
	doc := `# Title
intro
## Current Behavior (Defect)
some text
## Expected Behavior
more text
`
	sections := splitSections(doc)
	if _, ok := sections["current behavior"]; !ok {
		t.Errorf("expected 'current behavior' key; got %v", sections)
	}
	if _, ok := sections["expected behavior"]; !ok {
		t.Errorf("expected 'expected behavior' key; got %v", sections)
	}
	if v := sections["current behavior"]; v != "some text" {
		t.Errorf("current behavior body: got %q, want %q", v, "some text")
	}
}

func TestSplitSections_Empty(t *testing.T) {
	if got := splitSections(""); len(got) != 0 {
		t.Errorf("splitSections(\"\") should return empty map; got %v", got)
	}
}

func TestNormaliseHeading(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Current Behavior", "current behavior"},
		{"Current Behavior (Defect)", "current behavior"},
		{"  Foo Bar  ", "foo bar"},
		{"X", "x"},
	}
	for _, c := range cases {
		if got := normaliseHeading(c.in); got != c.want {
			t.Errorf("normaliseHeading(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
