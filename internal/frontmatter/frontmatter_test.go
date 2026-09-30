package frontmatter

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestParse_EmptyBody(t *testing.T) {
	fm, body, err := Parse(strings.NewReader("---\nmode: always\n---\n"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := fm["mode"]; got != "always" {
		t.Errorf("mode = %v, want always", got)
	}
	if string(body) != "" {
		t.Errorf("body = %q, want empty", body)
	}
}

func TestParse_MultilineBody(t *testing.T) {
	in := "---\nmode: always\n---\n# Heading\n\nline1\nline2\n"
	fm, body, err := Parse(strings.NewReader(in))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if fm["mode"] != "always" {
		t.Errorf("mode = %v, want always", fm["mode"])
	}
	if !strings.Contains(string(body), "# Heading") || !strings.Contains(string(body), "line2") {
		t.Errorf("body missing expected lines: %q", body)
	}
}

func TestParse_BOM(t *testing.T) {
	// UTF-8 BOM: EF BB BF
	in := "\xEF\xBB\xBF---\nmode: auto\n---\nbody\n"
	fm, body, err := Parse(strings.NewReader(in))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if fm["mode"] != "auto" {
		t.Errorf("mode = %v, want auto", fm["mode"])
	}
	if string(body) != "body\n" {
		t.Errorf("body = %q, want %q", body, "body\n")
	}
}

func TestParse_CRLF(t *testing.T) {
	// CRLF line endings (Windows). bufio.Scanner handles \r\n if we set
	// the SplitFunc; the default Scanner splits on \n only, so the \r
	// becomes part of the line. parseLines treats the \r as trailing
	// whitespace via strings.TrimSpace on the fence check and parseKV's
	// value trim, so values come out clean.
	in := "---\r\nmode: always\r\ndescription: hi\r\n---\r\nbody\r\n"
	fm, body, err := Parse(strings.NewReader(in))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if fm["mode"] != "always" {
		t.Errorf("mode = %v, want always", fm["mode"])
	}
	if fm["description"] != "hi" {
		t.Errorf("description = %v, want hi", fm["description"])
	}
	if !strings.Contains(string(body), "body") {
		t.Errorf("body = %q, missing 'body'", body)
	}
}

func TestParse_MissingFence(t *testing.T) {
	_, _, err := Parse(strings.NewReader("no fence here\njust text\n"))
	if !errors.Is(err, ErrNoFrontmatter) {
		t.Errorf("err = %v, want ErrNoFrontmatter", err)
	}
}

func TestParse_InvalidKey(t *testing.T) {
	// Keys with leading digit or space are skipped (parseKV returns ok=false),
	// but a valid key followed by garbage still parses the valid one.
	in := "---\n1bad: skip\nmode: auto\n---\n"
	fm, _, err := Parse(strings.NewReader(in))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if _, present := fm["1bad"]; present {
		t.Errorf("expected 1bad to be skipped, got fm=%v", fm)
	}
	if fm["mode"] != "auto" {
		t.Errorf("mode = %v, want auto", fm["mode"])
	}
}

func TestValidate_MissingRequired(t *testing.T) {
	fm := Frontmatter{"mode": "always"}
	s := Schema{
		"mode":        {Required: true},
		"description": {Required: true},
	}
	err := Validate(fm, s)
	if err == nil || !strings.Contains(err.Error(), "description") {
		t.Errorf("err = %v, want missing required description", err)
	}
}

func TestValidate_UnknownFieldRejected(t *testing.T) {
	// Schema with extra key `unexpected`: value present in fm → allowed
	// (Validate is non-strict; unknown fields pass unless schema
	// explicitly checks key set).
	fm := Frontmatter{"mode": "always", "unexpected": "x"}
	s := Schema{
		"mode": {Required: true, Allowed: []string{"always", "auto"}},
	}
	if err := Validate(fm, s); err != nil {
		t.Errorf("Validate = %v, want nil (non-strict)", err)
	}
}

func TestValidate_Allowed(t *testing.T) {
	fm := Frontmatter{"mode": "manual"}
	s := Schema{
		"mode": {Required: true, Allowed: []string{"always", "auto", "manual", "filematch"}},
	}
	if err := Validate(fm, s); err != nil {
		t.Errorf("Validate(mode=manual) = %v, want nil", err)
	}
	fm["mode"] = "bogus"
	if err := Validate(fm, s); err == nil {
		t.Errorf("Validate(mode=bogus) = nil, want error")
	}
}

func TestValidate_TypeCoercion(t *testing.T) {
	s := Schema{"count": {Required: true, Type: FieldTypeInt}}
	if err := Validate(Frontmatter{"count": "42"}, s); err != nil {
		t.Errorf("int=42: %v", err)
	}
	if err := Validate(Frontmatter{"count": "not-a-number"}, s); err == nil {
		t.Errorf("int=not-a-number: want error")
	}
	b := Schema{"on": {Required: true, Type: FieldTypeBool}}
	if err := Validate(Frontmatter{"on": "true"}, b); err != nil {
		t.Errorf("bool=true: %v", err)
	}
	if err := Validate(Frontmatter{"on": "maybe"}, b); err == nil {
		t.Errorf("bool=maybe: want error")
	}
}

func TestMarshal_RoundTrip(t *testing.T) {
	in := Frontmatter{"mode": "always", "description": "hello"}
	body := []byte("# Body\n\nstuff\n")
	out, err := Marshal(in, body)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("---\n")) {
		t.Errorf("missing opening fence: %q", out)
	}
	if !bytes.Contains(out, []byte("mode: always")) {
		t.Errorf("missing mode line: %q", out)
	}
	if !bytes.HasSuffix(out, body) {
		t.Errorf("body not at end: %q", out)
	}
	// Round-trip back through Parse.
	fm, gotBody, err := Parse(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("Parse round-trip: %v", err)
	}
	if fm["mode"] != "always" || fm["description"] != "hello" {
		t.Errorf("fm round-trip = %v", fm)
	}
	if !bytes.Equal(gotBody, body) {
		t.Errorf("body round-trip = %q, want %q", gotBody, body)
	}
}

func TestMarshal_InvalidKey(t *testing.T) {
	_, err := Marshal(Frontmatter{"1bad": "x"}, nil)
	if err == nil || !strings.Contains(err.Error(), "invalid key") {
		t.Errorf("err = %v, want invalid key", err)
	}
}

func TestParse_QuotedValue(t *testing.T) {
	in := "---\ndescription: \"hello world\"\nmode: 'auto'\n---\n"
	fm, _, err := Parse(strings.NewReader(in))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if fm["description"] != "hello world" {
		t.Errorf("description = %v, want 'hello world'", fm["description"])
	}
	if fm["mode"] != "auto" {
		t.Errorf("mode = %v, want auto", fm["mode"])
	}
}

func TestParse_UnclosedFence(t *testing.T) {
	_, _, err := Parse(strings.NewReader("---\nmode: always\n"))
	if !errors.Is(err, ErrNoFrontmatter) {
		t.Errorf("err = %v, want ErrNoFrontmatter", err)
	}
}
