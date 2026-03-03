package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Helper to create a Panel with a real tview.Table for testing
func newTestPanel(path string) *Panel {
	table := tview.NewTable()
	table.SetBorder(true)
	table.SetSelectable(true, false)
	table.SetSelectedStyle(tcell.StyleDefault.Reverse(true))

	return &Panel{
		Path:       path,
		Table:      table,
		ShowHidden: false,
		Filter:     "",
	}
}

// Helper to set up a standard test directory
func setupTestDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()

	os.Mkdir(filepath.Join(tmp, "docs"), 0755)
	os.Mkdir(filepath.Join(tmp, "src"), 0755)
	os.WriteFile(filepath.Join(tmp, "README.md"), []byte("# README"), 0644)
	os.WriteFile(filepath.Join(tmp, "config.json"), []byte(`{"key":"val"}`), 0644)
	os.Mkdir(filepath.Join(tmp, ".git"), 0755)
	os.WriteFile(filepath.Join(tmp, ".gitignore"), []byte("*.log"), 0644)

	return tmp
}

// Test LoadDir populates entries correctly
func TestPanel_LoadDir(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	// Should have: .., docs, src, README.md, config.json (hidden files off)
	if len(p.Entries) != 5 {
		t.Errorf("expected 5 entries (including ..), got %d", len(p.Entries))
		for _, e := range p.Entries {
			t.Logf("  %q", e.Name)
		}
	}
	// First entry should be ..
	if p.Entries[0].Name != ".." {
		t.Errorf("first entry should be '..', got %q", p.Entries[0].Name)
	}
	// Directories should come before files (after ..)
	if !p.Entries[1].IsDir || !p.Entries[2].IsDir {
		t.Error("entries[1] and entries[2] should be directories")
	}
}

// Test LoadDir sort order: .., dirs alphabetical, files alphabetical
func TestPanel_LoadDir_SortOrder(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	expected := []string{"..", "docs", "src", "config.json", "README.md"}
	if len(p.Entries) != len(expected) {
		t.Fatalf("expected %d entries, got %d", len(expected), len(p.Entries))
	}
	for i, name := range expected {
		if p.Entries[i].Name != name {
			t.Errorf("entry[%d] = %q, want %q", i, p.Entries[i].Name, name)
		}
	}
}

// Test LoadDir at root has no .. entry (EC-2)
func TestPanel_LoadDir_AtRoot(t *testing.T) {
	p := newTestPanel("/")
	p.LoadDir()

	for _, e := range p.Entries {
		if e.Name == ".." {
			t.Error("root directory should not have '..' entry")
		}
	}
	if len(p.Entries) == 0 {
		t.Error("root directory should have at least some entries")
	}
}

// Test NavigateInto changes directory (TS-10)
func TestPanel_NavigateInto(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	p.NavigateInto("docs")

	expectedPath := filepath.Join(tmp, "docs")
	if p.Path != expectedPath {
		t.Errorf("Path = %q, want %q", p.Path, expectedPath)
	}
	// First entry should be ..
	if len(p.Entries) == 0 || p.Entries[0].Name != ".." {
		t.Error("after NavigateInto, first entry should be '..'")
	}
	// Cursor should be at row 0
	row, _ := p.Table.GetSelection()
	if row != 0 {
		t.Errorf("cursor should be at row 0 after NavigateInto, got %d", row)
	}
}

// Test NavigateInto clears filter (TS-30)
func TestPanel_NavigateInto_ClearsFilter(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.Filter = "something"
	p.LoadDir()

	p.NavigateInto("docs")

	if p.Filter != "" {
		t.Errorf("NavigateInto should clear filter, got %q", p.Filter)
	}
}

// Test NavigateUp returns to parent and reports previous dir name (TS-13, TS-14)
func TestPanel_NavigateUp(t *testing.T) {
	tmp := setupTestDir(t)
	docsPath := filepath.Join(tmp, "docs")
	p := newTestPanel(docsPath)
	p.LoadDir()

	prevName := p.NavigateUp()

	if prevName != "docs" {
		t.Errorf("NavigateUp should return 'docs', got %q", prevName)
	}
	if p.Path != tmp {
		t.Errorf("Path should be %q after NavigateUp, got %q", tmp, p.Path)
	}
}

