package lint

// LintBugfix checks bugfix.md for the Current / Expected / Unchanged
// Behavior contract of a bug-fix spec.
//
// The defect (Current Behavior) is *incorrect* behavior and must NOT use
// "THE SYSTEM SHALL" — defects are wrong, not "should". The corrective
// acceptance criteria (Expected and Unchanged Behavior) MUST use "THE
// SYSTEM SHALL" (Unchanged uses the "… SHALL CONTINUE TO" regression-
// prevention form). This mirrors kiro-clone / Kiro's bug-fix contract:
//
//   - `## Expected Behavior` is mandatory and must contain a SHALL (ERROR
//     if missing).
//   - `## Unchanged Behavior` is mandatory; missing → ERROR; uses SHALL
//     CONTINUE TO → preferred (WARNING if missing).
//   - `## Current Behavior` is optional (WARNING if missing) — but when
//     present, it must not assert a SHALL (ERROR `defect-uses-shall`).
//
// Pure function over the document text.
func LintBugfix(text string) []LintIssue {
	var out []LintIssue
	if isEmpty(text) {
		out = append(out, LintIssue{
			Severity: SeverityError,
			Code:     "empty-bugfix",
			Message:  "bugfix.md is empty",
			Location: "bugfix.md",
		})
		return out
	}

	sections := splitSections(text)
	current := sections["current behavior"]
	expected := sections["expected behavior"]
	unchanged := sections["unchanged behavior"]

	// Expected Behavior is the corrective acceptance criterion.
	if expected == "" {
		out = append(out, LintIssue{
			Severity: SeverityError,
			Code:     "missing-expected",
			Message:  "no '## Expected Behavior' section — the correct behavior MUST use THE SYSTEM SHALL",
			Location: "bugfix.md",
		})
	} else if !SHALLRe.MatchString(expected) {
		out = append(out, LintIssue{
			Severity: SeverityError,
			Code:     "no-ears-expected",
			Message:  "Expected Behavior section has no 'THE SYSTEM SHALL' acceptance criterion",
			Location: "bugfix.md:Expected Behavior",
		})
	}

	// Unchanged Behavior is regression prevention.
	if unchanged == "" {
		out = append(out, LintIssue{
			Severity: SeverityError,
			Code:     "missing-unchanged",
			Message:  "no '## Unchanged Behavior' section — regression prevention MUST use THE SYSTEM SHALL CONTINUE TO",
			Location: "bugfix.md",
		})
	} else if !SHALLContinueRe.MatchString(unchanged) {
		out = append(out, LintIssue{
			Severity: SeverityWarning,
			Code:     "no-shall-continue",
			Message:  "Unchanged Behavior section does not use 'THE SYSTEM SHALL CONTINUE TO'",
			Location: "bugfix.md:Unchanged Behavior",
		})
	}

	// Current Behavior is the defect description — optional but, when
	// present, MUST NOT assert a SHALL (that would legitimise the bug).
	if current == "" {
		out = append(out, LintIssue{
			Severity: SeverityWarning,
			Code:     "missing-current",
			Message:  "no '## Current Behavior' section — describe the observed defect",
			Location: "bugfix.md",
		})
	} else if SHALLRe.MatchString(current) {
		out = append(out, LintIssue{
			Severity: SeverityError,
			Code:     "defect-uses-shall",
			Message:  "Current Behavior (Defect) uses 'THE SYSTEM SHALL' — the defect is incorrect behavior and must NOT use SHALL",
			Location: "bugfix.md:Current Behavior",
		})
	}

	return out
}