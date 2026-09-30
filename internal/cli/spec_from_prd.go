package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

// FetchPRD downloads a web page (HTML or Markdown) and returns the
// extracted visible text + page title. Designed for "spec new --from-prd
// <url>" — pulls a product requirements doc from Notion / Confluence /
// any HTTP URL and turns it into a spec prompt.
//
// Content extraction is intentionally simple:
//   - HTML: parse with x/net/html, drop <script>/<style>/<nav>/<footer>/<header>,
//     concatenate visible text.
//   - Plain text / Markdown: return the body as-is.
//
// Not a full Markdown or Notion parser — those would require heavy
// dependencies. Good enough for "give me the gist of this PRD".
func FetchPRD(ctx context.Context, rawURL string) (title, body string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", ferrors.Wrap("spec.from-prd", err, "parse URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", "", ferrors.New("spec.from-prd",
			"only http(s) URLs are supported; got "+u.Scheme)
	}

	// 1 MB cap on response body. PRDs can be long but anything past
	// 1 MB is almost certainly not a readable spec.
	const maxBody = 1 << 20
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return "", "", ferrors.Wrap("spec.from-prd", err, "build request")
	}
	req.Header.Set("User-Agent", "free-kiro (https://github.com/jingyu525/free-kiro)")
	req.Header.Set("Accept", "text/html, text/markdown, text/plain;q=0.8, */*;q=0.5")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", ferrors.Wrap("spec.from-prd", err, "GET "+rawURL)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		return "", "", ferrors.New("spec.from-prd",
			fmt.Sprintf("GET %s returned HTTP %d", rawURL, resp.StatusCode))
	}
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return "", "", ferrors.Wrap("spec.from-prd", err, "read body")
	}

	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "html") || looksLikeHTML(bodyBytes) {
		return extractHTML(bodyBytes)
	}
	// Plain text or markdown: use as-is. Trim trailing whitespace.
	return extractTitleFromFilename(u), strings.TrimSpace(string(bodyBytes)), nil
}

// extractHTML returns the page <title> + concatenated visible body
// text. Walks the parse tree, drops non-content nodes, collects text.
func extractHTML(data []byte) (title, body string, err error) {
	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return "", "", ferrors.Wrap("spec.from-prd", err, "parse HTML")
	}

	// Walk the tree; render only visible text from non-skipped nodes.
	bodyBuf := &strings.Builder{}
	dropBuf := &strings.Builder{}
	walk(doc, bodyBuf, dropBuf)
	cleanedBody := collapseWhitespace(strings.TrimSpace(bodyBuf.String()))
	if cleanedBody == "" {
		cleanedBody = strings.TrimSpace(dropBuf.String())
	}
	title = strings.TrimSpace(extractTitle(doc))
	if title == "" {
		title = "(untitled PRD)"
	}
	if len(cleanedBody) > 32000 {
		cleanedBody = cleanedBody[:32000] + "\n\n[… truncated; full doc lives at the URL you provided …]"
	}
	return title, cleanedBody, nil
}

// walk does a pre-order traversal. Visible text from "content"
// elements goes into bodyBuf; everything else (script/style/nav/footer/
// header/aside) is dropped into dropBuf (used as fallback when body is empty).
func walk(n *html.Node, body, drop *strings.Builder) {
	if n.Type == html.ElementNode {
		tag := strings.ToLower(n.Data)
		switch tag {
		case "script", "style", "noscript", "iframe", "svg":
			// Never render; don't recurse into children either.
			return
		case "nav", "footer", "header", "aside", "form":
			drop.WriteString(collectText(n) + "\n")
			return
		}
	}
	if n.Type == html.TextNode {
		body.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walk(c, body, drop)
	}
	if n.Type == html.ElementNode {
		switch strings.ToLower(n.Data) {
		case "p", "div", "br", "li", "h1", "h2", "h3", "h4", "h5", "h6", "tr":
			body.WriteString("\n")
		}
	}
}

func collectText(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(x *html.Node) {
		if x.Type == html.TextNode {
			sb.WriteString(x.Data)
		}
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return sb.String()
}

func extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && strings.ToLower(n.Data) == "title" {
		return collectText(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if t := extractTitle(c); t != "" {
			return t
		}
	}
	return ""
}

// collapseWhitespace collapses runs of whitespace into single spaces
// and trims leading/trailing whitespace per line. Keeps newlines so
// the spec still has visible structure.
func collapseWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		// Collapse runs of spaces/tabs inside each line.
		collapsed := collapseSpaces(ln)
		collapsed = strings.TrimSpace(collapsed)
		if collapsed != "" {
			out = append(out, collapsed)
		}
	}
	return strings.Join(out, "\n")
}

func collapseSpaces(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	lastSpace := false
	for _, r := range s {
		if r == ' ' || r == '\t' {
			if !lastSpace && b.Len() > 0 {
				b.WriteByte(' ')
			}
			lastSpace = true
			continue
		}
		lastSpace = false
		b.WriteRune(r)
	}
	return b.String()
}

// looksLikeHTML returns true if the body starts with what looks like
// a HTML tag (browsers don't always honor the Content-Type header,
// especially for SPA fragments).
func looksLikeHTML(b []byte) bool {
	s := strings.TrimSpace(string(b))
	return strings.HasPrefix(s, "<!DOCTYPE") || strings.HasPrefix(s, "<html") ||
		(strings.HasPrefix(s, "<") && strings.Contains(s[:pickMin(200, len(s))], ">"))
}

func extractTitleFromFilename(u *url.URL) string {
	p := u.Path
	if p == "" || p == "/" {
		return u.Host
	}
	parts := strings.Split(p, "/")
	last := parts[len(parts)-1]
	last = strings.TrimSuffix(last, ".html")
	last = strings.TrimSuffix(last, ".htm")
	last = strings.TrimSuffix(last, ".md")
	last = strings.ReplaceAll(last, "-", " ")
	last = strings.ReplaceAll(last, "_", " ")
	if last == "" {
		return u.Host
	}
	return last
}

func pickMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
