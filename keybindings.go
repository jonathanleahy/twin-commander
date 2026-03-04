package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)


// macOptionKeyMap maps Unicode characters produced by macOS Option+key
// (US keyboard layout) to menu indices. On macOS Terminal.app with default
// settings, Option+F sends 'ƒ' instead of Alt+F, etc.
var macOptionKeyMap = map[rune]int{
	'ƒ': 0, // Opt+F → File
	'√': 1, // Opt+V → View
	'ß': 2, // Opt+S → Search
	'©': 3, // Opt+G → Go
	'†': 4, // Opt+T → Tools
	'ø': 5, // Opt+O → Options
}

// App is the application controller.
type App struct {
	Application *tview.Application
	Pages       *tview.Pages
	LeftPanel   *Panel
	RightPanel  *Panel
	ActivePanel *Panel
	TreePanel    *TreePanel
	Viewer       *Viewer
	ViewerActive bool
	ViewMode     ViewMode
	FilterMode   bool
	FilterInput *tview.InputField
	LeftStatus  *tview.TextView
	RightStatus *tview.TextView
	RootFlex    *tview.Flex
	// TreeFocused tracks whether the tree has focus in hybrid mode
	TreeFocused bool
	KeySeq        *KeySequence
	YankBuffer    []string // paths marked for yank/paste
	SearchMode    bool
	SearchInput   *tview.InputField
	SearchTable   *tview.Table
	SearchCancel  chan struct{}
	GitRepo       *GitRepo
	PreviewPane    *tview.TextView
	PreviewWrapper *tview.Flex // Flex containing PreviewPane + scrollbar
	PreviewActive  bool
	PreviewFocused bool

	// Resizable pane proportions (percentage 10-90)
	HSplit int // Horizontal split: left pane percentage (default 50 for dual, 33 for hybrid)
	VSplit int // Vertical split: file list percentage vs preview (default 50)
	MenuBar       *MenuBar
	MenuActive    bool
	Config        Config
	Bookmarks     *BookmarkManager
	ActiveTheme   ThemeColors
	DialogActive  bool
	statusTimer   *time.Timer // auto-clear timer for status messages

	// Fuzzy finder state
	FuzzyMode   bool
	FuzzyInput  *tview.InputField
	FuzzyTable  *tview.Table
	FuzzyCancel chan struct{}

	// Directory size cache
	DirSizeCache *DirSizeCache

	// Workspace management
	WorkspaceMgr *WorkspaceManager
}

// NewApp creates and initializes the application.
// handleKeyEvent is the global InputCapture handler.
func (a *App) handleKeyEvent(event *tcell.EventKey) *tcell.EventKey {
	// Ctrl+C always quits
	if event.Key() == tcell.KeyCtrlC {
		a.saveSessionOnQuit()
		a.Application.Stop()
		return nil
	}

	// Alt+key combinations (menus, resize, clipboard)
	if event.Modifiers()&tcell.ModAlt != 0 {
		switch event.Key() {
		case tcell.KeyRune:
			// Alt+1-9: switch workspace
			if event.Rune() >= '1' && event.Rune() <= '9' {
				idx := int(event.Rune() - '1')
				a.switchWorkspace(idx)
				return nil
			}
			// Alt+C: copy selected path to clipboard
			if event.Rune() == 'c' || event.Rune() == 'C' {
				a.copyPathToClipboard()
				return nil
			}
			idx := a.MenuBar.MenuForHotkey(event.Rune())
			if idx >= 0 {
				a.activateMenuAt(idx)
				return nil
			}
		case tcell.KeyLeft:
			a.resizeSplit(-5, true)
			return nil
		case tcell.KeyRight:
			a.resizeSplit(5, true)
			return nil
		case tcell.KeyUp:
			a.resizeSplit(-5, false)
			return nil
		case tcell.KeyDown:
			a.resizeSplit(5, false)
			return nil
		}
	}

	// macOS: Option+key sends Unicode characters instead of ModAlt.
	// Map common macOS Option+key outputs to menu hotkeys.
	if event.Key() == tcell.KeyRune && !a.MenuActive && !a.FilterMode && !a.DialogActive {
		if idx, ok := macOptionKeyMap[event.Rune()]; ok {
			a.activateMenuAt(idx)
			return nil
		}
		// Option+C on macOS sends 'ç'
		if event.Rune() == 'ç' {
			a.copyPathToClipboard()
			return nil
		}
	}

	if a.MenuActive {
		return a.handleMenuKey(event)
	}
	if a.ViewerActive {
		return a.handleViewerKey(event)
	}
	if a.SearchMode {
		return a.handleSearchKey(event)
	}
	if a.FuzzyMode {
		return a.handleFuzzyKey(event)
	}
	if a.FilterMode {
		return a.handleFilterModeKey(event)
	}
	if a.PreviewFocused {
		return a.handlePreviewKey(event)
	}
	// When a dialog is open, let events pass through to the focused widget
	if a.DialogActive {
		return event
	}
	return a.handleNormalModeKey(event)
}

