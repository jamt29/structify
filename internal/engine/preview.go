package engine

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamt29/structify/internal/template"
)

// FileTree is a hierarchical representation of files for scaffold preview.
type FileTree struct {
	Root     string
	Children []*TreeNode
	Total    int
	Steps    int
}

// TreeNode represents a directory or file in preview tree output.
type TreeNode struct {
	Name     string
	IsDir    bool
	Children []*TreeNode
	Skipped  bool
}

// PreviewFiles computes scaffold output without writing to disk.
// It reuses ProcessFiles and ExecuteSteps in dry-run mode and returns
// a hierarchical tree suitable for TUI rendering.
func (e *Engine) PreviewFiles(req *template.ScaffoldRequest) (*FileTree, error) {
	if req == nil {
		return nil, nil
	}
	previewReq := &template.ScaffoldRequest{
		Template:  req.Template,
		OutputDir: req.OutputDir,
		Variables: req.Variables,
		DryRun:    true,
	}
	created, skipped, err := ProcessFiles(previewReq)
	if err != nil {
		return nil, err
	}

	steps := 0
	if previewReq.Template != nil && previewReq.Template.Manifest != nil {
		stepResults, err := ExecuteSteps(previewReq.Template.Manifest.Steps, previewReq.Variables, previewReq.OutputDir, true)
		if err != nil {
			return nil, err
		}
		for _, s := range stepResults {
			if !s.Skipped && s.Error == nil {
				steps++
			}
		}
	}

	rootName := strings.TrimSpace(filepath.Base(req.OutputDir))
	if rootName == "." || rootName == "/" || rootName == "" {
		rootName = "output"
	}

	tree := &FileTree{
		Root:  rootName,
		Total: len(created),
		Steps: steps,
	}

	for _, p := range created {
		insertPath(tree, p, false)
	}
	for _, p := range skipped {
		insertPath(tree, p, true)
	}
	sortTree(tree.Children)
	computeSkipState(tree.Children)
	return tree, nil
}

func insertPath(tree *FileTree, relPath string, skipped bool) {
	parts := splitRelPath(relPath)
	if len(parts) == 0 {
		return
	}
	nodes := &tree.Children
	for i, part := range parts {
		last := i == len(parts)-1
		isDir := !last
		n := findOrCreateNode(nodes, part, isDir)
		if skipped && last {
			n.Skipped = true
		}
		nodes = &n.Children
	}
}

func splitRelPath(relPath string) []string {
	clean := filepath.ToSlash(strings.TrimSpace(relPath))
	clean = strings.TrimPrefix(clean, "./")
	clean = strings.Trim(clean, "/")
	if clean == "" {
		return nil
	}
	return strings.Split(clean, "/")
}

func previewDisplayName(name string, isDir bool) string {
	if isDir {
		return name
	}
	return strings.TrimSuffix(name, ".tmpl")
}

func findOrCreateNode(nodes *[]*TreeNode, name string, isDir bool) *TreeNode {
	display := previewDisplayName(name, isDir)
	for _, n := range *nodes {
		if n.Name == display {
			// Keep directory semantics if discovered later.
			if isDir {
				n.IsDir = true
			}
			return n
		}
	}
	n := &TreeNode{Name: display, IsDir: isDir}
	*nodes = append(*nodes, n)
	return n
}

func sortTree(nodes []*TreeNode) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir
		}
		return nodes[i].Name < nodes[j].Name
	})
	for _, n := range nodes {
		sortTree(n.Children)
	}
}

func computeSkipState(nodes []*TreeNode) bool {
	if len(nodes) == 0 {
		return false
	}
	allSkipped := true
	for _, n := range nodes {
		if n.IsDir && len(n.Children) > 0 {
			n.Skipped = computeSkipState(n.Children)
		}
		if !n.Skipped {
			allSkipped = false
		}
	}
	return allSkipped
}
