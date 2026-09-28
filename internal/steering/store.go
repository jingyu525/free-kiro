package steering

import (
	"os"
	"path/filepath"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"

	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/workspace"
)

// GlobalDirName is the conventional name of the user's global steering
// directory inside ~/.kiro. Exposed so callers can relocate it in tests.
const GlobalDirName = "steering"

// Store loads steering docs from both scopes (global + workspace) and
// merges them with workspace overriding global by name. Use Assemble to
// project the loaded docs into a context block for a generation request.
type Store struct {
	ws        *workspace.Workspace
	globalDir string
}

// NewStore constructs a Store for the given workspace. Pass an empty
// globalDir to use the default (~/.kiro/steering); supply a path to
// relocate the global scope (mainly for tests).
func NewStore(ws *workspace.Workspace, globalDir string) *Store {
	if globalDir == "" {
		globalDir = filepath.Join(homeDir(), ".kiro", GlobalDirName)
	}
	return &Store{ws: ws, globalDir: globalDir}
}

// homeDir returns $HOME, falling back to "." when unset (very rare on
// macOS/Linux but defensive).
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return "."
}

// LoadAll returns every steering doc from both scopes, workspace
// overriding global by name. AGENTS.md is loaded as a doc named "AGENTS"
// from both <root>/AGENTS.md and ~/.kiro/AGENTS.md.
//
// Malformed files (invalid frontmatter / mode) raise a SteeringError.
func (s *Store) LoadAll() []models.SteeringDoc {
	merged := map[string]models.SteeringDoc{}

	// Global scope first (lower priority).
	if entries, err := os.ReadDir(s.globalDir); err == nil {
		for _, ent := range entries {
			if ent.IsDir() || !hasMDExt(ent.Name()) {
				continue
			}
			doc, err := loadDoc(filepath.Join(s.globalDir, ent.Name()), "global")
			if err != nil {
				// Skip malformed files rather than abort the whole load;
				// this matches kiro's lenient behaviour.
				continue
			}
			merged[doc.Name] = doc
		}
	}
	if data, err := os.ReadFile(filepath.Join(s.globalDir, "..", "AGENTS.md")); err == nil {
		if doc, err := parseDoc(data, "global", "AGENTS"); err == nil {
			merged[doc.Name] = doc
		}
	}

	// Workspace scope — overrides by name.
	if entries, err := os.ReadDir(s.ws.SteeringDir()); err == nil {
		for _, ent := range entries {
			if ent.IsDir() || !hasMDExt(ent.Name()) {
				continue
			}
			doc, err := loadDoc(filepath.Join(s.ws.SteeringDir(), ent.Name()), "workspace")
			if err != nil {
				continue
			}
			merged[doc.Name] = doc
		}
	}
	if data, err := os.ReadFile(filepath.Join(s.ws.Root(), "AGENTS.md")); err == nil {
		if doc, err := parseDoc(data, "workspace", "AGENTS"); err == nil {
			merged[doc.Name] = doc
		}
	}

	out := make([]models.SteeringDoc, 0, len(merged))
	for _, d := range merged {
		out = append(out, d)
	}
	return out
}

// Get returns one doc by name (across both scopes), or nil.
func (s *Store) Get(name string) *models.SteeringDoc {
	for _, d := range s.LoadAll() {
		if d.Name == name {
			return &d
		}
	}
	return nil
}

// loadDoc reads + parses one steering file. The doc's Name is derived
// from the filename (without `.md`); AGENTS.md uses a special-cased name.
func loadDoc(path, scope string) (models.SteeringDoc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return models.SteeringDoc{}, ferrors.Wrap("steering.load", err, path)
	}
	name := ""
	if filepath.Base(path) == "AGENTS.md" {
		name = "AGENTS"
	} else {
		name = trimExt(filepath.Base(path))
	}
	return parseDoc(data, scope, name)
}

// trimExt removes the trailing `.md` (or any `.xxx`) extension.
func trimExt(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '.' {
			return p[:i]
		}
	}
	return p
}

// parseDoc is the pure-data path: takes raw bytes and the doc name
// (already known from the filename).
func parseDoc(data []byte, scope, name string) (models.SteeringDoc, error) {
	meta, body := ParseFrontmatter(string(data))
	body = trimTrailingWS(body)
	raw := meta["inclusion"]
	if raw == "" {
		raw = meta["mode"]
	}
	if raw == "" {
		raw = "always"
	}
	if !ValidModes[raw] {
		return models.SteeringDoc{}, ferrors.New("steering.frontmatter",
			"invalid mode "+raw+" (expected always / auto / manual / filematch)")
	}
	doc := models.SteeringDoc{
		Name:        name,
		Mode:        raw,
		Description: meta["description"],
		Content:     body,
		Scope:       scope,
	}
	if raw == "filematch" {
		doc.FilePatterns = ParseFilePatterns(meta["fileMatchPattern"])
	}
	return doc, nil
}

// trimTrailingWS strips trailing newlines/spaces from a doc body.
func trimTrailingWS(s string) string {
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t' || s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

// deriveName produces a stable doc name from the body. We look for the
// first H1 (`# Title`); if missing, fall back to the first non-empty
// line. Used when the filename is non-meaningful (e.g. AGENTS.md).
func deriveName(body string) string {
	for _, line := range splitLinesForSteering(body) {
		if line == "" {
			continue
		}
		if len(line) > 2 && line[0] == '#' && line[1] == ' ' {
			return trim(line[2:])
		}
		break
	}
	return "untitled"
}

func hasMDExt(name string) bool {
	return filepath.Ext(name) == ".md"
}