// handlePreviewKey handles keys when the preview pane is focused.
// handlePreviewKey handles keys when the preview pane is focused.
func (a *App) handlePreviewKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		a.switchPanel()
		return nil
	case tcell.KeyBacktab:
		a.switchPanelReverse()
		return nil
	case tcell.KeyEscape:
		a.PreviewFocused = false
		a.PreviewWrapper.SetBorderColor(a.ActiveTheme.PanelBorderInactive)
		a.closePreview()
		return nil
	case tcell.KeyPgDn:
		row, col := a.PreviewPane.GetScrollOffset()
		a.PreviewPane.ScrollTo(row+20, col)
		return nil
	case tcell.KeyPgUp:
		row, col := a.PreviewPane.GetScrollOffset()
		if row > 20 {
			a.PreviewPane.ScrollTo(row-20, col)
		} else {
			a.PreviewPane.ScrollTo(0, col)
		}
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case 'j':
			row, col := a.PreviewPane.GetScrollOffset()
			a.PreviewPane.ScrollTo(row+1, col)
			return nil
		case 'k':
			row, col := a.PreviewPane.GetScrollOffset()
			if row > 0 {
				a.PreviewPane.ScrollTo(row-1, col)
			}
			return nil
		case 'G':
			a.PreviewPane.ScrollToEnd()
			return nil
		case 'g':
			a.PreviewPane.ScrollToBeginning()
			return nil
		}
	}
	return event
}

// handleNormalModeKey handles keys in normal (non-filter) mode.
// handleNormalModeKey handles keys in normal (non-filter) mode.
func (a *App) handleNormalModeKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlT:
		a.toggleViewMode()
		return nil
	case tcell.KeyCtrlG:
		a.handleGitDiff()
		return nil
	case tcell.KeyCtrlB:
		a.showBookmarks()
		return nil
	case tcell.KeyCtrlF:
		a.enterSearchMode()
		return nil
	case tcell.KeyCtrlP:
		a.enterFuzzyMode()
		return nil
	case tcell.KeyCtrlN:
		a.createWorkspace()
		return nil
	case tcell.KeyCtrlW:
		a.closeWorkspace()
		return nil
	case tcell.KeyCtrlL:
		a.showGoToPathDialog()
		return nil
	case tcell.KeyF3:
		a.enterSearchMode()
		return nil
	case tcell.KeyCtrlUnderscore: // Ctrl+/
		a.enterContentSearch()
		return nil
	case tcell.KeyRune:
		r := event.Rune()

		// Try multi-key sequence first
		action, consumed := a.KeySeq.Feed(r)
		if action != KeyActionNone {
			a.handleKeyAction(action)
			return nil
		}
		if consumed {
			return nil // Key is pending in sequence
		}

		// Single-key bindings
		switch r {
		case 'q':
			a.saveSessionOnQuit()
			a.Application.Stop()
			return nil
		case 'j':
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k':
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		case 'h':
			a.navigateUp()
			return nil
		case 'l':
			a.handleEnter()
			return nil
		case 'G':
			a.jumpToBottom()
			return nil
		case 'p':
			a.handlePaste()
			return nil
		case '.':
			a.toggleHidden()
			return nil
		case 'r':
			a.refreshActive()
			return nil
		case '/':
			a.enterFilterMode()
			return nil
		case 's':
			a.ActivePanel.SortMode = NextSortMode(a.ActivePanel.SortMode)
			a.ActivePanel.Refresh()
			a.updateStatusBars()
			return nil
		case 'S':
			a.ActivePanel.SortOrder = ToggleSortOrder(a.ActivePanel.SortOrder)
			a.ActivePanel.Refresh()
			a.updateStatusBars()
			return nil
		case 'c':
			a.handleCopy()
			return nil
		case 'm':
			a.handleMove()
			return nil
		case 'n':
			a.handleMkdir()
			return nil
		case 'R':
			a.handleRename()
			return nil
		case 'b':
			a.handleBComp()
			return nil
		case 'e':
			a.handleOpenEditor()
			return nil
		case 't':
			a.togglePreviewPane()
			return nil
		case '~':
			a.jumpToHome()
			return nil
		case '\\':
			a.jumpToRoot()
			return nil
		case ' ':
			a.handleSelectionToggle()
			return nil
		case 'v':
			a.handleVisualStart()
			return nil
		case 'V':
			a.handleVisualEnd()
			return nil
		case '*':
			a.handleSelectionInvert()
			return nil
		case '+':
			a.handleSelectionPattern()
			return nil
		case '-':
			a.handleHistoryBack()
			return nil
		case '=':
			a.handleHistoryForward()
			return nil
		case 'o':
			a.handleOpenDefault()
			return nil
		case ':':
			a.enterCommandMode()
			return nil
		default:
			// Number keys 1-9 for bookmark jumps
			if r >= '1' && r <= '9' {
				idx := int(r - '1')
				path := a.Bookmarks.Get(idx)
				if path != "" {
					a.jumpToBookmark(path)
				}
				return nil
			}
		}
	case tcell.KeyF9, tcell.KeyF10:
		a.activateMenuBar()
		return nil
	case tcell.KeyF5:
		a.handleCopy()
		return nil
	case tcell.KeyF6:
		a.handleMove()
		return nil
	case tcell.KeyF7:
		a.handleMkdir()
		return nil
	case tcell.KeyF8:
		a.handleDelete()
		return nil
	case tcell.KeyF2:
		a.handleRename()
		return nil
	case tcell.KeyEscape:
		if a.PreviewActive {
			a.closePreview()
			return nil
		}
		return event
	case tcell.KeyTab:
		a.switchPanel()
		return nil
	case tcell.KeyBacktab:
		a.switchPanelReverse()
		return nil
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		a.navigateUp()
		return nil
	case tcell.KeyEnter:
		a.handleEnter()
		return nil
	case tcell.KeyUp:
		return event
	case tcell.KeyDown:
		return event
	}

	return event
}

