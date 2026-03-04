package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// newTestApp creates a fully initialized App backed by the given directory,
// suitable for feeding key events without a real terminal. Both panels point
// to dir, the tree is rooted at dir, and the view mode is dual-pane so tests
// can exercise left/right panel switching easily.
func newTestApp(t *testing.T, dir string) *App {
	t.Helper()

	// Temporarily change working directory so NewApp picks it up
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	// Override config location so tests don't pollute the real config
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	// Suppress nerd font detection
	t.Setenv("TERM", "dumb")

	app := NewApp("")

	// Switch to dual-pane mode for predictable left/right panel tests
	app.ViewMode = ViewDualPane
	app.TreeFocused = false
	app.ActivePanel = app.LeftPanel
	app.LeftPanel.SetActive(true)
	app.RightPanel.SetActive(false)

	// Load both panels
	app.LeftPanel.Path = dir
	app.LeftPanel.LoadDir()
	app.RightPanel.Path = dir
	app.RightPanel.LoadDir()
	app.updateStatusBars()

	return app
}

// pressKey feeds a single key event into the App's global key handler.
// If handleKeyEvent returns a remapped event (e.g. j→Down), it is forwarded
// to the focused widget's InputHandler, simulating what tview's event loop does.
func pressKey(app *App, key tcell.Key, ch rune, mod tcell.ModMask) {
	ev := tcell.NewEventKey(key, ch, mod)
	result := app.handleKeyEvent(ev)
	if result != nil {
		// The event was not consumed — forward to the active widget
		dispatchToFocusedWidget(app, result)
	}
}

// dispatchToFocusedWidget sends a key event to the appropriate widget's InputHandler,
// replicating what tview.Application does when an event is not consumed by InputCapture.
func dispatchToFocusedWidget(app *App, ev *tcell.EventKey) {
	// Determine which widget has focus
	var handler func(event *tcell.EventKey, setFocus func(p tview.Primitive))
	if app.MenuActive {
		handler = app.MenuBar.Dropdown.InputHandler()
	} else if app.ViewMode == ViewHybridTree && app.TreeFocused {
		handler = app.TreePanel.TreeView.InputHandler()
	} else {
		handler = app.ActivePanel.Table.InputHandler()
	}
	if handler != nil {
		handler(ev, func(p tview.Primitive) {})
	}
}

// pressRune is a shortcut for pressing a printable rune with no modifiers.
func pressRune(app *App, ch rune) {
	pressKey(app, tcell.KeyRune, ch, tcell.ModNone)
}

// setupIntegrationDir creates a temp directory with a known file structure:
//
//	alpha/
//	beta/
//	file1.txt
//	file2.go
//	file3.md
func setupIntegrationDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	os.Mkdir(filepath.Join(dir, "alpha"), 0755)
	os.Mkdir(filepath.Join(dir, "beta"), 0755)
	os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(dir, "file2.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "file3.md"), []byte("# Title"), 0644)

	return dir
}

// ---------- Integration Tests ----------

// TestIntegration_TabSwitchesPanel verifies that Tab moves focus between panels.
func TestIntegration_TabSwitchesPanel(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	if app.ActivePanel != app.LeftPanel {
		t.Fatal("expected left panel active initially")
	}

	// Tab should switch to right panel
	pressKey(app, tcell.KeyTab, 0, tcell.ModNone)
	if app.ActivePanel != app.RightPanel {
		t.Error("expected right panel active after Tab")
	}

	// Tab again should switch back to left panel
	pressKey(app, tcell.KeyTab, 0, tcell.ModNone)
	if app.ActivePanel != app.LeftPanel {
		t.Error("expected left panel active after second Tab")
	}
}

// TestIntegration_CursorNavigation verifies j/k move the cursor.
func TestIntegration_CursorNavigation(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Initial position should be row 0 (..)
	row, _ := app.ActivePanel.Table.GetSelection()
	if row != 0 {
		t.Fatalf("expected initial row 0, got %d", row)
	}

	// j should move down
	pressRune(app, 'j')
	row, _ = app.ActivePanel.Table.GetSelection()
	if row != 1 {
		t.Errorf("expected row 1 after 'j', got %d", row)
	}

	// j again
	pressRune(app, 'j')
	row, _ = app.ActivePanel.Table.GetSelection()
	if row != 2 {
		t.Errorf("expected row 2 after second 'j', got %d", row)
	}

	// k should move back up
	pressRune(app, 'k')
	row, _ = app.ActivePanel.Table.GetSelection()
	if row != 1 {
		t.Errorf("expected row 1 after 'k', got %d", row)
	}
}

