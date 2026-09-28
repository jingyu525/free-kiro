package lint

import "regexp"

// placeholderRe matches free-kiro's `<TODO:…>` and `<TODO …>` template
// placeholders. When a placeholder appears inside what would otherwise
// count as an EARS acceptance criterion, the criterion is treated as
// unfilled and reported as `placeholder-ac`.
var placeholderRe = regexp.MustCompile(`<TODO[:\s]`)

// LintRequirements checks requirements.md for the structural contract of a
// feature spec:
//
//   - The document must not be empty.
//   - It MUST contain at least one EARS acceptance criterion (any of the six
//     templates) and that criterion MUST NOT contain `<TODO…>` placeholders.
//     Missing → ERROR `no-ears`; placeholders → ERROR `placeholder-ac`
//     (both block advance/approve).
//   - It SHOULD contain a User Stories section. Missing → WARNING
//     `no-user-stories` (advisory — does not block).
//
// Pure function over the document text.
func LintRequirements(text string) []LintIssue {
	var out []LintIssue
	if isEmpty(text) {
		out = append(out, LintIssue{
			Severity: SeverityError,
			Code:     "empty-requirements",
			Message:  "requirements.md is empty",
			Location: "requirements.md",
		})
		return out
	}
	if !EARSRe.MatchString(text) {
		out = append(out, LintIssue{
			Severity: SeverityError,
			Code:     "no-ears",
			Message: "no EARS acceptance criteria found (use WHEN/WHILE/WHERE/UNLESS/" +
				"IF…THEN … THE SYSTEM SHALL …, or the ubiquitous " +
				"\"The system shall …\" form)",
			Location: "requirements.md",
		})
	} else if placeholderRe.MatchString(text) {
		// EARS regex matched only because of `<TODO:…>` placeholders inside
		// template-generated AC lines. Flag the placeholder as the real
		// problem so the author replaces it with real content.
		out = append(out, LintIssue{
			Severity: SeverityError,
			Code:     "placeholder-ac",
			Message:  "acceptance criteria still contain <TODO…> placeholders — fill them in with real content before approving",
			Location: "requirements.md",
		})
	}
	if !UserStoryRe.MatchString(text) {
		out = append(out, LintIssue{
			Severity: SeverityWarning,
			Code:     "no-user-stories",
			Message:  "no User Stories section found",
			Location: "requirements.md",
		})
	}
	return out
}

func isEmpty(s string) bool {
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}