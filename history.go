package main

// History provides browser-style back/forward directory navigation.
// It is a pure data model with no UI dependencies.
type History struct {
	stack []string
	fwd   []string
	max   int
}

// NewHistory creates a History that retains at most maxDepth entries.
func NewHistory(maxDepth int) *History {
	return &History{
		stack: make([]string, 0, maxDepth),
		fwd:   make([]string, 0, maxDepth),
		max:   maxDepth,
	}
}

// Push records a directory visit. It clears the forward stack and suppresses
// consecutive duplicates. If the stack exceeds max depth the oldest entry is
// dropped.
func (h *History) Push(dir string) {
	// Duplicate suppression: ignore if dir equals the top of the stack.
	if len(h.stack) > 0 && h.stack[len(h.stack)-1] == dir {
		return
	}

	h.stack = append(h.stack, dir)

	// Drop the oldest entry when we exceed the limit.
	if len(h.stack) > h.max {
		h.stack = h.stack[1:]
	}

	// Any new push invalidates forward history.
	h.fwd = h.fwd[:0]
}

// Back pops the most recent entry from the history stack, pushes currentDir
// onto the forward stack, and returns the popped directory. It returns
// ("", false) when the stack is empty.
func (h *History) Back(currentDir string) (string, bool) {
	if len(h.stack) == 0 {
		return "", false
	}

	// Pop from stack.
	dir := h.stack[len(h.stack)-1]
	h.stack = h.stack[:len(h.stack)-1]

	// Save current position for Forward.
	h.fwd = append(h.fwd, currentDir)

	return dir, true
}

// Forward pops the most recent entry from the forward stack, pushes currentDir
// onto the history stack, and returns the popped directory. It returns
// ("", false) when the forward stack is empty.
func (h *History) Forward(currentDir string) (string, bool) {
	if len(h.fwd) == 0 {
		return "", false
	}

	// Pop from forward stack.
	dir := h.fwd[len(h.fwd)-1]
	h.fwd = h.fwd[:len(h.fwd)-1]

	// Record current position in the back stack.
	h.stack = append(h.stack, currentDir)

	return dir, true
}

// CanGoBack reports whether there are entries in the back stack.
func (h *History) CanGoBack() bool {
	return len(h.stack) > 0
}

// CanGoForward reports whether there are entries in the forward stack.
func (h *History) CanGoForward() bool {
	return len(h.fwd) > 0
}

// Recent returns the most recently visited directories (newest first),
// deduplicated, up to max entries.
func (h *History) Recent(max int) []string {
	seen := make(map[string]bool)
	var result []string
	// Walk the stack from newest to oldest
	for i := len(h.stack) - 1; i >= 0; i-- {
		dir := h.stack[i]
		if !seen[dir] {
			seen[dir] = true
			result = append(result, dir)
			if len(result) >= max {
				break
			}
		}
	}
	return result
}
