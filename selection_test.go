package main

import (
	"testing"
)

// testEntries returns a standard set of FileEntry values and base dir
// used across selection tests.
func testEntries() ([]FileEntry, string) {
	entries := []FileEntry{
		{Name: "..", IsDir: true},
		{Name: "alpha.go", Accessible: true},
		{Name: "beta.txt", Accessible: true},
		{Name: "gamma.go", Accessible: true},
		{Name: "delta.md", Accessible: true},
	}
	baseDir := "/home/user/project"
	return entries, baseDir
}

func TestSelection_NewIsEmpty(t *testing.T) {
	s := NewSelection()
	if s.Count() != 0 {
		t.Fatalf("expected count 0, got %d", s.Count())
	}
	if s.IsSelected("/any/path") {
		t.Fatal("nothing should be selected in a new Selection")
	}
}

func TestSelection_Toggle(t *testing.T) {
	s := NewSelection()

	s.Toggle("/home/user/file.txt")
	if !s.IsSelected("/home/user/file.txt") {
		t.Fatal("expected file.txt to be selected after first toggle")
	}
	if s.Count() != 1 {
		t.Fatalf("expected count 1, got %d", s.Count())
	}

	s.Toggle("/home/user/file.txt")
	if s.IsSelected("/home/user/file.txt") {
		t.Fatal("expected file.txt to be deselected after second toggle")
	}
	if s.Count() != 0 {
		t.Fatalf("expected count 0, got %d", s.Count())
	}
}