// TestIntegration_EnterDirectory verifies Enter/l on a directory navigates into it.
func TestIntegration_EnterDirectory(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Move to "alpha" (row 1, after "..")
	pressRune(app, 'j')
	entry := app.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name != "alpha" {
		name := ""
		if entry != nil {
			name = entry.Name
		}
		t.Fatalf("expected 'alpha' at row 1, got %q", name)
	}

	// l should navigate into alpha
	pressRune(app, 'l')
	expected := filepath.Join(dir, "alpha")
	if app.ActivePanel.Path != expected {
		t.Errorf("expected path %q after 'l', got %q", expected, app.ActivePanel.Path)
	}
}

// TestIntegration_NavigateUp verifies h/Backspace navigates to parent.
func TestIntegration_NavigateUp(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Navigate into alpha
	pressRune(app, 'j') // move to alpha
	pressRune(app, 'l') // enter alpha
	expected := filepath.Join(dir, "alpha")
	if app.ActivePanel.Path != expected {
		t.Fatalf("setup: expected path %q, got %q", expected, app.ActivePanel.Path)
	}

	// h should navigate back up
	pressRune(app, 'h')
	if app.ActivePanel.Path != dir {
		t.Errorf("expected path %q after 'h', got %q", dir, app.ActivePanel.Path)
	}
}

// TestIntegration_ToggleHidden verifies . toggles hidden files.
func TestIntegration_ToggleHidden(t *testing.T) {
	dir := setupIntegrationDir(t)
	// Create a hidden file
	os.WriteFile(filepath.Join(dir, ".hidden"), []byte("secret"), 0644)

	app := newTestApp(t, dir)
	initialCount := len(app.ActivePanel.Entries)

	// Toggle hidden on
	pressRune(app, '.')
	if !app.ActivePanel.ShowHidden {
		t.Error("expected ShowHidden=true after '.'")
	}
	afterToggle := len(app.ActivePanel.Entries)
	if afterToggle <= initialCount {
		t.Errorf("expected more entries with hidden shown, got %d (was %d)", afterToggle, initialCount)
	}

	// Toggle hidden off
	pressRune(app, '.')
	if app.ActivePanel.ShowHidden {
		t.Error("expected ShowHidden=false after second '.'")
	}
}

// TestIntegration_SpaceSelection verifies Space toggles file selection.
func TestIntegration_SpaceSelection(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Move to first file (skip ".." and dirs)
	// Entries are: .., alpha/, beta/, file1.txt, file2.go, file3.md
	pressRune(app, 'j') // alpha
	pressRune(app, 'j') // beta
	pressRune(app, 'j') // file1.txt

	entry := app.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name != "file1.txt" {
		name := ""
		if entry != nil {
			name = entry.Name
		}
		t.Fatalf("expected 'file1.txt', got %q", name)
	}

	// Space should select
	pressKey(app, tcell.KeyRune, ' ', tcell.ModNone)
	if app.ActivePanel.Selection.Count() != 1 {
		t.Errorf("expected 1 selected item, got %d", app.ActivePanel.Selection.Count())
	}
	selectedPath := filepath.Join(dir, "file1.txt")
	if !app.ActivePanel.Selection.IsSelected(selectedPath) {
		t.Error("file1.txt should be selected")
	}
}

