// Package spec owns the spec lifecycle engine: creating specs, generating
// planning documents from templates, advancing phases, capturing drift
// baselines, and emitting advisory consistency reports.
package spec

import (
	"bytes"
	"embed"
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

// generateOne writes the document for a single (specType, phase) pair to

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

// generateOne is the sole public-facing planner; everything above (renderTemplate)
// is private scaffolding for it.
