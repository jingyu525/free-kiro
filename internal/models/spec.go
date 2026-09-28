package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

// Workflow constants. A spec picks one at creation; it only changes the
// preferred order of the three planning docs, not the phase machine.
const (
	WorkflowRequirementsFirst = "requirements-first"
	WorkflowDesignFirst       = "design-first"
)

// SpecType constants. A feature spec uses requirements.md; a bug-fix spec
// uses bugfix.md for its first planning document.
const (
	SpecTypeFeature = "feature"
	SpecTypeBugfix  = "bugfix"
)

// Generator identifies how a spec document was produced. The clone only
// ships a template generator; model-driven generators can plug in later
// without changing the file format.
const (
	GeneratorTemplate = "template"
)

// SpecMeta is the persistent metadata for one spec, stored as
// .kiro/specs/<name>/.meta.json. Baseline captures an immutable snapshot
// of the spec at approval time (acceptance-criteria count, task count) so
// drift detection has a reference.
type SpecMeta struct {
	Name       string         `json:"name"`
	Phase      Phase          `json:"phase"`
	Workflow   string         `json:"workflow"`
	SpecType   string         `json:"spec_type"`
	Quick      bool           `json:"quick"`
	Approved   bool           `json:"approved"`
	Generator  string         `json:"generator"`
	Prompt     string         `json:"prompt"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
	Baseline   map[string]int `json:"baseline"`
}

// MetaFileName is the filename inside SpecDir() that stores SpecMeta.
const MetaFileName = ".meta.json"

// LoadSpecMeta reads .meta.json from the given spec directory. Returns a
// descriptive KiroError on any failure.
func LoadSpecMeta(specDir string) (*SpecMeta, error) {
	path := filepath.Join(specDir, MetaFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, ferrors.Wrap("spec.meta", err, "read "+path)
	}
	var m SpecMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, ferrors.Wrap("spec.meta", err, "parse "+path)
	}
	m.applyDefaults()
	return &m, nil
}

// Save writes the metadata to .meta.json with 2-space indent.
func (m *SpecMeta) Save(specDir string) error {
	m.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return ferrors.Wrap("spec.meta", err, "marshal")
	}
	data = append(data, '\n')
	path := filepath.Join(specDir, MetaFileName)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return ferrors.Wrap("spec.meta", err, "write "+path)
	}
	return nil
}

// NewSpecMeta constructs a SpecMeta for a fresh spec.
func NewSpecMeta(name, prompt, workflow, specType string, quick bool) *SpecMeta {
	if workflow == "" {
		workflow = WorkflowRequirementsFirst
	}
	if specType == "" {
		specType = SpecTypeFeature
	}
	now := time.Now().UTC().Format(time.RFC3339)
	return &SpecMeta{
		Name:      name,
		Phase:     PhaseDraft,
		Workflow:  workflow,
		SpecType:  specType,
		Quick:     quick,
		Approved:  false,
		Generator: GeneratorTemplate,
		Prompt:    prompt,
		CreatedAt: now,
		UpdatedAt: now,
		Baseline:  map[string]int{},
	}
}

// PhaseEnum returns the spec's phase as a typed Phase value. Equivalent
// to Phase(m.Phase) but reads better at call sites that need to switch
// on the phase.
func (m *SpecMeta) PhaseEnum() Phase { return Phase(m.Phase) }

func (m *SpecMeta) applyDefaults() {
	if m.Phase == "" {
		m.Phase = PhaseDraft
	}
	if m.Workflow == "" {
		m.Workflow = WorkflowRequirementsFirst
	}
	if m.SpecType == "" {
		m.SpecType = SpecTypeFeature
	}
	if m.Generator == "" {
		m.Generator = GeneratorTemplate
	}
	if m.Baseline == nil {
		m.Baseline = map[string]int{}
	}
}