package lint

import (
	"testing"
)

// --- AC-1 CheckEtcList ---

func TestCheckEtcList(t *testing.T) {
	t.Run("hit: etc in AC", func(t *testing.T) {
		issues := CheckEtcList("WHEN user configures foo THE SYSTEM SHALL validate etc inputs.")
		assertOneIssue(t, issues, "ears-etc-list", SeverityError)
	})
	t.Run("hit: and/or in AC", func(t *testing.T) {
		issues := CheckEtcList("WHEN failure happens THE SYSTEM SHALL retry and/or log.")
		assertOneIssue(t, issues, "ears-etc-list", SeverityError)
	})
	t.Run("miss: clean AC", func(t *testing.T) {
		issues := CheckEtcList("WHEN user clicks login THE SYSTEM SHALL redirect within 200 ms.")
		if len(issues) != 0 {
			t.Errorf("expected no issues; got %v", issues)
		}
	})
	t.Run("miss: etc only in prose (not in EARS line)", func(t *testing.T) {
		// `etc` appears in a paragraph that is not an EARS AC; the rule
		// only scans lines that match EARSRe.
		doc := "Some prose mentioning etc. and other things.\n\n" +
			"WHEN foo THE SYSTEM SHALL bar."
		issues := CheckEtcList(doc)
		if len(issues) != 0 {
			t.Errorf("expected no issues; got %v", issues)
		}
	})
}

// --- AC-2 CheckSingleSHALLLine ---

func TestCheckSingleSHALLLine(t *testing.T) {
	t.Run("hit: two SHALLs on one line", func(t *testing.T) {
		line := "WHEN login THE SYSTEM SHALL redirect AND WHEN token expires THE SYSTEM SHALL refresh."
		issues := CheckSingleSHALLLine(line)
		assertOneIssue(t, issues, "ears-multi-shall-line", SeverityWarning)
	})
	t.Run("miss: one SHALL", func(t *testing.T) {
		issues := CheckSingleSHALLLine("WHEN login THE SYSTEM SHALL redirect.")
		if len(issues) != 0 {
			t.Errorf("expected no issues; got %v", issues)
		}
	})
	t.Run("miss: SHALLOW continuation across lines", func(t *testing.T) {
		// Two ACs on separate lines, each with one SHALL, should not
		// trigger the per-line rule.
		doc := "WHEN login THE SYSTEM SHALL redirect.\n" +
			"WHEN token expires THE SYSTEM SHALL refresh."
		issues := CheckSingleSHALLLine(doc)
		if len(issues) != 0 {
			t.Errorf("expected no issues; got %v", issues)
		}
	})
}

// --- AC-3 CheckFewAC ---

func TestCheckFewAC(t *testing.T) {
	t.Run("hit: 0 ACs against min 3", func(t *testing.T) {
		issues := CheckFewAC("no EARS here", 3)
		assertOneIssue(t, issues, "ears-few-ac", SeverityWarning)
	})
	t.Run("hit: 2 ACs against min 3", func(t *testing.T) {
		doc := "WHEN foo THE SYSTEM SHALL bar.\nWHEN baz THE SYSTEM SHALL qux."
		issues := CheckFewAC(doc, 3)
		assertOneIssue(t, issues, "ears-few-ac", SeverityWarning)
	})
	t.Run("miss: 3 ACs against min 3", func(t *testing.T) {
		doc := "WHEN foo THE SYSTEM SHALL bar.\n" +
			"WHEN baz THE SYSTEM SHALL qux.\n" +
			"THE SYSTEM SHALL persist preferences."
		issues := CheckFewAC(doc, 3)
		if len(issues) != 0 {
			t.Errorf("expected no issues; got %v", issues)
		}
	})
}

// --- AC-4 CheckTemplateDiversity ---

func TestCheckTemplateDiversity(t *testing.T) {
	t.Run("hit: only WHEN", func(t *testing.T) {
		doc := "WHEN foo THE SYSTEM SHALL bar.\nWHEN baz THE SYSTEM SHALL qux."
		issues := CheckTemplateDiversity(doc)
		assertOneIssue(t, issues, "ears-low-template-diversity", SeverityWarning)
	})
	t.Run("miss: WHEN + WHILE", func(t *testing.T) {
		doc := "WHEN foo THE SYSTEM SHALL bar.\nWHILE active THE SYSTEM SHALL baz."
		issues := CheckTemplateDiversity(doc)
		if len(issues) != 0 {
			t.Errorf("expected no issues; got %v", issues)
		}
	})
	t.Run("miss: WHEN + ubiquitous", func(t *testing.T) {
		doc := "WHEN foo THE SYSTEM SHALL bar.\nTHE SYSTEM SHALL persist preferences."
		issues := CheckTemplateDiversity(doc)
		if len(issues) != 0 {
			t.Errorf("expected no issues; got %v", issues)
		}
	})
	t.Run("miss: IF-THEN counts", func(t *testing.T) {
		doc := "WHEN foo THE SYSTEM SHALL bar.\nIF rate exceeded THEN THE SYSTEM SHALL return 429."
		issues := CheckTemplateDiversity(doc)
		if len(issues) != 0 {
			t.Errorf("IF-THEN should count via UsesIFTHEN; got %v", issues)
		}
	})
}

