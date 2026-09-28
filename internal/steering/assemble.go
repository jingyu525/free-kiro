package steering

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jingyu525/free-kiro/internal/models"
)

// Assemble builds the steering context block for a generation request.
//
// Inclusion rules (per doc mode):
//
//   - always    → always included
//   - auto      → included iff prompt keywords overlap the doc's description
//   - manual    → included only when the name is in `manual`
//   - filematch → included iff `currentFile` matches one of the doc's
//                 fileMatchPattern globs (or name is in `manual`)
//
// Returns a single concatenated string suitable for splicing into a
// generation prompt. Returns "" when nothing matches.
func (s *Store) Assemble(prompt string, manual []string, currentFile string) string {
	docs := s.LoadAll()
	promptKw := keywords(prompt)
	manualSet := toSet(manual)
	cf := filepath.ToSlash(currentFile)

	var blocks []string
	for _, doc := range docs {
		if !shouldInclude(doc, promptKw, manualSet, cf) {
			continue
		}
		if doc.Content == "" {
			continue
		}
		header := "[steering:" + doc.Name + " | mode=" + doc.Mode + " | scope=" + doc.Scope + "]"
		blocks = append(blocks, header+"\n"+doc.Content)
	}
	return strings.Join(blocks, "\n\n")
}

func shouldInclude(doc models.SteeringDoc, promptKw map[string]bool, manualSet map[string]bool, currentFile string) bool {
	switch doc.Mode {
	case "always":
		return true
	case "auto":
		descKw := keywords(doc.Description)
		// Include when any keyword overlaps.
		for k := range descKw {
			if promptKw[k] {
				return true
			}
		}
		return false
	case "filematch":
		if currentFile == "" {
			return manualSet[doc.Name]
		}
		for _, pat := range doc.FilePatterns {
			if GlobMatch(pat, currentFile) {
				return true
			}
		}
		return manualSet[doc.Name]
	default: // manual
		return manualSet[doc.Name]
	}
}

// keywords extracts normalised keywords (lowercased, alphanumeric only,
// length > 2) from text. Used for the prompt↔description overlap check
// on `auto` mode docs.
func keywords(text string) map[string]bool {
	out := map[string]bool{}
	for _, w := range wordRe.FindAllString(text, -1) {
		w = strings.ToLower(w)
		if len(w) > 2 {
			out[w] = true
		}
	}
	return out
}

var wordRe = regexp.MustCompile(`[A-Za-z][A-Za-z0-9_]+`)

func toSet(xs []string) map[string]bool {
	out := map[string]bool{}
	for _, x := range xs {
		out[x] = true
	}
	return out
}