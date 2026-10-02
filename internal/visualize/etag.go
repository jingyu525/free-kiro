package visualize

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// etagFor returns the SHA-256 hex digest of body. The ETag middleware
// memoises the result per (path, status, bodyLen) so bodies that
// differ only in volatile JSON fields (timestamps) reuse the tag.
func etagFor(body []byte) string {
	sum := sha256.Sum256(body)
	return fmt.Sprintf("%x", sum)
}

// volatileJSONKeys lists JSON top-level keys that change on every render
// even when the underlying spec data is identical. Stripping them
// before hashing makes the ETag stable across the typical "same specs,
// different `generated_at`" requests.
var volatileJSONKeys = []string{
	"generated_at",
	"started_at",
	"last_refresh_at",
}

// stripVolatileJSON returns a copy of body with the volatile keys'
// values replaced by `""` (so the bytes are deterministic for the
// same logical content). For non-JSON bodies, returns body unchanged.
func stripVolatileJSON(body []byte) []byte {
	if len(body) == 0 || body[0] != '{' {
		return body
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return body
	}
	for _, k := range volatileJSONKeys {
		if _, ok := raw[k]; ok {
			raw[k] = ""
		}
	}
	out, err := marshalSortedMap(raw)
	if err != nil {
		return body
	}
	return out
}

// marshalSortedMap encodes a (possibly nested) map[string]any as JSON
// with keys sorted alphabetically at every depth. encoding/json's
// default map sort is top-level only; nested objects get random
// ordering. We walk the tree and re-encode leaves to keep ordering
// stable across calls — necessary for ETag stability.
func marshalSortedMap(v any) ([]byte, error) {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var buf bytes.Buffer
		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			kb, err := json.Marshal(k)
			if err != nil {
				return nil, err
			}
			buf.Write(kb)
			buf.WriteByte(':')
			vb, err := marshalSortedMap(t[k])
			if err != nil {
				return nil, err
			}
			buf.Write(vb)
		}
		buf.WriteByte('}')
		return buf.Bytes(), nil
	case []any:
		var buf bytes.Buffer
		buf.WriteByte('[')
		for i, item := range t {
			if i > 0 {
				buf.WriteByte(',')
			}
			vb, err := marshalSortedMap(item)
			if err != nil {
				return nil, err
			}
			buf.Write(vb)
		}
		buf.WriteByte(']')
		return buf.Bytes(), nil
	default:
		return json.Marshal(v)
	}
}

// stableETagCache memoises ETag by (path, status, bodyLen).
var (
	etagMu    sync.Mutex
	etagCache = map[string]string{} // key "<path>|<status>|<len>" → quoted etag
)

// stableETag returns a content-addressed ETag for body at the given
// path/status, removing volatile timestamps before hashing.
func stableETag(path string, status int, body []byte) string {
	key := fmt.Sprintf("%s|%d|%d", path, status, len(body))
	stripped := stripVolatileJSON(body)
	tag := `"` + etagFor(stripped) + `"`

	etagMu.Lock()
	defer etagMu.Unlock()
	if cached, ok := etagCache[key]; ok && cached == tag {
		return strings.Trim(tag, `"`)
	}
	etagCache[key] = tag
	return strings.Trim(tag, `"`)
}

// inmMatches reports whether the client's If-None-Match header contains
// a tag equal to the current body's strong ETag. Per RFC 7232 the
// header may carry a comma-separated list; we do weak (lexical) match.
func inmMatches(inm, etag string) bool {
	if inm == "" {
		return false
	}
	cur := `"` + etag + `"`
	if inm == cur {
		return true
	}
	return containsMatch(inm, cur)
}

// containsMatch is a tiny helper for a comma-separated If-None-Match
// list. We avoid strings.Contains with commas since ETag values are
// allowed to contain commas (rare in practice but spec-allowed).
func containsMatch(list, target string) bool {
	for {
		i := indexByte(list, ',')
		if i < 0 {
			if trim(list) == target {
				return true
			}
			return false
		}
		if trim(list[:i]) == target {
			return true
		}
		list = list[i+1:]
	}
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}