// Test NavigateUp cursor positioning on previous directory (TS-14)
func TestPanel_NavigateUp_CursorPosition(t *testing.T) {
	tmp := setupTestDir(t)
	srcPath := filepath.Join(tmp, "src")
	p := newTestPanel(srcPath)
	p.LoadDir()

	p.NavigateUp()

	// Cursor should be on "src" entry
	row, _ := p.Table.GetSelection()
	if row >= 0 && row < len(p.Entries) {
		if p.Entries[row].Name != "src" {
			t.Errorf("cursor should be on 'src', got %q at row %d", p.Entries[row].Name, row)
		}
	} else {
		t.Errorf("cursor row %d out of range", row)
	}
}

// Test NavigateUp clears filter (TS-30)
func TestPanel_NavigateUp_ClearsFilter(t *testing.T) {
	tmp := setupTestDir(t)
	docsPath := filepath.Join(tmp, "docs")
	p := newTestPanel(docsPath)
	p.Filter = "active"
	p.LoadDir()

	p.NavigateUp()

	if p.Filter != "" {
		t.Errorf("NavigateUp should clear filter, got %q", p.Filter)
	}
}

// Test NavigateUp at root is a no-op (EC-9)
func TestPanel_NavigateUp_AtRoot(t *testing.T) {
	p := newTestPanel("/")
	p.LoadDir()

	prevName := p.NavigateUp()

	if p.Path != "/" {
		t.Errorf("NavigateUp at root should stay at /, got %q", p.Path)
	}
	if prevName != "" {
		t.Errorf("NavigateUp at root should return empty string, got %q", prevName)
	}
}

// Test ToggleHidden shows/hides dotfiles (TS-23, TS-24)
func TestPanel_ToggleHidden(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	initialCount := len(p.Entries)

	// Toggle hidden on
	p.ToggleHidden()
	if !p.ShowHidden {
		t.Error("ShowHidden should be true after toggle")
	}
	// Should now have more entries (dotfiles visible)
	if len(p.Entries) <= initialCount {
		t.Errorf("expected more entries with hidden visible, got %d (was %d)", len(p.Entries), initialCount)
	}

	// Toggle hidden off
	p.ToggleHidden()
	if p.ShowHidden {
		t.Error("ShowHidden should be false after second toggle")
	}
	if len(p.Entries) != initialCount {
		t.Errorf("expected %d entries after toggle off, got %d", initialCount, len(p.Entries))
	}
}

// Test SetFilter narrows entries (TS-26)
func TestPanel_SetFilter(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	p.SetFilter("md")

	// Should match: .., README.md
	found := false
	for _, e := range p.Entries {
		if e.Name == "README.md" {
			found = true
		}
	}
	if !found {
		t.Error("filter 'md' should include README.md")
	}
	// .. should still be present
	if p.Entries[0].Name != ".." {
		t.Error(".. should always be present during filtering")
	}
}

// Test SetFilter updates status bar count (TS-26)
func TestPanel_SetFilter_StatusBarUpdates(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	p.SetFilter("md")

	text := p.StatusText()
	// Only README.md matches, so 1 item
	if !strings.Contains(text, "1 items") {
		t.Errorf("filtered status should show '1 items', got %q", text)
	}
}

// Test ClearFilter restores full listing (TS-27)
func TestPanel_ClearFilter(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()
	fullCount := len(p.Entries)

	p.SetFilter("md")
	filteredCount := len(p.Entries)
	if filteredCount >= fullCount {
		t.Error("filter should reduce entry count")
	}

	p.ClearFilter()
	if p.Filter != "" {
		t.Error("ClearFilter should set Filter to empty string")
	}
	if len(p.Entries) != fullCount {
		t.Errorf("after ClearFilter, expected %d entries, got %d", fullCount, len(p.Entries))
	}
}

// Test Refresh reloads directory (TS-33)
func TestPanel_Refresh(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()
	initialCount := len(p.Entries)

	// Add a new file externally
	os.WriteFile(filepath.Join(tmp, "newfile.txt"), []byte("new"), 0644)

	p.Refresh()

	if len(p.Entries) != initialCount+1 {
		t.Errorf("after Refresh, expected %d entries, got %d", initialCount+1, len(p.Entries))
	}
}

// Test Refresh preserves cursor on existing entry (TS-33)
func TestPanel_Refresh_PreservesCursor(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	// Select "README.md"
	for i, e := range p.Entries {
		if e.Name == "README.md" {
			p.Table.Select(i, 0)
			break
		}
	}

	p.Refresh()

	row, _ := p.Table.GetSelection()
	if row >= 0 && row < len(p.Entries) && p.Entries[row].Name != "README.md" {
		t.Errorf("after Refresh, cursor should be on README.md, got %q (row %d)", p.Entries[row].Name, row)
	}
}

