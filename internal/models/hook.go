package models

// Hook is an event-driven automation loaded from .kiro/hooks/*.json.
//
// Accepts both free-kiro's flat shape and Kiro's official
// {version:"v1", hooks:[…]} envelope (see internal/hooks/envelope): name
// / trigger map to id / event, a Kiro matcher (regex) is stored with
// IsRegex=true, and action.type:"command" is aliased to "shell".
type Hook struct {
	ID          string
	Event       string // file.save | file.create | file.delete | prompt.submit | task.run | manual + Kiro PascalCase triggers
	Glob        string // filter pattern; empty matches everything. regex when IsRegex is true.
	ActionType  string // "shell" (alias "command") | "agent"
	Action      string // shell command or agent prompt
	Description string
	IsRegex     bool   // true when the filter came from a Kiro matcher
	Timeout     *int   // command timeout seconds; nil = engine default (30s), 0 = disabled
	Enabled     bool   // false skips the hook without deleting it
}