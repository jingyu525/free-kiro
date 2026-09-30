package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

// bskDefaultBinary is the bsk CLI name (PATH lookup). Override via
// FREE_KIRO_BSK_BIN environment variable for tests.
const bskDefaultBinary = "bsk"

// FetchBrowserHTML uses browser-skill's `bsk` CLI to render <rawURL> in
// the user's logged-in browser and return the resulting HTML's visible
// text + title. This is the `--from-browser` flow for `free-kiro spec
// new` — it lets the agent capture JS-rendered pages (Notion SPAs,
// authenticated content) that net/http cannot.
//
// Requires `bsk` on PATH; otherwise returns an actionable error pointing
// to the browser-skill install instructions and suggesting --from-prd
// as a no-bsk fallback for non-JS pages.
func FetchBrowserHTML(ctx context.Context, rawURL string) (title, body string, err error) {
	bin := os.Getenv("FREE_KIRO_BSK_BIN")
	if bin == "" {
		bin = bskDefaultBinary
	}
	path, lookErr := exec.LookPath(bin)
	if lookErr != nil {
		return "", "", ferrors.New("spec.from-browser",
			"`bsk` not found on PATH; install browser-skill first:\n"+
				"  see https://github.com/jingyu525/browser-skill  (or run `bsk --help` after install)\n"+
				"  tip: --from-prd <url> works for non-JS pages without bsk")
	}

	// 1. Navigate. We use `bsk navigate` against the user's existing
	// session (browser-skill re-uses one session per command by default).
	nav := exec.CommandContext(ctx, path, "navigate", rawURL)
	if out, err := nav.CombinedOutput(); err != nil {
		return "", "", ferrors.Wrap("spec.from-browser", err,
			fmt.Sprintf("bsk navigate failed for %s: %s", rawURL, trimTrailing(string(out))))
	}

	// 2. Dump rendered HTML to a temp file.
	tmp, err := os.CreateTemp("", "free-kiro-browser-*.html")
	if err != nil {
		return "", "", ferrors.Wrap("spec.from-browser", err, "create temp html")
	}
	tmpPath := tmp.Name()
	_ = tmp.Close()
	defer func() { _ = os.Remove(tmpPath) }() // best-effort cleanup

	dump := exec.CommandContext(ctx, path, "get-html", "--out", tmpPath)
	if out, err := dump.CombinedOutput(); err != nil {
		return "", "", ferrors.Wrap("spec.from-browser", err,
			"bsk get-html failed: "+trimTrailing(string(out)))
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", "", ferrors.Wrap("spec.from-browser", err, "read rendered html")
	}

	// 3. Reuse the existing HTML→(title, body) pipeline.
	return extractHTML(data)
}