// Test Refresh when selected entry disappears (EC-19)
func TestPanel_Refresh_EntryDisappears(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmp, "b.txt"), []byte("b"), 0644)
	os.WriteFile(filepath.Join(tmp, "c.txt"), []byte("c"), 0644)

	p := newTestPanel(tmp)
	p.LoadDir()

	// Select b.txt
	for i, e := range p.Entries {
		if e.Name == "b.txt" {
			p.Table.Select(i, 0)
			break
		}
	}

	// Delete b.txt externally
	os.Remove(filepath.Join(tmp, "b.txt"))
	p.Refresh()

	// Cursor should be on a valid row (no panic)
	row, _ := p.Table.GetSelection()
	if row < 0 || row >= len(p.Entries) {
		t.Errorf("cursor row %d out of range after entry deleted", row)
	}
}

// Test Refresh re-applies active filter (FR-16)
func TestPanel_Refresh_ReappliesFilter(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "readme.md"), []byte("r"), 0644)
	os.WriteFile(filepath.Join(tmp, "config.json"), []byte("c"), 0644)

	p := newTestPanel(tmp)
	p.SetFilter("md")

	// Only .., readme.md visible
	if len(p.Entries) != 2 {
		t.Errorf("expected 2 filtered entries, got %d", len(p.Entries))
	}

	// Add another .md file externally
	os.WriteFile(filepath.Join(tmp, "notes.md"), []byte("n"), 0644)
	p.Refresh()

	// Should now have .., readme.md, notes.md
	if len(p.Entries) != 3 {
		t.Errorf("after refresh, expected 3 filtered entries, got %d", len(p.Entries))
	}
}

// Test StatusText format (TS-16)
func TestPanel_StatusText(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	text := p.StatusText()
	if text == "" {
		t.Error("StatusText should not be empty")
	}
	// Should be "4 items, SIZE" (docs, src, README.md, config.json — excludes ..)
	if !strings.Contains(text, "4 items") {
		t.Errorf("StatusText should contain '4 items', got %q", text)
	}
}

// Test StatusText with hidden files indicator (TS-42)
func TestPanel_StatusText_HiddenIndicator(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.ShowHidden = true
	p.LoadDir()

	text := p.StatusText()
	if !strings.Contains(text, "[H]") {
		t.Errorf("StatusText with ShowHidden should contain '[H]', got %q", text)
	}
	// Should count 6 items (docs, src, README.md, config.json, .git, .gitignore)
	if !strings.Contains(text, "6 items") {
		t.Errorf("StatusText with hidden should show '6 items', got %q", text)
	}
}

// Test StatusText without hidden indicator when off
func TestPanel_StatusText_NoHiddenIndicator(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	text := p.StatusText()
	if strings.Contains(text, "[H]") {
		t.Errorf("StatusText without ShowHidden should not contain '[H]', got %q", text)
	}
}

// Test SetActive changes border color (FR-3)
func TestPanel_SetActive(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)

	p.SetActive(true)
	// No panic = pass (can't easily introspect tview border color)
	p.SetActive(false)
}

// Test title updates on LoadDir (FR-8)
func TestPanel_TitleUpdates(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	title := p.Table.GetTitle()
	if title != tmp {
		t.Errorf("panel title should be %q, got %q", tmp, title)
	}
}

// Test title updates after navigation (TS-15)
func TestPanel_TitleUpdatesOnNavigation(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	p.NavigateInto("docs")

	expectedPath := filepath.Join(tmp, "docs")
	title := p.Table.GetTitle()
	if title != expectedPath {
		t.Errorf("title after NavigateInto should be %q, got %q", expectedPath, title)
	}
}

// Test NavigateInto with inaccessible directory (TS-37, TS-38)
func TestPanel_NavigateInto_Inaccessible(t *testing.T) {
	tmp := setupTestDir(t)
	restricted := filepath.Join(tmp, "restricted")
	os.Mkdir(restricted, 0000)
	defer os.Chmod(restricted, 0755)

	p := newTestPanel(tmp)
	p.LoadDir()

	originalPath := p.Path
	err := p.TryNavigateInto("restricted")
	if err == nil {
		t.Error("TryNavigateInto on inaccessible dir should return error")
	}
	if p.Path != originalPath {
		t.Error("panel should stay in original directory after failed navigation")
	}
}

