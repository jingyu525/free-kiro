package visualize

import "time"

// timeMustParse is a tiny helper for tests; parses RFC3339 or fails.
func timeMustParse(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}