package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TreePanel provides a tree-view navigation panel.
type TreePanel struct {
	TreeView   *tview.TreeView
	RootPath   string
	ShowHidden bool
}

// NewTreePanel creates a tree panel rooted at the given path.
func NewTreePanel(rootPath string) *TreePanel {
	tp := &TreePanel{
		TreeView: tview.NewTreeView(),
		RootPath: rootPath,
	}

	tp.TreeView.SetBorder(true)
	tp.TreeView.SetBorderPadding(0, 0, 1, 1)
	tp.TreeView.SetTitle(rootPath)
	tp.TreeView.SetGraphics(true)
	tp.TreeView.SetGraphicsColor(tcell.ColorDefault)

	root := tp.buildNode(rootPath, filepath.Base(rootPath))
	root.SetExpanded(true)
	tp.TreeView.SetRoot(root)
	tp.TreeView.SetCurrentNode(root)

	tp.TreeView.SetSelectedFunc(func(node *tview.TreeNode) {
		if node.GetText() == ".." {
			// Navigate up
			parent := filepath.Dir(tp.RootPath)
			if parent != tp.RootPath {
				tp.SetRootPath(parent)
			}
			return
		}

		ref := node.GetReference()
		if ref == nil {
			return
		}
		path := ref.(string)

		info, err := os.Stat(path)
		if err != nil {
			return
		}

		if info.IsDir() {
			if len(node.GetChildren()) > 0 {
				node.SetExpanded(!node.IsExpanded())
			} else {
				tp.expandNode(node, path)
				node.SetExpanded(true)
			}
		}
	})

	return tp
}

// buildNode creates a tree node for the given path.
func (tp *TreePanel) buildNode(path, name string) *tview.TreeNode {
	node := tview.NewTreeNode(name).
		SetReference(path).
		SetSelectable(true)

	info, err := os.Stat(path)
	if err != nil {
		node.SetColor(tcell.ColorDarkGray)
		return node
	}

	if info.IsDir() {
		node.SetColor(tcell.ColorBlue)
		icon := FileIcon(name, true, false, false)
		node.SetText(icon + name + "/")
	} else {
		icon := FileIcon(name, false, false, false)
		node.SetText(icon + name)
	}

	return node
}

// expandNode populates children of a directory node.
func (tp *TreePanel) expandNode(node *tview.TreeNode, path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}

	// Add ".." for non-root
	if path != "/" {
		dotdot := tview.NewTreeNode("..").
			SetReference(filepath.Dir(path)).
			SetSelectable(true).
			SetColor(tcell.ColorBlue)
		node.AddChild(dotdot)
	}

	// Directories first, then files
	var dirs, files []os.DirEntry
	for _, e := range entries {
		if !tp.ShowHidden && strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if e.IsDir() {
			dirs = append(dirs, e)
		} else {
			files = append(files, e)
		}
	}

	for _, d := range dirs {
		childPath := filepath.Join(path, d.Name())
		child := tp.buildNode(childPath, d.Name())
		node.AddChild(child)
	}
	for _, f := range files {
		childPath := filepath.Join(path, f.Name())
		child := tp.buildNode(childPath, f.Name())
		node.AddChild(child)
	}
}

// SetRootPath changes the tree root to a new directory.
func (tp *TreePanel) SetRootPath(path string) {
	tp.RootPath = path
	root := tp.buildNode(path, filepath.Base(path))
	tp.expandNode(root, path)
	root.SetExpanded(true)
	tp.TreeView.SetRoot(root)
	tp.TreeView.SetCurrentNode(root)
	tp.TreeView.SetTitle(path)
}

// SelectedPath returns the path of the currently selected node.
func (tp *TreePanel) SelectedPath() string {
	node := tp.TreeView.GetCurrentNode()
	if node == nil {
		return tp.RootPath
	}
	ref := node.GetReference()
	if ref == nil {
		return tp.RootPath
	}
	return ref.(string)
}

// NavigateToPath changes the tree root to include the given path.
func (tp *TreePanel) NavigateToPath(path string) {
	// If path is under current root, try to find and select it
	// Otherwise, reset the root
	if !strings.HasPrefix(path, tp.RootPath) {
		tp.SetRootPath(path)
		return
	}

	// For simplicity, just reset root to the target
	tp.SetRootPath(path)
}

// ToggleHidden flips hidden file visibility and refreshes.
func (tp *TreePanel) ToggleHidden() {
	tp.ShowHidden = !tp.ShowHidden
	tp.Refresh()
}

// Refresh reloads the tree from the current root path.
func (tp *TreePanel) Refresh() {
	tp.SetRootPath(tp.RootPath)
}