// handleViewerKey handles keys when the viewer is active.
// handleViewerKey handles keys when the viewer is active.
func (a *App) handleViewerKey(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyRune && event.Rune() == 't' {
		// t: restore back to inline preview (keep preview open)
		a.restoreToPreview()
		return nil
	}
	if event.Key() == tcell.KeyEscape {
		// Esc: close viewer and preview entirely
		a.closeViewer()
		return nil
	}
	// Let viewer handle its own keys (j/k/g/G/PgUp/PgDn)
	return event
}

// openViewer opens a file in the viewer overlay.
// handleSearchKey handles keys when search mode is active.
func (a *App) handleSearchKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		a.exitSearchMode()
		return nil
	case tcell.KeyTab:
		// Toggle focus between input and results
		if a.Application.GetFocus() == a.SearchInput {
			a.Application.SetFocus(a.SearchTable)
		} else {
			a.Application.SetFocus(a.SearchInput)
		}
		return nil
	case tcell.KeyEnter:
		// If results table is focused, navigate to selected result
		if a.Application.GetFocus() == a.SearchTable {
			a.navigateToSearchResult()
			return nil
		}
		// If input is focused, switch to results
		if a.SearchTable.GetRowCount() > 0 {
			a.Application.SetFocus(a.SearchTable)
			a.SearchTable.Select(0, 0)
		}
		return nil
	}
	return event
}

// handleFuzzyKey handles keys when fuzzy finder is active.
func (a *App) handleFuzzyKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		a.exitFuzzyMode()
		return nil
	case tcell.KeyTab:
		// Toggle focus between input and results
		if a.Application.GetFocus() == a.FuzzyInput {
			a.Application.SetFocus(a.FuzzyTable)
		} else {
			a.Application.SetFocus(a.FuzzyInput)
		}
		return nil
	case tcell.KeyEnter:
		if a.Application.GetFocus() == a.FuzzyTable {
			a.navigateToFuzzyResult()
			return nil
		}
		// If input is focused, switch to results
		if a.FuzzyTable.GetRowCount() > 0 {
			a.Application.SetFocus(a.FuzzyTable)
			a.FuzzyTable.Select(0, 0)
		}
		return nil
	case tcell.KeyRune:
		// When table is focused, j/k navigate
		if a.Application.GetFocus() == a.FuzzyTable {
			switch event.Rune() {
			case 'j':
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k':
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			}
		}
	}
	return event
}

// navigateToSearchResult opens the directory containing the selected search result.
// handleFilterModeKey handles keys in filter mode.
func (a *App) handleFilterModeKey(event *tcell.EventKey) *tcell.EventKey {
	return event
}

// toggleViewMode switches between hybrid tree and dual-pane views.
// handleKeyAction dispatches a completed multi-key action.
func (a *App) handleKeyAction(action KeyAction) {
	switch action {
	case KeyActionJumpTop:
		a.jumpToTop()
	case KeyActionDelete:
		a.handleDelete()
	case KeyActionYank:
		a.handleYank()
	case KeyActionGitStage:
		a.handleGitStage()
	}
}

