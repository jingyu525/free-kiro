package models

// SteeringDoc is a persistent project-context document (product / structure
// / tech / AGENTS) loaded from .kiro/steering/*.md or ~/.kiro/steering/*.md.
//
// Two scopes are merged with workspace overriding global by name. Each doc
// carries a Mode (always / auto / manual / filematch) and an optional
// Description used to match prompts (for auto) and FilePatterns used to
// match the path of the file being worked on (for filematch).
type SteeringDoc struct {
	Name         string
	Mode         string   // "always" | "auto" | "manual" | "filematch"
	Description  string
	Content      string
	FilePatterns []string // fileMatch patterns (populated only when Mode == "filematch")
	Scope        string   // "workspace" (project) or "global" (user ~/.kiro)
}