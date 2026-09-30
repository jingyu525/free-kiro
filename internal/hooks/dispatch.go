package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/jingyu525/free-kiro/internal/models"
)

// Result is the outcome of one hook execution.
type Result struct {
	ID     string
	Event  string
	OK     bool
	Output string
	Error  string
}

// AgentFn is the delegate point for `agent` action hooks. free-kiro
// does not run models itself; the host (CLI, IDE hook runner) wires
// this callback when it wants to dispatch agent prompts to its own
// model. A nil AgentFn produces a placeholder result.
type AgentFn func(prompt string) (string, error)

// Dispatch runs every hook matching (event, file). Shell actions are
// executed via os/exec with the event context as JSON on STDIN. Agent
// actions call agentFn if supplied.
//
// Returns one Result per matched hook; never aborts on a single hook
// failure.
func (r *Registry) Dispatch(ctx context.Context, event string, file string, agentFn AgentFn) ([]Result, error) {
	matched, err := r.Match(event, file)
	if err != nil {
		return nil, err
	}
	cwd := r.ws.Root()
	payload, _ := json.Marshal(map[string]any{
		"event": event,
		"file":  file,
		"cwd":   cwd,
	})

	out := make([]Result, 0, len(matched))
	for _, h := range matched {
		if h.ActionType == "shell" {
			out = append(out, runShellAction(ctx, h, string(payload), cwd))
		} else {
			out = append(out, runAgentAction(agentFn, h))
		}
	}
	return out, nil
}

func runShellAction(ctx context.Context, h *models.Hook, stdin, cwd string) Result {
	res := Result{ID: h.ID, Event: h.Event}
	if h.Disabled {
		// Explicit opt-out (was overloaded on Timeout=0 before
		// housekeeping-cleanup). Distinct from a Timeout that simply
		// wasn't set, so authors get one knob per intent.
		res.OK = false
		res.Error = "hook disabled"
		return res
	}
	timeout := 30 * time.Second
	if h.Timeout != nil {
		timeout = time.Duration(*h.Timeout) * time.Second
	}
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(tctx, "sh", "-c", h.Action)
	cmd.Dir = cwd
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.CombinedOutput()
	res.Output = string(out)
	if err != nil {
		res.OK = false
		res.Error = err.Error()
		if tctx.Err() == context.DeadlineExceeded {
			res.Error = fmt.Sprintf("hook timed out after %s", timeout)
		}
		return res
	}
	res.OK = true
	return res
}

func runAgentAction(fn AgentFn, h *models.Hook) Result {
	res := Result{ID: h.ID, Event: h.Event}
	if fn == nil {
		res.OK = true
		res.Output = fmt.Sprintf(
			"[agent hook] would execute prompt: %q\n"+
				"(free-kiro does not run agents — wire `agent_fn` to your own "+
				"model/runner to enable this delegate point)",
			h.Action)
		return res
	}
	out, err := fn(h.Action)
	if err != nil {
		res.OK = false
		res.Error = err.Error()
		return res
	}
	res.OK = true
	res.Output = out
	return res
}