// TestIntegration_SortCycle verifies s cycles sort modes.
func TestIntegration_SortCycle(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	if app.ActivePanel.SortMode != SortByName {
		t.Fatalf("expected initial sort by name, got %d", app.ActivePanel.SortMode)
	}

	pressRune(app, 's')
	if app.ActivePanel.SortMode != SortBySize {
		t.Errorf("expected sort by size after 's', got %d", app.ActivePanel.SortMode)
	}

	pressRune(app, 's')
	if app.ActivePanel.SortMode != SortByDate {
		t.Errorf("expected sort by date after second 's', got %d", app.ActivePanel.SortMode)
	}

	pressRune(app, 's')
	if app.ActivePanel.SortMode != SortByExtension {
		t.Errorf("expected sort by extension after third 's', got %d", app.ActivePanel.SortMode)
	}

	pressRune(app, 's')
	if app.ActivePanel.SortMode != SortByName {
		t.Errorf("expected sort by name after fourth 's', got %d", app.ActivePanel.SortMode)
	}
}

// TestIntegration_SortOrderToggle verifies S toggles asc/desc.
func TestIntegration_SortOrderToggle(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	if app.ActivePanel.SortOrder != SortAsc {
		t.Fatal("expected initial asc sort")
	}

	pressRune(app, 'S')
	if app.ActivePanel.SortOrder != SortDesc {
		t.Error("expected desc after 'S'")
	}

	pressRune(app, 'S')
	if app.ActivePanel.SortOrder != SortAsc {
		t.Error("expected asc after second 'S'")
	}
}

// TestIntegration_JumpToTop verifies gg jumps to top.
func TestIntegration_JumpToTop(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Move down several times
	pressRune(app, 'j')
	pressRune(app, 'j')
	pressRune(app, 'j')
	row, _ := app.ActivePanel.Table.GetSelection()
	if row < 2 {
		t.Fatalf("expected row >= 2, got %d", row)
	}

	// gg should jump to top
	pressRune(app, 'g')
	pressRune(app, 'g')
	row, _ = app.ActivePanel.Table.GetSelection()
	if row != 0 {
		t.Errorf("expected row 0 after 'gg', got %d", row)
	}
}

// TestIntegration_JumpToBottom verifies G jumps to last entry.
func TestIntegration_JumpToBottom(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	pressRune(app, 'G')
	row, _ := app.ActivePanel.Table.GetSelection()
	lastRow := len(app.ActivePanel.Entries) - 1
	if row != lastRow {
		t.Errorf("expected row %d after 'G', got %d", lastRow, row)
	}
}

// TestIntegration_HistoryBackForward verifies - and = navigate directory history.
func TestIntegration_HistoryBackForward(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Navigate into alpha
	pressRune(app, 'j') // alpha
	pressRune(app, 'l') // enter alpha
	alphaPath := filepath.Join(dir, "alpha")
	if app.ActivePanel.Path != alphaPath {
		t.Fatalf("setup: expected %q, got %q", alphaPath, app.ActivePanel.Path)
	}

	// Navigate back
	pressRune(app, '-')
	if app.ActivePanel.Path != dir {
		t.Errorf("expected %q after '-', got %q", dir, app.ActivePanel.Path)
	}

	// Navigate forward
	pressRune(app, '=')
	if app.ActivePanel.Path != alphaPath {
		t.Errorf("expected %q after '=', got %q", alphaPath, app.ActivePanel.Path)
	}
}

// TestIntegration_RefreshPreservesPosition verifies r refreshes without moving cursor.
func TestIntegration_RefreshPreservesPosition(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Move to row 2
	pressRune(app, 'j')
	pressRune(app, 'j')
	row, _ := app.ActivePanel.Table.GetSelection()
	if row != 2 {
		t.Fatalf("expected row 2, got %d", row)
	}

	// Refresh
	pressRune(app, 'r')
	row, _ = app.ActivePanel.Table.GetSelection()
	if row != 2 {
		t.Errorf("expected row 2 after refresh, got %d", row)
	}
}

// TestIntegration_MenuActivation verifies F9 opens the menu.
func TestIntegration_MenuActivation(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	if app.MenuActive {
		t.Fatal("menu should not be active initially")
	}

	pressKey(app, tcell.KeyF9, 0, tcell.ModNone)
	if !app.MenuActive {
		t.Error("menu should be active after F9")
	}

	// Escape should close
	pressKey(app, tcell.KeyEscape, 0, tcell.ModNone)
	if app.MenuActive {
		t.Error("menu should be closed after Escape")
	}
}

