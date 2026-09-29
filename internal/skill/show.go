package skill

import (
	"os"
	"path/filepath"
)

// InstalledState describes one installed app's bundle state. Used by the
// `free-kiro skill show` command.
type InstalledState struct {
	App          App
	SkillsDir    string
	Installed    bool
	Version      string
	FreeKiroMin  string
	Experimental bool
}

// ShowInstalled inspects each app's skills/<subdir> directory and returns
// one InstalledState per app. apps is the list to inspect (e.g. from
// ParseApp("all") → AllApps()).
func ShowInstalled(home, subdir string) []InstalledState {
	if home == "" {
		h, err := HomeDir()
		if err != nil {
			h = ""
		}
		home = h
	}
	if subdir == "" {
		subdir = "free-kiro"
	}
	out := make([]InstalledState, 0, 4)
	for _, app := range AllApps() {
		dir := SkillsDir(app, home, subdir)
		s := InstalledState{App: app, SkillsDir: dir}
		if _, err := os.Stat(filepath.Join(dir, "skill.json")); err == nil {
			s.Installed = true
			if m, err := LoadManifest(dir); err == nil {
				s.Version = m.Version
				s.FreeKiroMin = m.FreeKiroMinVersion
			}
		}
		// Flag experimental status from the manifest template.
		if app == AppCodeBuddy {
			s.Experimental = true
		}
		out = append(out, s)
	}
	return out
}