// jumpToTop moves cursor to the first entry.
// handleMenuKey processes key events when the menu bar is active.
func (a *App) handleMenuKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		a.deactivateMenuBar()
		return nil
	case tcell.KeyLeft:
		a.MenuBar.ActiveMenu--
		if a.MenuBar.ActiveMenu < 0 {
			a.MenuBar.ActiveMenu = len(a.MenuBar.Menus) - 1
		}
		a.MenuBar.openDropdown()
		a.Pages.RemovePage("menu-dropdown")
		a.showMenuDropdown()
		return nil
	case tcell.KeyRight:
		a.MenuBar.ActiveMenu++
		if a.MenuBar.ActiveMenu >= len(a.MenuBar.Menus) {
			a.MenuBar.ActiveMenu = 0
		}
		a.MenuBar.openDropdown()
		a.Pages.RemovePage("menu-dropdown")
		a.showMenuDropdown()
		return nil
	case tcell.KeyEnter:
		idx := a.MenuBar.Dropdown.GetCurrentItem()
		menuIdx := a.MenuBar.ActiveMenu
		if menuIdx >= 0 && menuIdx < len(a.MenuBar.Menus) {
			items := a.MenuBar.Menus[menuIdx].Items
			if idx >= 0 && idx < len(items) && !items[idx].Disabled {
				a.deactivateMenuBar()
				if items[idx].Action != nil {
					items[idx].Action()
				}
			}
		}
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case 'j':
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k':
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		case 'h':
			return tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone)
		case 'l':
			return tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone)
		case 'q':
			a.deactivateMenuBar()
			return nil
		}
	}
	return event
}

// handleMouseEvent is the global mouse capture handler.
// It detects which panel was clicked and focuses it, while letting tview's
// built-in widget mouse handlers manage row selection, scrolling, etc.
func (a *App) handleMouseEvent(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
	// Only act on left clicks
	if action != tview.MouseLeftClick && action != tview.MouseLeftDown {
		return event, action
	}

	x, y := event.Position()

	// Handle clicks on menu bar (row 0) — works whether menu is open or not
	if y == 0 {
		idx := a.MenuBar.MenuIndexAtX(x)
		if idx >= 0 {
			a.activateMenuAt(idx)
		}
		return event, action
	}

	// When menu is open, clicking outside the dropdown closes it
	if a.MenuActive {
		dropX, dropY, dropW, dropH := a.MenuBar.Dropdown.GetRect()
		if x >= dropX && x < dropX+dropW && y >= dropY && y < dropY+dropH {
			// Click is inside dropdown — let tview handle item selection
			return event, action
		}
		// Click outside menu — close it
		a.deactivateMenuBar()
		return event, action
	}

	// Don't interfere with dialogs or overlays
	if a.DialogActive || a.ViewerActive || a.SearchMode || a.FilterMode || a.FuzzyMode {
		return event, action
	}

	// Determine which widget was clicked by checking bounding rects
	if a.ViewMode == ViewHybridTree {
		// Hybrid mode: tree panel on left, right panel on right
		treeX, treeY, treeW, treeH := a.TreePanel.TreeView.GetRect()
		if x >= treeX && x < treeX+treeW && y >= treeY && y < treeY+treeH {
			if !a.TreeFocused {
				a.focusTree()
			}
			return event, action
		}

		rightX, rightY, rightW, rightH := a.RightPanel.Table.GetRect()
		if x >= rightX && x < rightX+rightW && y >= rightY && y < rightY+rightH {
			if a.TreeFocused {
				a.focusRightPanel()
			}
			return event, action
		}
	} else {
		// Dual-pane mode: left panel and right panel
		leftX, leftY, leftW, leftH := a.LeftPanel.Table.GetRect()
		if x >= leftX && x < leftX+leftW && y >= leftY && y < leftY+leftH {
			if a.ActivePanel != a.LeftPanel {
				a.focusLeftPanel()
			}
			return event, action
		}

		rightX, rightY, rightW, rightH := a.RightPanel.Table.GetRect()
		if x >= rightX && x < rightX+rightW && y >= rightY && y < rightY+rightH {
			if a.ActivePanel != a.RightPanel {
				a.focusRightPanel()
			}
			return event, action
		}
	}

	// Check preview pane
	if a.PreviewActive {
		px, py, pw, ph := a.PreviewWrapper.GetRect()
		if x >= px && x < px+pw && y >= py && y < py+ph {
			if !a.PreviewFocused {
				a.PreviewFocused = true
				a.TreeFocused = false
				a.LeftPanel.SetActive(false)
				a.RightPanel.SetActive(false)
				a.PreviewWrapper.SetBorderColor(a.ActiveTheme.PanelBorderActive)
				a.Application.SetFocus(a.PreviewPane)
			}
			return event, action
		}
	}

	return event, action
}