// TestIntegration_VisualSelection verifies v starts visual selection mode.
func TestIntegration_VisualSelection(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Move to alpha
	pressRune(app, 'j')

	// Start visual selection
	pressRune(app, 'v')
	if !app.ActivePanel.Selection.IsVisual() {
		t.Error("expected visual mode after 'v'")
	}

	// Move down to select range
	pressRune(app, 'j') // beta
	pressRune(app, 'j') // file1.txt

	// Should have selected multiple items
	count := app.ActivePanel.Selection.Count()
	if count < 2 {
		t.Errorf("expected at least 2 selected items in visual mode, got %d", count)
	}

	// V should end visual mode but keep selection
	pressRune(app, 'V')
	if app.ActivePanel.Selection.IsVisual() {
		t.Error("expected visual mode ended after 'V'")
	}
	if app.ActivePanel.Selection.Count() == 0 {
		t.Error("selection should be preserved after ending visual mode")
	}
}

// TestIntegration_InvertSelection verifies * inverts selection.
func TestIntegration_InvertSelection(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	initialCount := len(app.ActivePanel.Entries)
	if app.ActivePanel.Selection.Count() != 0 {
		t.Fatal("expected no selection initially")
	}

	// * should invert (select all except ..)
	pressRune(app, '*')
	inverted := app.ActivePanel.Selection.Count()
	if inverted == 0 {
		t.Error("expected items selected after '*'")
	}
	// Should skip ".." entry in selection
	expectedSelected := initialCount - 1 // everything except ".."
	if inverted != expectedSelected {
		t.Errorf("expected %d selected after invert, got %d", expectedSelected, inverted)
	}
}

// TestIntegration_DualPaneNavigationIndependent verifies panels navigate independently.
func TestIntegration_DualPaneNavigationIndependent(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Navigate left panel into alpha
	pressRune(app, 'j') // alpha
	pressRune(app, 'l') // enter alpha
	alphaPath := filepath.Join(dir, "alpha")
	if app.LeftPanel.Path != alphaPath {
		t.Fatalf("left panel: expected %q, got %q", alphaPath, app.LeftPanel.Path)
	}

	// Switch to right panel
	pressKey(app, tcell.KeyTab, 0, tcell.ModNone)
	if app.ActivePanel != app.RightPanel {
		t.Fatal("expected right panel active")
	}

	// Right panel should still be at original dir
	if app.RightPanel.Path != dir {
		t.Errorf("right panel should still be at %q, got %q", dir, app.RightPanel.Path)
	}

	// Navigate right panel into beta
	pressRune(app, 'j') // alpha
	pressRune(app, 'j') // beta
	pressRune(app, 'l') // enter beta
	betaPath := filepath.Join(dir, "beta")
	if app.RightPanel.Path != betaPath {
		t.Errorf("right panel: expected %q, got %q", betaPath, app.RightPanel.Path)
	}

	// Left panel should be unchanged
	if app.LeftPanel.Path != alphaPath {
		t.Errorf("left panel should still be at %q, got %q", alphaPath, app.LeftPanel.Path)
	}
}

// TestIntegration_FuzzyModeActivation verifies Ctrl+P enters fuzzy mode and Esc exits.
func TestIntegration_FuzzyModeActivation(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	if app.FuzzyMode {
		t.Fatal("fuzzy mode should not be active initially")
	}

	// Ctrl+P should enter fuzzy mode
	pressKey(app, tcell.KeyCtrlP, 0, tcell.ModNone)
	if !app.FuzzyMode {
		t.Error("expected fuzzy mode active after Ctrl+P")
	}

	// Escape should exit fuzzy mode
	pressKey(app, tcell.KeyEscape, 0, tcell.ModNone)
	if app.FuzzyMode {
		t.Error("expected fuzzy mode inactive after Escape")
	}
}

