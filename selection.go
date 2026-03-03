package main

import (
	"path/filepath"
	"sort"
	"strings"
)

// Selection tracks multi-file selection state with support for
// visual (range) mode and pattern matching. It is a pure data model
// with no UI dependencies.
type Selection struct {
	items     map[string]bool
	anchorIdx int
	visual    bool
}

// NewSelection returns an initialised, empty Selection.
func NewSelection() *Selection {
	return &Selection{
		items: make(map[string]bool),
	}
}

// Toggle flips the selection state for path.
func (s *Selection) Toggle(path string) {
	if s.items[path] {
		delete(s.items, path)
	} else {
		s.items[path] = true
	}
}

// Set explicitly selects or deselects path.
func (s *Selection) Set(path string, selected bool) {
	if selected {
		s.items[path] = true
	} else {
		delete(s.items, path)
	}
}

// Clear removes all selections.
func (s *Selection) Clear() {
	s.items = make(map[string]bool)
}

// IsSelected reports whether path is currently selected.
func (s *Selection) IsSelected(path string) bool {
	return s.items[path]
}

// Paths returns all selected paths in sorted order.
func (s *Selection) Paths() []string {
	paths := make([]string, 0, len(s.items))
	for p := range s.items {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}

// Count returns the number of selected items.
func (s *Selection) Count() int {
	return len(s.items)
}

// InvertFromEntries inverts the selection against the given entry list.
// Entries whose Name is ".." are skipped.
func (s *Selection) InvertFromEntries(entries []FileEntry, baseDir string) {
	for _, e := range entries {
		if e.Name == ".." {
			continue
		}
		p := filepath.Join(baseDir, e.Name)
		if s.items[p] {
			delete(s.items, p)
		} else {
			s.items[p] = true
		}
	}
}

// StartVisual enters visual (range) selection mode and records the
// anchor index.
func (s *Selection) StartVisual(anchorIdx int) {
	s.visual = true
	s.anchorIdx = anchorIdx
}

// UpdateVisual selects every entry between the anchor and currentIdx
// (inclusive). Previous selections are cleared first so that the range
// tracks the cursor exactly. Entries whose Name is ".." are skipped.
func (s *Selection) UpdateVisual(currentIdx int, entries []FileEntry, baseDir string) {
	s.items = make(map[string]bool)

	lo, hi := s.anchorIdx, currentIdx
	if lo > hi {
		lo, hi = hi, lo
	}

	for i := lo; i <= hi && i < len(entries); i++ {
		if entries[i].Name == ".." {
			continue
		}
		s.items[filepath.Join(baseDir, entries[i].Name)] = true
	}
}

// EndVisual exits visual mode while keeping the current selections.
func (s *Selection) EndVisual() {
	s.visual = false
}

// IsVisual reports whether visual mode is active.
func (s *Selection) IsVisual() bool {
	return s.visual
}

// MatchPattern selects every entry whose name contains the given
// substring (case-insensitive). Entries whose Name is ".." are skipped.
func (s *Selection) MatchPattern(pattern string, entries []FileEntry, baseDir string) {
	lowerPattern := strings.ToLower(pattern)
	for _, e := range entries {
		if e.Name == ".." {
			continue
		}
		if strings.Contains(strings.ToLower(e.Name), lowerPattern) {
			s.items[filepath.Join(baseDir, e.Name)] = true
		}
	}
}
