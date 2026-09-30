package hooks

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"

	"github.com/jingyu525/free-kiro/internal/models"
)

// rawHook is the loose on-disk shape accepted by the loader. Either
// free-kiro's flat fields or Kiro's nested envelope fields may be
// present; _normalize folds both into a typed models.Hook.
type rawHook map[string]any

// envelopeFile mirrors Kiro's official {version, hooks[]} shape.
type envelopeFile struct {
	Version string    `json:"version"`
	Hooks   []rawHook `json:"hooks"`
}

// normaliseHook coerces a single raw entry into a typed Hook. Accepts
// either free-kiro's flat fields or Kiro's nested envelope fields.
func normaliseHook(raw rawHook) (*models.Hook, error) {
	// Action: either nested {type, command/prompt} (Kiro / free-kiro flat
	// both support this shape) OR flat `action_type` + `action` fields.
	var actionType, actionText string
	if a, ok := raw["action"].(map[string]any); ok {
		if t, ok := a["type"].(string); ok {
			actionType = t
		}
		if c, ok := a["command"].(string); ok {
			actionText = c
		} else if p, ok := a["prompt"].(string); ok {
			actionText = p
		}
	} else {
		if t, ok := raw["action_type"].(string); ok {
			actionType = t
		} else if t, ok := raw["actionType"].(string); ok {
			actionType = t
		}
		if c, ok := raw["action"].(string); ok {
			actionText = c
		}
	}
	// Alias Kiro's "command" → free-kiro's "shell".
	if actionType == "command" {
		actionType = "shell"
	}
	if actionType != "shell" && actionType != "agent" {
		return nil, ferrors.New("hooks.envelope",
			fmt.Sprintf("action type must be 'shell'/'command' or 'agent'; got %q", actionType))
	}

	// Event: free-kiro uses "event"; Kiro uses "trigger" (PascalCase).
	event, _ := raw["event"].(string)
	if event == "" {
		event, _ = raw["trigger"].(string)
	}
	if event == "" {
		return nil, ferrors.New("hooks.envelope", "hook missing event/trigger")
	}

	// Filter: Kiro uses "matcher" (regex); free-kiro uses "glob".
	isRegex := false
	var filter string
	if m, ok := raw["matcher"].(string); ok {
		filter = m
		isRegex = true
	} else if g, ok := raw["glob"].(string); ok {
		filter = g
	} else if g, ok := raw["pattern"].(string); ok {
		filter = g
	}

	id, _ := raw["id"].(string)
	if id == "" {
		id, _ = raw["name"].(string)
	}
	if id == "" {
		return nil, ferrors.New("hooks.envelope", "hook missing id/name")
	}

	desc, _ := raw["description"].(string)

	enabled := true
	if v, ok := raw["enabled"].(bool); ok {
		enabled = v
	} else if s, ok := raw["enabled"].(string); ok {
		enabled = !strings.EqualFold(s, "false")
	}

	var timeout *int
	switch v := raw["timeout"].(type) {
	case int:
		timeout = &v
	case float64:
		i := int(v)
		timeout = &i
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			timeout = &i
		}
	}

	return &models.Hook{
		ID:          id,
		Event:       event,
		Glob:        filter,
		ActionType:  actionType,
		Action:      actionText,
		Description: desc,
		IsRegex:     isRegex,
		Timeout:     timeout,
		Enabled:     enabled,
	}, nil
}

// encodeEnvelope renders a single hook as Kiro's official envelope.
// Lets free-kiro-authored files be loaded unchanged by official Kiro.
func encodeEnvelope(h *models.Hook) ([]byte, error) {
	entry := map[string]any{
		"name":        h.ID,
		"trigger":     h.Event,
		"action":      map[string]any{},
		"enabled":     h.Enabled,
		"description": h.Description,
	}
	action := entry["action"].(map[string]any)
	if h.ActionType == "shell" {
		action["type"] = "command"
		action["command"] = h.Action
	} else {
		action["type"] = "agent"
		action["prompt"] = h.Action
	}
	if h.Glob != "" {
		if h.IsRegex {
			entry["matcher"] = h.Glob
		} else {
			entry["glob"] = h.Glob
		}
	}
	if h.Timeout != nil {
		entry["timeout"] = *h.Timeout
	}
	env := envelopeFile{Version: "v1", Hooks: []rawHook{entry}}
	return json.MarshalIndent(env, "", "  ")
}