// TestIntegration_WorkspaceCreate verifies Ctrl+N creates a new workspace.
func TestIntegration_WorkspaceCreate(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	if app.WorkspaceMgr.Count() != 1 {
		t.Fatalf("expected 1 workspace initially, got %d", app.WorkspaceMgr.Count())
	}

	// Ctrl+N should create a new workspace
	pressKey(app, tcell.KeyCtrlN, 0, tcell.ModNone)
	if app.WorkspaceMgr.Count() != 2 {
		t.Errorf("expected 2 workspaces after Ctrl+N, got %d", app.WorkspaceMgr.Count())
	}
	if app.WorkspaceMgr.Active != 1 {
		t.Errorf("expected active workspace 1, got %d", app.WorkspaceMgr.Active)
	}
}

// TestIntegration_WorkspaceSwitchPreservesPath verifies switching workspaces preserves state.
func TestIntegration_WorkspaceSwitchPreservesPath(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Navigate into alpha
	pressRune(app, 'j') // alpha
	pressRune(app, 'l') // enter alpha
	alphaPath := filepath.Join(dir, "alpha")
	if app.ActivePanel.Path != alphaPath {
		t.Fatalf("setup: expected %q, got %q", alphaPath, app.ActivePanel.Path)
	}

	// Create new workspace (Ctrl+N)
	pressKey(app, tcell.KeyCtrlN, 0, tcell.ModNone)

	// New workspace should inherit the current path
	if app.ActivePanel.Path != alphaPath {
		t.Fatalf("new workspace should inherit path %q, got %q", alphaPath, app.ActivePanel.Path)
	}

	// Navigate to a different directory in workspace 2
	pressRune(app, 'h') // go up to parent
	if app.ActivePanel.Path != dir {
		t.Fatalf("expected %q after going up, got %q", dir, app.ActivePanel.Path)
	}

	// Switch back to workspace 1 (Alt+1)
	pressKey(app, tcell.KeyRune, '1', tcell.ModAlt)
	if app.ActivePanel.Path != alphaPath {
		t.Errorf("workspace 1 should still be at %q, got %q", alphaPath, app.ActivePanel.Path)
	}

	// Switch to workspace 2 (Alt+2) — should be at parent dir
	pressKey(app, tcell.KeyRune, '2', tcell.ModAlt)
	if app.ActivePanel.Path != dir {
		t.Errorf("workspace 2 should be at %q, got %q", dir, app.ActivePanel.Path)
	}
}

// TestIntegration_AnchorToggle verifies 'a' anchors and un-anchors.
func TestIntegration_AnchorToggle(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	if app.AnchorActive {
		t.Fatal("anchor should not be active initially")
	}

	// Press 'a' to anchor
	pressRune(app, 'a')
	if !app.AnchorActive {
		t.Error("expected anchor active after 'a'")
	}
	if app.AnchorPath != dir {
		t.Errorf("expected anchor path %q, got %q", dir, app.AnchorPath)
	}

	// Press 'a' again to release
	pressRune(app, 'a')
	if app.AnchorActive {
		t.Error("expected anchor released after second 'a'")
	}
	if app.AnchorPath != "" {
		t.Errorf("expected empty anchor path, got %q", app.AnchorPath)
	}
}

// TestIntegration_AnchorBlocksNavigateUp verifies anchor prevents navigating above anchor path.
func TestIntegration_AnchorBlocksNavigateUp(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Anchor at current dir
	pressRune(app, 'a')
	if !app.AnchorActive {
		t.Fatal("expected anchor active")
	}

	// Try to navigate up — should be blocked
	pressRune(app, 'h')
	if app.ActivePanel.Path != dir {
		t.Errorf("expected path unchanged at %q, got %q", dir, app.ActivePanel.Path)
	}
}

// TestIntegration_AnchorNavigateWithinScope verifies navigation works within anchor scope.
func TestIntegration_AnchorNavigateWithinScope(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Anchor at current dir
	pressRune(app, 'a')

	// Navigate into alpha (should work)
	pressRune(app, 'j') // alpha
	pressRune(app, 'l') // enter alpha
	alphaPath := filepath.Join(dir, "alpha")
	if app.ActivePanel.Path != alphaPath {
		t.Errorf("expected path %q after entering alpha, got %q", alphaPath, app.ActivePanel.Path)
	}

	// Navigate back up to anchor root (should work)
	pressRune(app, 'h')
	if app.ActivePanel.Path != dir {
		t.Errorf("expected path %q after 'h', got %q", dir, app.ActivePanel.Path)
	}

	// Try to navigate above anchor root — should be blocked
	pressRune(app, 'h')
	if app.ActivePanel.Path != dir {
		t.Errorf("anchor should block navigation above root, path is %q", app.ActivePanel.Path)
	}
}

