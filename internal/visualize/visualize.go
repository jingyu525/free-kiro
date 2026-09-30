// Package visualize renders free-kiro spec state as ASCII trees, Mermaid
// diagrams, and Markdown reports. Designed for embedding into PRs / docs
// / Slack without any external dependencies.
package visualize

import (
	"strings"
)

// TreeNode is a single node in an ASCII tree. Children are rendered
// immediately below the parent; Meta is rendered as `key: value` pairs
// on indented continuation lines when non-empty.
type TreeNode struct {
	Text     string
	Children []*TreeNode
	Meta     map[string]string // optional "key: value" annotations
}

// RenderTree draws a tree using box-drawing characters. The root node
// is rendered without any connector (it's the trunk); children and
// descendants use ├─ / └─ / │ branches.
func RenderTree(root *TreeNode) string {
	var sb strings.Builder
	renderTreeNode(&sb, root, "", true, true)
	return sb.String()
}

// renderTreeNode writes a single node and recurses into its children.
//
// Parameters:
//
//	prefix   — vertical-bar indent that should precede this node's
//	           connector ("" for top-level).
//	isLast   — whether this node is the last sibling (controls ├─ vs └─).
//	isRoot   — whether this node is the root (skips any connector).
func renderTreeNode(sb *strings.Builder, n *TreeNode, prefix string, isLast, isRoot bool) {
	var connector string
	switch {
	case isRoot:
		connector = ""
	case isLast:
		connector = prefix + "└─ "
	default:
		connector = prefix + "├─ "
	}
	sb.WriteString(connector)
	sb.WriteString(n.Text)
	sb.WriteByte('\n')

	// Meta lines: align under the text (skip the connector).
	if len(n.Meta) > 0 {
		indent := strings.Repeat(" ", runeLen(connector))
		keys := sortedKeys(n.Meta)
		for _, k := range keys {
			sb.WriteString(indent)
			sb.WriteString(k)
			sb.WriteString(": ")
			sb.WriteString(n.Meta[k])
			sb.WriteByte('\n')
		}
	}

	// Child prefix: under the connector, no bar for the last sibling.
	childPrefix := prefix
	if !isRoot {
		if isLast {
			childPrefix = prefix + "   "
		} else {
			childPrefix = prefix + "│  "
		}
	}
	for i, c := range n.Children {
		renderTreeNode(sb, c, childPrefix, i == len(n.Children)-1, false)
	}
}

// runeLen counts Unicode code points (not bytes). Used for visual
// indentation width — Chinese characters occupy 2 cells in monospace
// fonts but count as 1 rune, which is close enough for tree output.
func runeLen(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j-1] > keys[j]; j-- {
			keys[j-1], keys[j] = keys[j], keys[j-1]
		}
	}
	return keys
}