// --- AC-5 CheckACMissingID ---

func TestCheckACMissingID(t *testing.T) {
	t.Run("hit: AC line without [AC-N] prefix", func(t *testing.T) {
		doc := "WHEN foo THE SYSTEM SHALL bar."
		issues := CheckACMissingID(doc)
		assertOneIssue(t, issues, "ears-ac-missing-id", SeverityWarning)
	})
	t.Run("miss: AC line with [AC-1] prefix", func(t *testing.T) {
		doc := "- [AC-1] WHEN foo THE SYSTEM SHALL bar within 200 ms."
		issues := CheckACMissingID(doc)
		if len(issues) != 0 {
			t.Errorf("expected no issues; got %v", issues)
		}
	})
	t.Run("hit: mixed — some have IDs, some don't", func(t *testing.T) {
		// AC-1 is ID'd and multi-line; a non-template prose line breaks
		// the continuation, so the next AC starts fresh without an ID.
		doc := "- [AC-1] WHEN foo THE SYSTEM SHALL bar.\n" +
			"  prose continuation that does NOT start with a template keyword.\n" +
			"- WHEN baz THE SYSTEM SHALL qux within 100 ms."
		issues := CheckACMissingID(doc)
		if len(issues) != 1 {
			t.Errorf("expected exactly 1 issue for the third line; got %v", issues)
		}
	})
}

// --- acIDPrefixRe + extractEARSLines regression (bullet-prefixed AC marker) ---
//
// Before this fix, `- [AC-N] WHEN ... THE SYSTEM SHALL ...` (bullet-
// prefixed canonical form) was not recognized as an AC marker line by
// acIDPrefixRe, so the following `  THE SYSTEM SHALL ...` continuation
// was misclassified as a fresh legacy AC — producing a spurious
// `ears-ac-missing-id` warning on every multi-line bullet-prefixed AC
// (e.g. performance-benchmarks:39, :53, :63).

func TestExtractEARSLinesBulletPrefix(t *testing.T) {
	t.Run("bullet AC marker with continuation is one AC", func(t *testing.T) {
		doc := "- [AC-1] WHEN foo THE SYSTEM SHALL do X\n" +
			"  THE SYSTEM SHALL continue with Y\n" +
			"- [AC-2] WHEN bar THE SYSTEM SHALL do Z.\n"
		acs := extractEARSLines(doc)
		if len(acs) != 2 {
			t.Fatalf("expected 2 ACs (continuation must not split AC-1); got %d: %+v", len(acs), acs)
		}
		if acs[0].LineNum != 1 {
			t.Errorf("AC-1 line should be 1; got %d", acs[0].LineNum)
		}
		if acs[1].LineNum != 3 {
			t.Errorf("AC-2 line should be 3; got %d", acs[1].LineNum)
		}
	})
	t.Run("no bullet AC marker still recognized", func(t *testing.T) {
		doc := "[AC-1] WHEN foo THE SYSTEM SHALL do X.\n" +
			"[AC-2] WHEN bar THE SYSTEM SHALL do Z.\n"
		acs := extractEARSLines(doc)
		if len(acs) != 2 {
			t.Fatalf("expected 2 ACs; got %d: %+v", len(acs), acs)
		}
	})
}

func TestACIDPrefixReAcceptsBullet(t *testing.T) {
	for _, line := range []string{
		"- [AC-1] WHEN foo THE SYSTEM SHALL bar.",
		"[AC-1] WHEN foo THE SYSTEM SHALL bar.",
		"  - [AC-2] WHEN foo THE SYSTEM SHALL bar.",
		"  [AC-2] WHEN foo THE SYSTEM SHALL bar.",
	} {
		if !acIDPrefixRe.MatchString(line) {
			t.Errorf("acIDPrefixRe must accept %q as an AC marker line", line)
		}
	}
}

// --- AC-6 CheckMeasurableResponse ---

func TestCheckMeasurableResponse(t *testing.T) {
	t.Run("hit: response has no measurable", func(t *testing.T) {
		issues := CheckMeasurableResponse("WHEN user clicks login THE SYSTEM SHALL respond fast.")
		assertOneIssue(t, issues, "ears-response-immeasurable", SeverityWarning)
	})
	t.Run("miss: response has 200 ms", func(t *testing.T) {
		issues := CheckMeasurableResponse("WHEN user clicks login THE SYSTEM SHALL respond within 200 ms.")
		if len(issues) != 0 {
			t.Errorf("expected no issues; got %v", issues)
		}
	})
	t.Run("miss: response has status code 429", func(t *testing.T) {
		issues := CheckMeasurableResponse("WHEN rate limit exceeded THE SYSTEM SHALL return 429.")
		if len(issues) != 0 {
			t.Errorf("expected no issues; got %v", issues)
		}
	})
	t.Run("miss: response has percentage", func(t *testing.T) {
		issues := CheckMeasurableResponse("WHEN cpu high THE SYSTEM SHALL throttle at 90% utilization.")
		if len(issues) != 0 {
			t.Errorf("expected no issues; got %v", issues)
		}
	})
	t.Run("miss: response has 'at most'", func(t *testing.T) {
		issues := CheckMeasurableResponse("WHEN burst THE SYSTEM SHALL queue at most 100 events.")
		if len(issues) != 0 {
			t.Errorf("expected no issues; got %v", issues)
		}
	})
}