// TestIntegration_AnchorBookmarkOutsideScope verifies bookmark outside anchor is blocked.
func TestIntegration_AnchorBookmarkOutsideScope(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Add a bookmark to /tmp
	app.Bookmarks.Add("/tmp")

	// Anchor at dir
	pressRune(app, 'a')

	// Try to jump to bookmark — should be blocked
	originalPath := app.ActivePanel.Path
	pressRune(app, '1')
	if app.ActivePanel.Path != originalPath {
		t.Errorf("expected path unchanged at %q, got %q (bookmark should be blocked)", originalPath, app.ActivePanel.Path)
	}
}

// TestIntegration_AnchorWorkspacePersistence verifies anchor state persists across workspace switches.
func TestIntegration_AnchorWorkspacePersistence(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Anchor in workspace 1
	pressRune(app, 'a')
	if !app.AnchorActive {
		t.Fatal("expected anchor active")
	}

	// Create workspace 2
	pressKey(app, tcell.KeyCtrlN, 0, tcell.ModNone)
	// Workspace 2 should inherit anchor state
	if !app.AnchorActive {
		t.Error("expected anchor active in new workspace")
	}

	// Release anchor in workspace 2
	pressRune(app, 'a')
	if app.AnchorActive {
		t.Error("expected anchor released in workspace 2")
	}

	// Switch back to workspace 1 — should still have anchor
	pressKey(app, tcell.KeyRune, '1', tcell.ModAlt)
	if !app.AnchorActive {
		t.Error("expected anchor still active in workspace 1")
	}
	if app.AnchorPath != dir {
		t.Errorf("expected anchor path %q in workspace 1, got %q", dir, app.AnchorPath)
	}
}

// TestIntegration_IsPathInScope verifies scope checking logic.
func TestIntegration_IsPathInScope(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Without anchor, everything is in scope
	if !app.isPathInScope("/tmp") {
		t.Error("without anchor, all paths should be in scope")
	}

	// Set anchor
	app.AnchorActive = true
	app.AnchorPath = dir

	// Path within scope
	if !app.isPathInScope(filepath.Join(dir, "alpha")) {
		t.Error("subdir should be in scope")
	}

	// Path outside scope
	if app.isPathInScope("/tmp") {
		t.Error("/tmp should be outside scope")
	}

	// Anchor path itself should be in scope
	if !app.isPathInScope(dir) {
		t.Error("anchor path itself should be in scope")
	}
}

// TestIntegration_GoDirModeActivation verifies gd enters directory jump mode and Esc exits.
func TestIntegration_GoDirModeActivation(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	if app.GoDirMode {
		t.Fatal("godir mode should not be active initially")
	}

	// gd should enter directory jump mode
	pressRune(app, 'g')
	pressRune(app, 'd')
	if !app.GoDirMode {
		t.Error("expected godir mode active after gd")
	}

	// Escape should exit godir mode
	pressKey(app, tcell.KeyEscape, 0, tcell.ModNone)
	if app.GoDirMode {
		t.Error("expected godir mode inactive after Escape")
	}
}

// TestIntegration_GoDirNavigates verifies selecting a result navigates to that directory.
func TestIntegration_GoDirNavigates(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Create a subdirectory to navigate to
	targetDir := filepath.Join(dir, "alpha")

	// Enter godir mode
	pressRune(app, 'g')
	pressRune(app, 'd')
	if !app.GoDirMode {
		t.Fatal("expected godir mode active")
	}

	// Simulate adding a result directly (since async search won't run in tests)
	app.GoDirTable.SetCell(0, 0,
		tview.NewTableCell("alpha").
			SetReference(targetDir))

	// Focus on table and select
	app.Application.SetFocus(app.GoDirTable)
	app.GoDirTable.Select(0, 0)

	// Press Enter to navigate
	pressKey(app, tcell.KeyEnter, 0, tcell.ModNone)

	if app.GoDirMode {
		t.Error("expected godir mode to be closed after navigating")
	}
	if app.ActivePanel.Path != targetDir {
		t.Errorf("expected panel path %q, got %q", targetDir, app.ActivePanel.Path)
	}
}