// Test that ".." entry renders without "/" suffix (FR-2)
func TestPanel_DotDotNoSlashSuffix(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	cell := p.Table.GetCell(0, 0)
	if cell == nil {
		t.Fatal("first cell should exist")
	}
	if !strings.Contains(cell.Text, "..") {
		t.Errorf("first cell text should contain '..' not %q", cell.Text)
	}
	if strings.HasSuffix(cell.Text, "/") {
		t.Errorf("'..' should not end with '/', got %q", cell.Text)
	}
}

// Test ".." has empty size and date columns (FR-2)
func TestPanel_DotDotEmptySizeAndDate(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	sizeCell := p.Table.GetCell(0, 1)
	dateCell := p.Table.GetCell(0, 2)
	if sizeCell == nil || dateCell == nil {
		t.Fatal("size and date cells should exist for ..")
	}
	if sizeCell.Text != "" {
		t.Errorf(".. size should be empty, got %q", sizeCell.Text)
	}
	if dateCell.Text != "" {
		t.Errorf(".. date should be empty, got %q", dateCell.Text)
	}
}

// Test that directories have "/" suffix in rendering (FR-2)
func TestPanel_DirectorySlashSuffix(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	for i, e := range p.Entries {
		if e.Name == "docs" {
			cell := p.Table.GetCell(i, 0)
			if !strings.Contains(cell.Text, "docs/") {
				t.Errorf("directory should contain 'docs/', got %q", cell.Text)
			}
			break
		}
	}
}

// Test regular files render without "/" suffix (FR-2)
func TestPanel_FileNoSlashSuffix(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	for i, e := range p.Entries {
		if e.Name == "README.md" {
			cell := p.Table.GetCell(i, 0)
			if !strings.Contains(cell.Text, "README.md") {
				t.Errorf("file should contain 'README.md', got %q", cell.Text)
			}
			if strings.HasSuffix(cell.Text, "/") {
				t.Errorf("file should not end with '/', got %q", cell.Text)
			}
			break
		}
	}
}

// Test symlink to directory renders with "/" suffix (FR-19)
func TestPanel_SymlinkToDirRendering(t *testing.T) {
	tmp := t.TempDir()
	os.Mkdir(filepath.Join(tmp, "realdir"), 0755)
	os.Symlink(filepath.Join(tmp, "realdir"), filepath.Join(tmp, "linkdir"))

	p := newTestPanel(tmp)
	p.LoadDir()

	for i, e := range p.Entries {
		if e.Name == "linkdir" {
			cell := p.Table.GetCell(i, 0)
			if !strings.Contains(cell.Text, "linkdir/") {
				t.Errorf("symlink to dir should contain 'linkdir/', got %q", cell.Text)
			}
			if !e.IsSymlink {
				t.Error("linkdir should be marked as symlink")
			}
			if !e.IsDir {
				t.Error("linkdir should be marked as dir (target is dir)")
			}
			break
		}
	}
}

// Test broken symlink renders with "---" for size and date (FR-19, EC-11)
func TestPanel_BrokenSymlinkRendering(t *testing.T) {
	tmp := t.TempDir()
	os.Symlink(filepath.Join(tmp, "nonexistent"), filepath.Join(tmp, "broken"))

	p := newTestPanel(tmp)
	p.LoadDir()

	for i, e := range p.Entries {
		if e.Name == "broken" {
			sizeCell := p.Table.GetCell(i, 1)
			dateCell := p.Table.GetCell(i, 2)
			if sizeCell.Text != "---" {
				t.Errorf("broken symlink size should be '---', got %q", sizeCell.Text)
			}
			if dateCell.Text != "---" {
				t.Errorf("broken symlink date should be '---', got %q", dateCell.Text)
			}
			break
		}
	}
}

// Test status bar counts with inaccessible entries (FR-9)
func TestPanel_StatusText_InaccessibleEntries(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "good.txt"), []byte("hello"), 0644)
	os.Symlink(filepath.Join(tmp, "nonexistent"), filepath.Join(tmp, "broken"))

	p := newTestPanel(tmp)
	p.LoadDir()

	text := p.StatusText()
	// Should count 2 items (good.txt + broken), total size = 5 (only good.txt)
	if !strings.Contains(text, "2 items") {
		t.Errorf("status should show '2 items', got %q", text)
	}
}

