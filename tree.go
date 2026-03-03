package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TreePanel provides a persistent tree-view navigation panel that can browse
// the entire filesystem. Directories are expanded/collapsed in place rather
// than resetting the root on every navigation.
type TreePanel struct {
	TreeView     *tview.TreeView
	RootPath     string          // filesystem root ("/" or $HOME)
	ShowHidden   bool
	expandedSet  map[string]bool // tracks expanded dirs by absolute path
	OnFileSelect func(path string)
}

// NewTreePanel creates a tree panel rooted at rootPath with startPath pre-expanded.
func NewTreePanel(rootPath, startPath string) *TreePanel {
	tp := &TreePanel{
		TreeView:    tview.NewTreeView(),
		RootPath:    rootPath,
		expandedSet: make(map[string]bool),
	}

	tp.TreeView.SetBorder(true)
	tp.TreeView.SetBorderPadding(0, 0, 1, 1)
	tp.TreeView.SetTitle(rootPath)
	tp.TreeView.SetGraphics(true)
	tp.TreeView.SetGraphicsColor(tcell.ColorDefault)

	root := tp.buildNode(rootPath, filepath.Base(rootPath))
	tp.expandNode(root, rootPath)
	root.SetExpanded(true)
	tp.expandedSet[rootPath] = true
	tp.TreeView.SetRoot(root)
	tp.TreeView.SetCurrentNode(root)

	tp.TreeView.SetSelectedFunc(func(node *tview.TreeNode) {
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
			tp.ToggleExpand(node, path)
		} else {
			// File selected — invoke callback
			if tp.OnFileSelect != nil {
				tp.OnFileSelect(path)
			}
		}
	})

	// Auto-expand to startPath
	if startPath != rootPath {
		tp.ExpandToPath(startPath)
	}

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