// TestIntegration_MkfileCreatesFile verifies N opens a dialog for creating a new file.
func TestIntegration_MkfileCreatesFile(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Press N to trigger new file dialog
	pressRune(app, 'N')
	if !app.DialogActive {
		t.Fatal("expected dialog active after N")
	}
}

// TestIntegration_RecentDirs verifies gr opens the recent directories dialog.
func TestIntegration_RecentDirs(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Navigate into alpha to build some history
	pressRune(app, 'j') // alpha
	pressRune(app, 'l') // enter alpha

	// Navigate back
	pressRune(app, 'h')

	// Now press gr to open recent dirs
	pressRune(app, 'g')
	pressRune(app, 'r')
	if !app.DialogActive {
		t.Error("expected dialog active after gr")
	}
}

// TestIntegration_FileDiffRequiresDualPane verifies Ctrl+D shows error in hybrid mode.
func TestIntegration_FileDiffRequiresDualPane(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Switch to hybrid mode
	app.ViewMode = ViewHybridTree

	pressKey(app, tcell.KeyCtrlD, 0, tcell.ModNone)
	// Should not open dialog — requires dual pane
	if app.DialogActive {
		t.Error("expected no dialog in hybrid mode")
	}
}

// TestIntegration_DiskUsage verifies D opens the disk usage dialog.
func TestIntegration_DiskUsage(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	pressRune(app, 'D')
	if !app.DialogActive {
		t.Error("expected dialog active after D")
	}
}

// TestIntegration_BulkRenameNoSelection verifies % shows error with no selection.
func TestIntegration_BulkRenameNoSelection(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	pressRune(app, '%')
	// Should not open dialog — no files selected
	if app.DialogActive {
		t.Error("expected no dialog with no selection")
	}
}

// TestIntegration_BulkRenameWithSelection verifies % opens dialog with selection.
func TestIntegration_BulkRenameWithSelection(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Move to first real entry, then select with Space
	pressRune(app, 'j')
	pressRune(app, ' ')

	pressRune(app, '%')
	if !app.DialogActive {
		t.Error("expected dialog active after % with selection")
	}
}

func TestIntegration_SymlinkCreate(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Move to first real entry (skip "..")
	pressRune(app, 'j')

	pressRune(app, 'L')
	if !app.DialogActive {
		t.Error("expected dialog active after L")
	}
}

func TestIntegration_SymlinkOnDotDot(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Cursor is on ".." — L should do nothing
	pressRune(app, 'L')
	if app.DialogActive {
		t.Error("expected no dialog when cursor is on ..")
	}
}

func TestIntegration_FileInfo(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Move to first real entry
	pressRune(app, 'j')

	pressRune(app, 'i')
	if !app.DialogActive {
		t.Error("expected dialog active after i")
	}
}

func TestIntegration_FileInfoOnDotDot(t *testing.T) {
	dir := setupIntegrationDir(t)
	app := newTestApp(t, dir)

	// Cursor is on ".." — i should do nothing
	pressRune(app, 'i')
	if app.DialogActive {
		t.Error("expected no dialog when cursor is on ..")
	}
}

func TestIntegration_ArchivePeek(t *testing.T) {
	dir := setupIntegrationDir(t)

	// Create a zip file in the test dir
	zipPath := filepath.Join(dir, "test.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	fw, _ := w.Create("hello.txt")
	fw.Write([]byte("hello"))
	w.Close()
	f.Close()

	app := newTestApp(t, dir)

	// Navigate to the zip file (entries are sorted, find it)
	for i := 0; i < len(app.ActivePanel.Entries); i++ {
		if app.ActivePanel.Entries[i].Name == "test.zip" {
			app.ActivePanel.Table.Select(i, 0)
			break
		}
	}

	pressKey(app, tcell.KeyEnter, 0, tcell.ModNone)
	if !app.ViewerActive {
		t.Error("expected viewer active after Enter on archive")
	}
}