// Test empty directory (EC-1)
func TestPanel_EmptyDirectory(t *testing.T) {
	tmp := t.TempDir()
	emptyDir := filepath.Join(tmp, "empty")
	os.Mkdir(emptyDir, 0755)

	p := newTestPanel(emptyDir)
	p.LoadDir()

	if len(p.Entries) != 1 {
		t.Errorf("empty dir should have 1 entry (just ..), got %d", len(p.Entries))
	}
	if p.Entries[0].Name != ".." {
		t.Errorf("only entry should be '..', got %q", p.Entries[0].Name)
	}
	if !strings.Contains(p.StatusText(), "0 items") {
		t.Errorf("status should show '0 items', got %q", p.StatusText())
	}
	if !strings.Contains(p.StatusText(), "0 items, 0") {
		t.Errorf("status should show '0 items, 0', got %q", p.StatusText())
	}
}

// Test filter with hidden files interaction (EC-13)
func TestPanel_FilterWithHiddenToggle(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "config.json"), []byte("{}"), 0644)
	os.Mkdir(filepath.Join(tmp, ".config"), 0755)

	p := newTestPanel(tmp)
	p.SetFilter("config")

	// Hidden off: should only see config.json + ..
	countHiddenOff := len(p.Entries)

	// Toggle hidden on
	p.ShowHidden = true
	p.LoadDir()

	// Should now also see .config/ matching "config"
	if len(p.Entries) <= countHiddenOff {
		t.Error("showing hidden files with config filter should reveal .config/")
	}
}

// Test SelectedEntry returns correct entry
func TestPanel_SelectedEntry(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	// Default selection is row 0 (..)
	entry := p.SelectedEntry()
	if entry == nil {
		t.Fatal("SelectedEntry should not be nil")
	}
	if entry.Name != ".." {
		t.Errorf("default selection should be '..', got %q", entry.Name)
	}

	// Select a different row
	p.Table.Select(1, 0)
	entry = p.SelectedEntry()
	if entry == nil {
		t.Fatal("SelectedEntry should not be nil at row 1")
	}
	if entry.Name != "docs" {
		t.Errorf("entry at row 1 should be 'docs', got %q", entry.Name)
	}
}

// Test LoadDir error handling (TS-38 / ERR-2)
func TestPanel_LoadDir_ErrorSetsStatus(t *testing.T) {
	p := newTestPanel("/nonexistent/path/that/does/not/exist")
	p.LoadDir()

	text := p.StatusText()
	if !strings.Contains(text, "Cannot read directory") {
		t.Errorf("status should show 'Cannot read directory' error, got %q", text)
	}
}

// Test table has correct number of rows matching entries
func TestPanel_TableRowCount(t *testing.T) {
	tmp := setupTestDir(t)
	p := newTestPanel(tmp)
	p.LoadDir()

	rowCount := p.Table.GetRowCount()
	if rowCount != len(p.Entries) {
		t.Errorf("table row count %d should match entries count %d", rowCount, len(p.Entries))
	}
}

// Test date format in rendered cells (FR-2)
func TestPanel_DateColumnFormat(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "test.txt"), []byte("hello"), 0644)

	p := newTestPanel(tmp)
	p.LoadDir()

	// Find test.txt (should be at row 1, after ..)
	for i, e := range p.Entries {
		if e.Name == "test.txt" {
			cell := p.Table.GetCell(i, 3) // Date moved to column 3 (after permissions column)
			if cell == nil {
				t.Fatal("date cell should exist")
			}
			// Verify YYYY-MM-DD format
			if len(cell.Text) != 10 || cell.Text[4] != '-' || cell.Text[7] != '-' {
				t.Errorf("date should be in YYYY-MM-DD format, got %q", cell.Text)
			}
			break
		}
	}
}

// Test StatusText precise format for known sizes
func TestPanel_StatusText_PreciseFormat(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("hello"), 0644)      // 5 bytes
	os.WriteFile(filepath.Join(tmp, "b.txt"), []byte("world!!!!"), 0644)  // 9 bytes

	p := newTestPanel(tmp)
	p.LoadDir()

	text := p.StatusText()
	expected := fmt.Sprintf("2 items, %s", FormatSize(14))
	if text != expected {
		t.Errorf("StatusText = %q, want %q", text, expected)
	}
}
