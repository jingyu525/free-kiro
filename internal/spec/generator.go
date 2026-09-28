// Package spec owns the spec lifecycle engine: creating specs, generating
// planning documents from templates, advancing phases, capturing drift
// baselines, and emitting advisory consistency reports.
package spec

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/models"
)

//go:embed templates/*.md.tmpl
var templateFS embed.FS

// templateData is the context handed to each planning-doc template. Only
// Name is required; spec_type is exposed so bugfix.md can render its own
// header without branching in the template.
type templateData struct {
	Name     string
	SpecType string
}

// renderTemplate executes the named template with data. Returns an error
// (wrapped) on parse or execution failure.
func renderTemplate(name string, data templateData) (string, error) {
	src, err := templateFS.ReadFile("templates/" + name)
	if err != nil {
		return "", ferrors.Wrap("spec.template", err, "read "+name)
	}
	tmpl, err := template.New(name).Parse(string(src))
	if err != nil {
		return "", ferrors.Wrap("spec.template", err, "parse "+name)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", ferrors.Wrap("spec.template", err, "execute "+name)
	}
	return buf.String(), nil
}

// writeIfMissing writes content to path only when the file does not
// already exist. Overwriting an existing planning document is almost
// always a mistake (the author has already invested work there) so the
// generator refuses silently rather than clobbering.
//
// Use --force on the CLI to override; see Engine.Generate with force=true.
func writeIfMissing(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// generateOne writes the document for a single (specType, phase) pair to
// the given spec directory. Returns the path that was written (or the
// existing path if the file already existed).
func generateOne(specDir, name, specType string, phase models.Phase, force bool) (string, error) {
	doc := models.PhaseDocFor(specType, phase)
	path := filepath.Join(specDir, doc)
	if !force {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	body, err := renderTemplate(doc+".tmpl", templateData{Name: name, SpecType: specType})
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return "", ferrors.Wrap("spec.generate", err, "write "+path)
	}
	return path, nil
}

// ensureMeta writes a default .meta.json into specDir if one does not
// already exist. Used by Generate when the caller has not yet created
// the spec — most callers (SpecEngine.NewSpec) write it explicitly.
func ensureMeta(specDir string, meta *models.SpecMeta) error {
	path := filepath.Join(specDir, models.MetaFileName)
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return meta.Save(specDir)
}

// listSpecDocs is a tiny helper for callers that want to print which docs
// exist. Returns full paths.
func listSpecDocs(specDir string) []string {
	docs := []string{"requirements.md", "design.md", "tasks.md", "bugfix.md"}
	var out []string
	for _, d := range docs {
		p := filepath.Join(specDir, d)
		if _, err := os.Stat(p); err == nil {
			out = append(out, p)
		}
	}
	return out
}

// firstDocExists reports whether the spec-type-appropriate first planning
// document is present. Used by NextAction to gate the "next" suggestion.
func firstDocExists(specDir, specType string) bool {
	doc := models.FirstPlanningDoc(specType)
	_, err := os.Stat(filepath.Join(specDir, doc))
	return err == nil
}

// fmtMissing prints a "missing …" hint with the spec name interpolated.
func fmtMissing(specName, what string) string {
	return fmt.Sprintf("%s: missing %s — run `free-kiro spec generate %s --phase %s`",
		specName, what, specName, phaseForDoc(what))
}

// phaseForDoc maps a document filename back to its phase (only the
// well-known planning docs). Returns "" for unknown inputs.
func phaseForDoc(doc string) string {
	switch doc {
	case "requirements.md", "bugfix.md":
		return string(models.PhaseRequirements)
	case "design.md":
		return string(models.PhaseDesign)
	case "tasks.md":
		return string(models.PhaseTasks)
	}
	return ""
}