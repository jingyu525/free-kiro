package visualize

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// atoiLocal parses a positive integer with a default fallback. Used
// when reading CLI flags that have a numeric default but accept
// arbitrary strings from the user.
func atoiLocal(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// detectBrowserAutoOpen reports whether the platform / env suggests
// auto-opening a browser on serve. Off by default — the user must
// pass --open to trigger.
func detectBrowserAutoOpen() bool { return false }

// openBrowser tries to launch the default browser via platform-specific
// commands. Best-effort: errors are silently ignored so the dashboard
// still serves on headless boxes.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return
	}
	_ = cmd.Start()
}

// detectBrowserOpen kept for future heuristics (env vars, TTY check).
func detectBrowserOpen() bool { return detectBrowserAutoOpen() }
