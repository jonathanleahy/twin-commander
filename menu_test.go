package main

import "testing"

// testMenuBar creates a MenuBar with the standard 6 menus for testing.
func testMenuBar() *MenuBar {
	menus := []Menu{
		{Title: "File", Hotkey: 'f'},
		{Title: "View", Hotkey: 'v'},
		{Title: "Search", Hotkey: 's'},
		{Title: "Go", Hotkey: 'g'},
		{Title: "Tools", Hotkey: 't'},
		{Title: "Options", Hotkey: 'o'},
	}
	return NewMenuBar(menus)
}

// TestMenuIndexAtX verifies click-target boundaries for all 6 menus.
// Each menu item is rendered as " Title " — width = len(Title) + 2.
func TestMenuIndexAtX(t *testing.T) {
	mb := testMenuBar()

	// Expected layout (0-indexed x positions):
	//   " File "   = x 0..5  (len 6)
	//   " View "   = x 6..11 (len 6)
	//   " Search " = x 12..19 (len 8)
	//   " Go "     = x 20..23 (len 4)
	//   " Tools "  = x 24..30 (len 7)  [actually len("Tools")+2 = 7]
	//   " Options " = x 31..39 (len 9)
	tests := []struct {
		x        int
		expected int
	}{
		// File: " File " spans x=0..5
		{0, 0}, {3, 0}, {5, 0},
		// View: " View " spans x=6..11
		{6, 1}, {8, 1}, {11, 1},
		// Search: " Search " spans x=12..19
		{12, 2}, {15, 2}, {19, 2},
		// Go: " Go " spans x=20..23
		{20, 3}, {22, 3}, {23, 3},
		// Tools: " Tools " spans x=24..30
		{24, 4}, {27, 4}, {30, 4},
		// Options: " Options " spans x=31..39
		{31, 5}, {35, 5}, {39, 5},
		// Past the end
		{40, -1}, {100, -1},
	}

	for _, tc := range tests {
		got := mb.MenuIndexAtX(tc.x)
		if got != tc.expected {
			t.Errorf("MenuIndexAtX(%d) = %d, want %d", tc.x, got, tc.expected)
		}
	}
}

// TestDropdownOffset verifies the horizontal offset for each menu's dropdown.
func TestDropdownOffset(t *testing.T) {
	mb := testMenuBar()

	// Cumulative widths: File=6, View=6, Search=8, Go=4, Tools=7, Options=9
	tests := []struct {
		activeMenu int
		expected   int
	}{
		{0, 0},                          // File: no preceding menus
		{1, 6},                          // View: after File (6)
		{2, 12},                         // Search: after File+View (6+6)
		{3, 20},                         // Go: after File+View+Search (6+6+8)
		{4, 24},                         // Tools: after ...+Go (20+4)
		{5, 31},                         // Options: after ...+Tools (24+7)
	}

	for _, tc := range tests {
		mb.ActiveMenu = tc.activeMenu
		got := mb.DropdownOffset()
		if got != tc.expected {
			t.Errorf("DropdownOffset() with ActiveMenu=%d = %d, want %d",
				tc.activeMenu, got, tc.expected)
		}
	}
}

// TestMenuIndexAtX_ConsistentWithDropdownOffset ensures that clicking at the
// dropdown offset position returns the correct menu index.
func TestMenuIndexAtX_ConsistentWithDropdownOffset(t *testing.T) {
	mb := testMenuBar()

	for i := range mb.Menus {
		mb.ActiveMenu = i
		offset := mb.DropdownOffset()
		got := mb.MenuIndexAtX(offset)
		if got != i {
			t.Errorf("MenuIndexAtX(DropdownOffset(%d)=%d) = %d, want %d",
				i, offset, got, i)
		}
	}
}

// TestMenuForHotkey verifies hotkey lookup.
func TestMenuForHotkey(t *testing.T) {
	mb := testMenuBar()

	tests := []struct {
		r        rune
		expected int
	}{
		{'f', 0}, {'F', 0},
		{'v', 1}, {'V', 1},
		{'s', 2},
		{'g', 3},
		{'t', 4},
		{'o', 5},
		{'x', -1}, {'z', -1},
	}

	for _, tc := range tests {
		got := mb.MenuForHotkey(tc.r)
		if got != tc.expected {
			t.Errorf("MenuForHotkey(%q) = %d, want %d", tc.r, got, tc.expected)
		}
	}
}
