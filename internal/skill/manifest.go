package skill

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

// Manifest is the parsed contents of skill.json. SHA256 is populated at
// release time by contrib/skills/scripts/compute-skill-sha.sh and used
// by VerifyBundle to check the bundle after download/extract.
type Manifest struct {
	Name               string            `json:"name"`
	Version            string            `json:"version"`
	FreeKiroMinVersion string            `json:"free_kiro_min_version"`
	License            string            `json:"license"`
	Repository         string            `json:"repository"`
	SHA256             map[string]string `json:"sha256"`
}

// LoadManifest parses skill.json at `dir/skill.json`.
func LoadManifest(dir string) (*Manifest, error) {
	data, err := os.ReadFile(filepath.Join(dir, "skill.json"))
	if err != nil {
		return nil, ferrors.Wrap("skill.manifest", err, "read skill.json in "+dir)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, ferrors.Wrap("skill.manifest", err, "parse skill.json in "+dir)
	}
	if m.Name == "" || m.Version == "" {
		return nil, ferrors.New("skill.manifest",
			"skill.json missing required fields (name, version) in "+dir)
	}
	return &m, nil
}

// VerifyBundle checks every file in `dir` against the sha256 map in
// the manifest. Returns the list of file names that failed verification
// (empty slice = all OK). An empty sha256 map (dev builds / before
// compute-skill-sha.sh runs) is treated as a warning, not an error —
// the installer proceeds without per-file verification.
func (m *Manifest) VerifyBundle(dir string) (failures []string, err error) {
	if len(m.SHA256) == 0 {
		return nil, nil
	}
	for rel, want := range m.SHA256 {
		got, err := hashFile(filepath.Join(dir, rel))
		if err != nil {
			failures = append(failures, rel+": "+err.Error())
			continue
		}
		if got != want {
			failures = append(failures, rel+": got "+got+", want "+want)
		}
	}
	sort.Strings(failures)
	return failures, nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}