func TestSelection_MultipleItems(t *testing.T) {
	s := NewSelection()

	s.Toggle("/a")
	s.Toggle("/b")
	s.Toggle("/c")

	if s.Count() != 3 {
		t.Fatalf("expected count 3, got %d", s.Count())
	}

	paths := s.Paths()
	expected := []string{"/a", "/b", "/c"}
	if len(paths) != len(expected) {
		t.Fatalf("expected %d paths, got %d", len(expected), len(paths))
	}
	for i, p := range paths {
		if p != expected[i] {
			t.Errorf("paths[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestSelection_Clear(t *testing.T) {
	s := NewSelection()
	s.Toggle("/a")
	s.Toggle("/b")
	s.Clear()

	if s.Count() != 0 {
		t.Fatalf("expected count 0 after clear, got %d", s.Count())
	}
	if s.IsSelected("/a") {
		t.Fatal("/a should not be selected after clear")
	}
}

func TestSelection_Set(t *testing.T) {
	s := NewSelection()

	s.Set("/x", true)
	if !s.IsSelected("/x") {
		t.Fatal("expected /x to be selected after Set(true)")
	}

	s.Set("/x", false)
	if s.IsSelected("/x") {
		t.Fatal("expected /x to be deselected after Set(false)")
	}
	if s.Count() != 0 {
		t.Fatalf("expected count 0, got %d", s.Count())
	}

	// Set false on something that was never selected — should be a no-op.
	s.Set("/never", false)
	if s.Count() != 0 {
		t.Fatalf("expected count 0 after Set(false) on unknown, got %d", s.Count())
	}
}

func TestSelection_InvertFromEntries(t *testing.T) {
	entries, baseDir := testEntries()
	s := NewSelection()

	// Pre-select alpha.go and gamma.go
	s.Set("/home/user/project/alpha.go", true)
	s.Set("/home/user/project/gamma.go", true)

	s.InvertFromEntries(entries, baseDir)

	// alpha.go and gamma.go should now be deselected;
	// beta.txt and delta.md should now be selected.
	// ".." should never appear.
	if s.IsSelected("/home/user/project/alpha.go") {
		t.Error("alpha.go should be deselected after invert")
	}
	if s.IsSelected("/home/user/project/gamma.go") {
		t.Error("gamma.go should be deselected after invert")
	}
	if !s.IsSelected("/home/user/project/beta.txt") {
		t.Error("beta.txt should be selected after invert")
	}
	if !s.IsSelected("/home/user/project/delta.md") {
		t.Error("delta.md should be selected after invert")
	}
	if s.IsSelected("/home/user/project/..") {
		t.Error(".. should never be selected")
	}
	if s.Count() != 2 {
		t.Fatalf("expected count 2, got %d", s.Count())
	}
}

func TestSelection_StartVisual(t *testing.T) {
	s := NewSelection()

	if s.IsVisual() {
		t.Fatal("should not be in visual mode initially")
	}

	s.StartVisual(2)
	if !s.IsVisual() {
		t.Fatal("should be in visual mode after StartVisual")
	}
}

func TestSelection_UpdateVisual(t *testing.T) {
	entries, baseDir := testEntries()
	s := NewSelection()

	// Visual from index 1 (alpha.go) to index 3 (gamma.go)
	s.StartVisual(1)
	s.UpdateVisual(3, entries, baseDir)

	if s.Count() != 3 {
		t.Fatalf("expected 3 selected, got %d", s.Count())
	}
	if !s.IsSelected("/home/user/project/alpha.go") {
		t.Error("alpha.go should be selected")
	}
	if !s.IsSelected("/home/user/project/beta.txt") {
		t.Error("beta.txt should be selected")
	}
	if !s.IsSelected("/home/user/project/gamma.go") {
		t.Error("gamma.go should be selected")
	}

	// Update visual to a narrower range: just index 1 (alpha.go)
	s.UpdateVisual(1, entries, baseDir)
	if s.Count() != 1 {
		t.Fatalf("expected 1 selected after narrowing, got %d", s.Count())
	}
	if !s.IsSelected("/home/user/project/alpha.go") {
		t.Error("alpha.go should be selected")
	}

	// Reverse direction: anchor=1, current=0 — index 0 is ".." so only alpha.go
	s.UpdateVisual(0, entries, baseDir)
	if s.Count() != 1 {
		t.Fatalf("expected 1 selected (reverse, skipping ..), got %d", s.Count())
	}
	if !s.IsSelected("/home/user/project/alpha.go") {
		t.Error("alpha.go should be selected in reverse range")
	}
}

func TestSelection_EndVisual(t *testing.T) {
	entries, baseDir := testEntries()
	s := NewSelection()

	s.StartVisual(1)
	s.UpdateVisual(2, entries, baseDir)
	s.EndVisual()

	if s.IsVisual() {
		t.Fatal("should not be in visual mode after EndVisual")
	}
	// Selections must be preserved.
	if s.Count() != 2 {
		t.Fatalf("expected 2 selected after EndVisual, got %d", s.Count())
	}
	if !s.IsSelected("/home/user/project/alpha.go") {
		t.Error("alpha.go should still be selected")
	}
	if !s.IsSelected("/home/user/project/beta.txt") {
		t.Error("beta.txt should still be selected")
	}
}

func TestSelection_MatchPattern(t *testing.T) {
	entries, baseDir := testEntries()
	s := NewSelection()

	s.MatchPattern(".go", entries, baseDir)

	if s.Count() != 2 {
		t.Fatalf("expected 2 matches for '.go', got %d", s.Count())
	}
	if !s.IsSelected("/home/user/project/alpha.go") {
		t.Error("alpha.go should match '.go'")
	}
	if !s.IsSelected("/home/user/project/gamma.go") {
		t.Error("gamma.go should match '.go'")
	}

	// Case-insensitive match
	s.Clear()
	s.MatchPattern(".GO", entries, baseDir)
	if s.Count() != 2 {
		t.Fatalf("expected 2 matches for '.GO' (case-insensitive), got %d", s.Count())
	}

	// Pattern that matches nothing
	s.Clear()
	s.MatchPattern("zzz", entries, baseDir)
	if s.Count() != 0 {
		t.Fatalf("expected 0 matches for 'zzz', got %d", s.Count())
	}

	// ".." should never be selected
	s.Clear()
	s.MatchPattern("..", entries, baseDir)
	if s.IsSelected("/home/user/project/..") {
		t.Error(".. should never be selected by MatchPattern")
	}
}

func TestSelection_PathsAfterToggle(t *testing.T) {
	s := NewSelection()

	s.Toggle("/z")
	s.Toggle("/a")
	s.Toggle("/m")
	// Toggle /a off
	s.Toggle("/a")

	paths := s.Paths()
	expected := []string{"/m", "/z"}
	if len(paths) != len(expected) {
		t.Fatalf("expected %d paths, got %d: %v", len(expected), len(paths), paths)
	}
	for i, p := range paths {
		if p != expected[i] {
			t.Errorf("paths[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestSelection_PathsSorted(t *testing.T) {
	s := NewSelection()

	s.Set("/zebra", true)
	s.Set("/apple", true)
	s.Set("/mango", true)
	s.Set("/banana", true)

	paths := s.Paths()
	expected := []string{"/apple", "/banana", "/mango", "/zebra"}
	if len(paths) != len(expected) {
		t.Fatalf("expected %d paths, got %d", len(expected), len(paths))
	}
	for i, p := range paths {
		if p != expected[i] {
			t.Errorf("paths[%d] = %q, want %q", i, p, expected[i])
		}
	}
}
