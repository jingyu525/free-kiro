// Package hooks implements event-driven automations compatible with both
// free-kiro's flat shape and Kiro's official {version:"v1", hooks:[…]}
// envelope.
//
// Hooks are JSON files under .kiro/hooks/*.json. Each declares an event
// (free-kiro: file.save / file.create / file.delete / prompt.submit /
// task.run / manual + Kiro's PascalCase triggers like PostFileSave /
// PreToolUse), an optional filter, and an action that is either a shell
// command or an agent prompt.
//
// Positioning: free-kiro is the *passive planning/spec layer* — it never
// fires events on its own and never blocks a tool. The engine only runs
// a hook when you call `free-kiro hook run <event>` (typically driven by
// the IDE's own hook system). Shell actions receive the event context as
// JSON on STDIN; agent actions are a delegate point (free-kiro does not
// run models).
package hooks

import "regexp"

// globToRegex translates a glob (supporting * and **) into an anchored
// regex. * matches within a path segment; ** matches across segments
// (including /).
func globToRegex(pattern string) string {
	out := "^"
	i := 0
	for i < len(pattern) {
		c := pattern[i]
		switch c {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				out += ".*"
				i += 2
				if i < len(pattern) && pattern[i] == '/' {
					i++
				}
				continue
			}
			out += "[^/]*"
			i++
		case '?':
			out += "[^/]"
			i++
		case '.', '+', '(', ')', '|', '^', '$', '{', '}', '\\':
			out += "\\" + string(c)
			i++
		default:
			out += string(c)
			i++
		}
	}
	out += "$"
	return out
}

// globMatch reports whether path matches the given glob pattern.
func globMatch(pattern, path string) bool {
	rx, err := regexp.Compile(globToRegex(pattern))
	if err != nil {
		return false
	}
	return rx.MatchString(path)
}