// --- AC-7 CheckTriggerObservable ---

func TestCheckTriggerObservable(t *testing.T) {
	t.Run("hit: WHEN with busy trigger", func(t *testing.T) {
		issues := CheckTriggerObservable("WHEN system is busy THE SYSTEM SHALL throttle.")
		assertOneIssue(t, issues, "ears-trigger-unobservable", SeverityWarning)
	})
	t.Run("hit: WHILE with recently", func(t *testing.T) {
		issues := CheckTriggerObservable("WHILE user recently logged in THE SYSTEM SHALL poll.")
		assertOneIssue(t, issues, "ears-trigger-unobservable", SeverityWarning)
	})
	t.Run("miss: WHEN with concrete rate", func(t *testing.T) {
		issues := CheckTriggerObservable("WHEN requests > 100/s THE SYSTEM SHALL throttle.")
		if len(issues) != 0 {
			t.Errorf("expected no issues; got %v", issues)
		}
	})
	t.Run("miss: WHERE line not checked", func(t *testing.T) {
		// trigger observability only looks at WHEN/WHILE lines
		issues := CheckTriggerObservable("WHERE fast mode is enabled THE SYSTEM SHALL enable turbo.")
		if len(issues) != 0 {
			t.Errorf("WHERE lines should not be checked; got %v", issues)
		}
	})
}

// --- AC-8 CheckKeywordMisuse ---

func TestCheckKeywordMisuse(t *testing.T) {
	t.Run("hit: WHILE with logged in", func(t *testing.T) {
		issues := CheckKeywordMisuse("WHILE user logged in THE SYSTEM SHALL poll.")
		assertOneIssue(t, issues, "ears-keyword-misuse-while-as-when", SeverityWarning)
	})
	t.Run("hit: WHILE with failed", func(t *testing.T) {
		issues := CheckKeywordMisuse("WHILE request failed THE SYSTEM SHALL retry.")
		assertOneIssue(t, issues, "ears-keyword-misuse-while-as-when", SeverityWarning)
	})
	t.Run("miss: WHILE with continuous state", func(t *testing.T) {
		// "session is active" — proper state phrase, no past-tense verb
		issues := CheckKeywordMisuse("WHILE session is active THE SYSTEM SHALL refresh tokens.")
		if len(issues) != 0 {
			t.Errorf("expected no issues; got %v", issues)
		}
	})
	t.Run("miss: WHEN with past-tense verb (not WHILE)", func(t *testing.T) {
		// WHEN lines are not checked — past-tense is correct for events
		issues := CheckKeywordMisuse("WHEN user logged in THE SYSTEM SHALL redirect.")
		if len(issues) != 0 {
			t.Errorf("WHEN lines should not be checked; got %v", issues)
		}
	})
}

// --- AC-9 CheckPassiveResponse ---

func TestCheckPassiveResponse(t *testing.T) {
	t.Run("hit: SHALL be fast", func(t *testing.T) {
		issues := CheckPassiveResponse("THE SYSTEM SHALL be fast.")
		assertOneIssue(t, issues, "ears-passive-response", SeverityWarning)
	})
	t.Run("hit: SHALL is reliable", func(t *testing.T) {
		issues := CheckPassiveResponse("THE SYSTEM SHALL is reliable.")
		assertOneIssue(t, issues, "ears-passive-response", SeverityWarning)
	})
	t.Run("miss: SHALL persist preferences", func(t *testing.T) {
		issues := CheckPassiveResponse("THE SYSTEM SHALL persist preferences.")
		if len(issues) != 0 {
			t.Errorf("expected no issues; got %v", issues)
		}
	})
	t.Run("miss: SHALL return 429", func(t *testing.T) {
		issues := CheckPassiveResponse("WHEN rate exceeded THE SYSTEM SHALL return 429.")
		if len(issues) != 0 {
			t.Errorf("expected no issues; got %v", issues)
		}
	})
}

// --- helpers ---

func assertOneIssue(t *testing.T, issues []Issue, code, wantSeverity string) {
	t.Helper()
	if len(issues) != 1 {
		t.Fatalf("expected exactly 1 issue; got %d: %v", len(issues), issues)
	}
	if issues[0].Code != code {
		t.Errorf("issue code: got %q want %q", issues[0].Code, code)
	}
	if issues[0].Severity != wantSeverity {
		t.Errorf("issue severity: got %q want %q", issues[0].Severity, wantSeverity)
	}
}
