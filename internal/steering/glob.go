package steering

import "regexp"

// GlobMatch reports whether path matches the given glob pattern.
//
// Supported syntax (linear scan → single anchored regex, no recursion):
//
//	*    matches any sequence of non-`/` characters
//	?    matches a single non-`/` character
//	**   matches any sequence of characters including `/`
//	      (used to mean "across path segments")
//
// Everything else is literal.
func GlobMatch(pattern, path string) bool {
	rx := compileGlob(pattern)
	return regexp.MustCompile("^" + rx + "$").MatchString(path)
}

// compileGlob translates a glob pattern into an anchored regex body.
// Implementation is a single linear scan that produces a string suitable
// for passing to regexp.MustCompile; terminates by construction.
func compileGlob(pattern string) string {
	out := ""
	i := 0
	for i < len(pattern) {
		c := pattern[i]
		switch c {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				// `**` — match across path separators.
				out += ".*"
				i += 2
				if i < len(pattern) && pattern[i] == '/' {
					i++ // consume an optional trailing slash
				}
			} else {
				// Single `*` — match within a single segment.
				out += "[^/]*"
				i++
			}
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
	return out
}