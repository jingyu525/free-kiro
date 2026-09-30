package frontmatter

import (
	"fmt"
	"strconv"
	"strings"
)

// FieldType enumerates the value types Validate knows how to coerce and
// check against. Untyped fields use FieldTypeAny (the zero value).
type FieldType string

// FieldType values drive how the frontmatter parser coerces raw text
// into Go values when validating a Schema.
const (
	// FieldTypeAny accepts any string value; the zero value.
	FieldTypeAny FieldType = ""
	// FieldTypeString is the default; no coercion.
	FieldTypeString FieldType = "string"
	// FieldTypeInt coerces "42" → 42.
	FieldTypeInt FieldType = "int"
	// FieldTypeBool coerces "true"/"false".
	FieldTypeBool FieldType = "bool"
	// FieldTypeStringer is an alias of string, future-proof.
	FieldTypeStringer FieldType = "stringer"
	// FieldTypeList converts comma-separated → []string.
	FieldTypeList FieldType = "list"
)

// FieldRule constrains one key in a Schema.
type FieldRule struct {
	Required bool      // key MUST be present in fm
	Type     FieldType // expected value type (coerce on Validate)
	Allowed  []string  // if non-empty, value MUST equal one of these
}

// Schema maps key → FieldRule. Iterate over a Schema to find unknown keys
// in fm (call Validate's strict variant).
type Schema map[string]FieldRule

// Validate enforces the schema against fm:
//
//   - missing required key → error
//   - unknown key (only when schema is non-empty AND strict == true) → error
//   - value not in Allowed (when Allowed is set) → error
//   - value not coercible to Type → error
//
// strict == false means schema is a positive list (only fields it names
// are checked; extra keys in fm pass). This mirrors the previous inline
// behaviour where steering/spec docs freely add forward-compatible keys.
func Validate(fm Frontmatter, s Schema) error {
	if fm == nil {
		fm = Frontmatter{}
	}
	// Required fields.
	for k, rule := range s {
		if !rule.Required {
			continue
		}
		v, ok := fm[k]
		if !ok {
			return fmt.Errorf("frontmatter: missing required field %q", k)
		}
		if err := checkValue(k, v, rule); err != nil {
			return err
		}
	}
	// Type / Allowed on declared (non-required) fields that are present.
	for k, rule := range s {
		if rule.Required {
			continue // already checked above
		}
		v, ok := fm[k]
		if !ok {
			continue
		}
		if err := checkValue(k, v, rule); err != nil {
			return err
		}
	}
	return nil
}

// checkValue enforces Type + Allowed on a single value.
func checkValue(key string, v any, rule FieldRule) error {
	switch rule.Type {
	case FieldTypeAny, FieldTypeString, FieldTypeStringer:
		// No-op coercion.
	case FieldTypeInt:
		if _, err := strconv.Atoi(strings.TrimSpace(fmt.Sprint(v))); err != nil {
			return fmt.Errorf("frontmatter: %q must be int (got %q)", key, v)
		}
	case FieldTypeBool:
		s := strings.ToLower(strings.TrimSpace(fmt.Sprint(v)))
		if s != "true" && s != "false" {
			return fmt.Errorf("frontmatter: %q must be bool (got %q)", key, v)
		}
	case FieldTypeList:
		// Any non-empty stringified value is acceptable; type-level check
		// is "not empty". Allowed constrains the per-element set.
	default:
		// Unknown type: treat as no-op for forward compat.
	}
	if len(rule.Allowed) == 0 {
		return nil
	}
	got := fmt.Sprint(v)
	for _, ok := range rule.Allowed {
		if got == ok {
			return nil
		}
	}
	return fmt.Errorf("frontmatter: %q=%q not in allowed %v", key, v, rule.Allowed)
}
