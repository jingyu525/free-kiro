package text

import (
	"reflect"
	"testing"
)

func TestRangeLines(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"single no newline", "abc", []string{"abc"}},
		{"single with newline", "abc\n", []string{"abc"}},
		{"two lines", "abc\ndef", []string{"abc", "def"}},
		{"two lines trailing newline", "abc\ndef\n", []string{"abc", "def"}},
		{"only newline", "\n", []string{""}},
		{"multiple consecutive newlines", "a\n\nb\n", []string{"a", "", "b"}},
		{"windows line endings kept literal", "a\r\nb", []string{"a\r", "b"}},
		{"unicode preserved", "你好\n世界", []string{"你好", "世界"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RangeLines(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("RangeLines(%q) = %#v; want %#v", tc.in, got, tc.want)
			}
		})
	}
}