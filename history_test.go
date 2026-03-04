package main

import (
	"testing"
)

func TestHistory_NewIsEmpty(t *testing.T) {
	h := NewHistory(100)
	if h.CanGoBack() {
		t.Error("new history should not be able to go back")
	}
	if h.CanGoForward() {
		t.Error("new history should not be able to go forward")
	}
}

func TestHistory_PushAndBack(t *testing.T) {
	h := NewHistory(100)
	h.Push("/a")
	h.Push("/b")

	// Going back from /c should return /b.
	dir, ok := h.Back("/c")
	if !ok || dir != "/b" {
		t.Errorf("expected /b, got %q (ok=%v)", dir, ok)
	}

	// Going back again should return /a.
	dir, ok = h.Back(dir)
	if !ok || dir != "/a" {
		t.Errorf("expected /a, got %q (ok=%v)", dir, ok)
	}

	// Stack should now be empty.
	if h.CanGoBack() {
		t.Error("stack should be empty after popping all entries")
	}
}

func TestHistory_BackAndForward(t *testing.T) {
	h := NewHistory(100)
	h.Push("/a")
	h.Push("/b")

	// Go back from /c -> returns /b, forward stack gets /c.
	dir, _ := h.Back("/c")
	if dir != "/b" {
		t.Fatalf("expected /b, got %q", dir)
	}

	// Go forward from /b -> returns /c.
	dir, ok := h.Forward("/b")
	if !ok || dir != "/c" {
		t.Errorf("expected /c, got %q (ok=%v)", dir, ok)
	}

	// After forward, back stack should contain /a and /b.
	if !h.CanGoBack() {
		t.Error("should be able to go back after forward")
	}
}

func TestHistory_PushClearsForward(t *testing.T) {
	h := NewHistory(100)
	h.Push("/a")
	h.Push("/b")

	// Go back to build a forward stack.
	h.Back("/c")
	if !h.CanGoForward() {
		t.Fatal("forward stack should not be empty after Back")
	}

	// A new push should clear the forward stack.
	h.Push("/d")
	if h.CanGoForward() {
		t.Error("forward stack should be cleared after Push")
	}
}

func TestHistory_DuplicateSuppression(t *testing.T) {
	h := NewHistory(100)
	h.Push("/a")
	h.Push("/a") // duplicate — should be suppressed

	// Only one entry should be on the stack, so one Back succeeds
	// and the next fails.
	dir, ok := h.Back("/x")
	if !ok || dir != "/a" {
		t.Errorf("expected /a, got %q (ok=%v)", dir, ok)
	}
	if h.CanGoBack() {
		t.Error("duplicate push should not have added a second entry")
	}
}

func TestHistory_MaxDepth(t *testing.T) {
	h := NewHistory(3)
	h.Push("/a")
	h.Push("/b")
	h.Push("/c")
	h.Push("/d") // /a should be dropped

	// Pop all three remaining entries.
	dir, _ := h.Back("/x")
	if dir != "/d" {
		t.Errorf("expected /d, got %q", dir)
	}
	dir, _ = h.Back(dir)
	if dir != "/c" {
		t.Errorf("expected /c, got %q", dir)
	}
	dir, _ = h.Back(dir)
	if dir != "/b" {
		t.Errorf("expected /b, got %q", dir)
	}

	// /a was dropped — stack must be empty now.
	if h.CanGoBack() {
		t.Error("oldest entry should have been dropped at max depth")
	}
}

func TestHistory_BackAtEmpty(t *testing.T) {
	h := NewHistory(100)
	dir, ok := h.Back("/current")
	if ok {
		t.Error("Back on empty history should return false")
	}
	if dir != "" {
		t.Errorf("Back on empty history should return empty string, got %q", dir)
	}
}

func TestHistory_ForwardAtEmpty(t *testing.T) {
	h := NewHistory(100)
	dir, ok := h.Forward("/current")
	if ok {
		t.Error("Forward on empty history should return false")
	}
	if dir != "" {
		t.Errorf("Forward on empty history should return empty string, got %q", dir)
	}
}

func TestHistory_Recent(t *testing.T) {
	h := NewHistory(100)
	h.Push("/a")
	h.Push("/b")
	h.Push("/c")
	h.Push("/a") // duplicate
	h.Push("/d")

	recent := h.Recent(10)
	if len(recent) != 4 {
		t.Fatalf("expected 4 recent dirs, got %d: %v", len(recent), recent)
	}
	// Newest first, deduplicated
	expected := []string{"/d", "/a", "/c", "/b"}
	for i, exp := range expected {
		if recent[i] != exp {
			t.Errorf("recent[%d] = %q, want %q", i, recent[i], exp)
		}
	}
}

func TestHistory_RecentMax(t *testing.T) {
	h := NewHistory(100)
	for i := 0; i < 30; i++ {
		h.Push(string(rune('a' + i)))
	}
	recent := h.Recent(5)
	if len(recent) != 5 {
		t.Fatalf("expected 5 recent dirs, got %d", len(recent))
	}
}

func TestHistory_RecentEmpty(t *testing.T) {
	h := NewHistory(100)
	recent := h.Recent(10)
	if len(recent) != 0 {
		t.Fatalf("expected 0 recent dirs, got %d", len(recent))
	}
}