// expandNode populates children of a directory node (lazy loading).
func (tp *TreePanel) expandNode(node *tview.TreeNode, path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return
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

// ToggleExpand collapses an expanded dir or expands a collapsed dir.
func (tp *TreePanel) ToggleExpand(node *tview.TreeNode, path string) {
	if tp.expandedSet[path] {
		// Collapse: clear children and remove from set
		node.ClearChildren()
		node.SetExpanded(false)
		delete(tp.expandedSet, path)
	} else {
		// Expand: lazy-load children
		tp.expandNode(node, path)
		node.SetExpanded(true)
		tp.expandedSet[path] = true
	}
}

// ExpandToPath walks from root expanding each segment to reach target,
// then sets the cursor on the target node.
func (tp *TreePanel) ExpandToPath(target string) {
	target = filepath.Clean(target)
	root := tp.TreeView.GetRoot()
	if root == nil {
		return
	}

	// Build the list of path segments to expand from root to target.
	// E.g. root="/home/user", target="/home/user/projects/foo"
	// segments = ["/home/user", "/home/user/projects", "/home/user/projects/foo"]
	segments := tp.pathSegments(target)

	current := root
	for _, seg := range segments {
		ref := current.GetReference()
		if ref == nil {
			break
		}
		currentPath := ref.(string)

		// If this node isn't expanded yet, expand it
		if !tp.expandedSet[currentPath] {
			tp.expandNode(current, currentPath)
			current.SetExpanded(true)
			tp.expandedSet[currentPath] = true
		}

		// Find the child matching the next segment
		if seg == currentPath {
			continue // This is the root itself
		}

		found := false
		for _, child := range current.GetChildren() {
			cRef := child.GetReference()
			if cRef == nil {
				continue
			}
			if cRef.(string) == seg {
				current = child
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	tp.TreeView.SetCurrentNode(current)
}

// pathSegments returns all intermediate paths from RootPath to target.
// E.g. root="/home", target="/home/user/docs" → ["/home", "/home/user", "/home/user/docs"]
func (tp *TreePanel) pathSegments(target string) []string {
	target = filepath.Clean(target)
	rootClean := filepath.Clean(tp.RootPath)

	if !strings.HasPrefix(target, rootClean) {
		return []string{target}
	}

	rel, err := filepath.Rel(rootClean, target)
	if err != nil {
		return []string{target}
	}

	if rel == "." {
		return []string{rootClean}
	}

	parts := strings.Split(rel, string(filepath.Separator))
	segments := make([]string, 0, len(parts)+1)
	segments = append(segments, rootClean)
	current := rootClean
	for _, p := range parts {
		current = filepath.Join(current, p)
		segments = append(segments, current)
	}
	return segments
}

// SetRootPath changes the tree root to a new directory, preserving no state.
func (tp *TreePanel) SetRootPath(path string) {
	tp.RootPath = path
	tp.expandedSet = make(map[string]bool)

	root := tp.buildNode(path, filepath.Base(path))
	tp.expandNode(root, path)
	root.SetExpanded(true)
	tp.expandedSet[path] = true

	tp.TreeView.SetRoot(root)
	tp.TreeView.SetCurrentNode(root)
	tp.TreeView.SetTitle(path)
}

// NavigateToPath expands the tree to show the given path.
// If the path is under the current root, it expands in-place.
// If outside the root, it changes root to "/" and then expands.
func (tp *TreePanel) NavigateToPath(path string) {
	path = filepath.Clean(path)
	rootClean := filepath.Clean(tp.RootPath)

	if strings.HasPrefix(path, rootClean) {
		tp.ExpandToPath(path)
	} else {
		// Target is outside current root — switch root to /
		tp.SetRootPath("/")
		tp.ExpandToPath(path)
	}
	tp.TreeView.SetTitle(tp.RootPath)
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

// SelectedIsDir returns true if the currently selected node is a directory.
func (tp *TreePanel) SelectedIsDir() bool {
	path := tp.SelectedPath()
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// SelectedIsExpanded returns true if the currently selected directory is expanded.
func (tp *TreePanel) SelectedIsExpanded() bool {
	return tp.expandedSet[tp.SelectedPath()]
}

// CollapseSelected collapses the currently selected node (if expanded).
// Returns true if a collapse happened.
func (tp *TreePanel) CollapseSelected() bool {
	node := tp.TreeView.GetCurrentNode()
	if node == nil {
		return false
	}
	ref := node.GetReference()
	if ref == nil {
		return false
	}
	path := ref.(string)
	if tp.expandedSet[path] {
		node.ClearChildren()
		node.SetExpanded(false)
		delete(tp.expandedSet, path)
		return true
	}
	return false
}

// MoveToParent moves the cursor to the parent directory of the currently
// selected node. Returns true if the cursor was moved.
func (tp *TreePanel) MoveToParent() bool {
	path := tp.SelectedPath()
	parent := filepath.Dir(path)
	if parent == path {
		return false // already at root
	}
	// Only navigate if parent is under our root
	rootClean := filepath.Clean(tp.RootPath)
	if !strings.HasPrefix(parent, rootClean) {
		return false
	}
	tp.ExpandToPath(parent)
	return true
}

// ToggleHidden flips hidden file visibility and refreshes.
func (tp *TreePanel) ToggleHidden() {
	tp.ShowHidden = !tp.ShowHidden
	tp.Refresh()
}

// ExpandedPaths returns a copy of the set of expanded directory paths.
func (tp *TreePanel) ExpandedPaths() map[string]bool {
	result := make(map[string]bool, len(tp.expandedSet))
	for k, v := range tp.expandedSet {
		result[k] = v
	}
	return result
}

// Refresh reloads the tree preserving the expanded set.
func (tp *TreePanel) Refresh() {
	currentPath := tp.SelectedPath()
	savedExpanded := make(map[string]bool)
	for k, v := range tp.expandedSet {
		savedExpanded[k] = v
	}

	// Rebuild from root
	tp.expandedSet = make(map[string]bool)
	root := tp.buildNode(tp.RootPath, filepath.Base(tp.RootPath))
	tp.rebuildExpanded(root, tp.RootPath, savedExpanded)
	root.SetExpanded(true)
	tp.expandedSet[tp.RootPath] = true

	tp.TreeView.SetRoot(root)
	tp.TreeView.SetTitle(tp.RootPath)

	// Restore cursor position
	tp.ExpandToPath(currentPath)
}

// rebuildExpanded recursively re-expands previously expanded directories.
func (tp *TreePanel) rebuildExpanded(node *tview.TreeNode, path string, saved map[string]bool) {
	if !saved[path] {
		return
	}
	tp.expandNode(node, path)
	node.SetExpanded(true)
	tp.expandedSet[path] = true

	for _, child := range node.GetChildren() {
		ref := child.GetReference()
		if ref == nil {
			continue
		}
		childPath := ref.(string)
		if saved[childPath] {
			tp.rebuildExpanded(child, childPath, saved)
		}
	}
}
