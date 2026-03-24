package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jamt29/structify/internal/engine"
)

var (
	treeConnectorStyle = lipgloss.NewStyle().Foreground(colorBorder)
	treeDirStyle       = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	treeFileStyle      = lipgloss.NewStyle().Foreground(colorText)
	treeSkippedStyle   = lipgloss.NewStyle().Foreground(colorMuted).Strikethrough(true)
	treeFooterStyle    = lipgloss.NewStyle().Foreground(colorMuted)
)

// RenderTree renders a scaffold file tree and footer.
func RenderTree(tree *engine.FileTree, width int, maxLines int) string {
	if tree == nil {
		return treeFooterStyle.Render("(sin datos)")
	}

	lines := make([]string, 0, 32)
	lines = append(lines, treeDirStyle.Render(tree.Root+"/"))
	for i, n := range tree.Children {
		last := i == len(tree.Children)-1
		renderTreeNode(n, "", last, &lines)
	}

	footer := treeFooterStyle.Render(fmt.Sprintf("%d archivos · %d steps", tree.Total, tree.Steps))
	if maxLines > 0 {
		keep := maxLines - 1 // reserve one line for footer
		if keep < 1 {
			keep = 1
		}
		if len(lines) > keep {
			hidden := len(lines) - keep
			lines = lines[:keep]
			lines = append(lines, treeFooterStyle.Render(fmt.Sprintf("(+ %d más)", hidden)))
		}
	}
	lines = append(lines, footer)

	out := strings.Join(lines, "\n")
	if width > 0 {
		out = lipgloss.NewStyle().MaxWidth(width).Render(out)
	}
	return out
}

func renderTreeNode(node *engine.TreeNode, prefix string, last bool, lines *[]string) {
	conn := "├── "
	nextPrefix := prefix + "│   "
	if last {
		conn = "└── "
		nextPrefix = prefix + "    "
	}

	label := node.Name
	if node.IsDir {
		label += "/"
	}

	var labelRendered string
	if node.Skipped {
		labelRendered = treeSkippedStyle.Render(label)
	} else if node.IsDir {
		labelRendered = treeDirStyle.Render(label)
	} else {
		labelRendered = treeFileStyle.Render(label)
	}

	line := prefix + treeConnectorStyle.Render(conn) + labelRendered
	*lines = append(*lines, line)
	for i, child := range node.Children {
		renderTreeNode(child, nextPrefix, i == len(node.Children)-1, lines)
	}
}
