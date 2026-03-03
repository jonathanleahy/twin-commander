package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
}

// NewApp creates and initializes the application.
func NewApp() *App {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine working directory: %v\n", err)
		os.Exit(1)
	}

	app := &App{
		Application: tview.NewApplication(),
		Pages:       tview.NewPages(),
		FilterMode:  false,
		ViewMode:    ViewHybridTree,
		TreeFocused: true,
		KeySeq:      NewKeySequence(),
		HSplit:      33, // tree takes 33%, file list 67%
		VSplit:      50, // file list and preview split evenly
	}

	// Create panels
	app.LeftPanel = app.createPanel(cwd)
	app.RightPanel = app.createPanel(cwd)
	app.ActivePanel = app.RightPanel

	// Determine tree root: $HOME by default, "/" if configured
	home, _ := os.UserHomeDir()
	if home == "" {
		home = cwd
	}
	treeRoot := home

	// Load config early to check tree_root preference
	app.Config = LoadConfig()
	if app.Config.TreeRoot == "/" {
		treeRoot = "/"
	}

	// Create tree panel and viewer
	app.TreePanel = NewTreePanel(treeRoot, cwd)
	app.Viewer = NewViewer()

	// Set file select callback: Enter on a file in the tree opens preview/viewer
	app.TreePanel.OnFileSelect = func(path string) {
		if app.PreviewActive {
			app.openViewer(path)
		} else {
			app.PreviewActive = true
			app.buildLayout()
			if app.ViewMode == ViewHybridTree {
				app.Pages.SwitchToPage("hybrid")
			} else {
				app.Pages.SwitchToPage("dual")
			}
			// Sync right panel to parent dir and update preview
			dir := filepath.Dir(path)
			if dir != app.RightPanel.Path {
				app.RightPanel.Path = dir
				app.RightPanel.LoadDir()
			}
			// Select the file in right panel
			baseName := filepath.Base(path)
			for i, e := range app.RightPanel.Entries {
				if e.Name == baseName {
					app.RightPanel.Table.Select(i, 0)
					break
				}
			}
			app.updatePreview()
			app.restoreFocus()
		}
	}

	// Sync right panel when tree selection changes
	app.TreePanel.TreeView.SetChangedFunc(func(node *tview.TreeNode) {
		if app.ViewMode != ViewHybridTree {
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
			// Directory selected — sync right panel to show its contents
			if path != app.RightPanel.Path {
				app.RightPanel.Path = path
				app.RightPanel.LoadDir()
				app.updateStatusBars()
			}
		} else {
			// File selected — sync right panel to parent dir, highlight file
			dir := filepath.Dir(path)
			if dir != app.RightPanel.Path {
				app.RightPanel.Path = dir
				app.RightPanel.LoadDir()
			}
			baseName := filepath.Base(path)
			for i, e := range app.RightPanel.Entries {
				if e.Name == baseName {
					app.RightPanel.Table.Select(i, 0)
					break
				}
			}
			app.updatePreview()
			app.updateStatusBars()
		}
	})

	// Set initial active state
	app.LeftPanel.SetActive(false)
	app.RightPanel.SetActive(true)

	// Update preview pane when cursor moves
	app.RightPanel.Table.SetSelectionChangedFunc(func(row, column int) {
		app.updatePreview()
	})
	app.LeftPanel.Table.SetSelectionChangedFunc(func(row, column int) {
		app.updatePreview()
	})

	// Create status bars with padding for spacious feel
	app.LeftStatus = tview.NewTextView().SetDynamicColors(false)
	app.LeftStatus.SetTextAlign(tview.AlignLeft)
	app.LeftStatus.SetBorderPadding(0, 0, 2, 1)
	app.RightStatus = tview.NewTextView().SetDynamicColors(false)
	app.RightStatus.SetTextAlign(tview.AlignLeft)
	app.RightStatus.SetBorderPadding(0, 0, 2, 1)

	// Create filter input (hidden by default)
	app.FilterInput = tview.NewInputField().
		SetLabel("Filter: ").
		SetFieldWidth(0)

	app.FilterInput.SetChangedFunc(func(text string) {
		app.ActivePanel.SetFilter(text)
		app.updateStatusBars()
	})

	app.FilterInput.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			app.exitFilterMode(false)
		case tcell.KeyEscape:
			app.exitFilterMode(true)
		}
	})

	// Create search UI
	app.SearchInput = tview.NewInputField().
		SetLabel("Search: ").
		SetFieldWidth(0)
	app.SearchTable = tview.NewTable().
		SetSelectable(true, false)
	app.SearchTable.SetBorder(true)
	app.SearchTable.SetTitle("Results")
	app.SearchTable.SetBorderPadding(0, 0, 1, 1)

	app.setupSearchHandlers()

	// Create inline preview pane with scrollbar (before applyConfig which themes it)
	app.PreviewPane = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	app.PreviewWrapper, _ = WrapWithScrollBar(app.PreviewPane)
	app.PreviewWrapper.SetBorder(true)
	app.PreviewWrapper.SetTitle(" Preview ")
	app.PreviewWrapper.SetBorderPadding(0, 0, 1, 1)

	// Build menu bar (before applyConfig which themes it)
	app.buildMenuBar()

	// Config already loaded early (for tree_root); set up bookmarks
	app.Bookmarks = NewBookmarkManager(app.Config.Bookmarks)
	app.applyConfig()

	// Build layouts for both modes
	app.buildLayout()

	// Load initial directories
	app.LeftPanel.LoadDir()
	app.RightPanel.LoadDir()

	// Detect git repo
	app.GitRepo = DetectGitRepo(cwd)
	app.LeftPanel.GitRepo = app.GitRepo
	app.RightPanel.GitRepo = app.GitRepo

	app.updateStatusBars()

	// Set up global key handler
	app.Application.SetInputCapture(app.handleKeyEvent)

	app.Application.SetRoot(app.Pages, true)
	app.setInitialFocus()

	// Check for nerd font in background (fc-list can be slow)
	if !app.Config.NerdFontDismissed {
		go func() {
			if !HasNerdFont() {
				app.Application.QueueUpdateDraw(func() {
					app.showNerdFontWarning()
				})
			}
		}()
	}

	return app
}

// showBookmarks opens the bookmark manager dialog.
func (a *App) showBookmarks() {
	a.DialogActive = true
	currentPath := a.ActivePanel.Path
	if a.ViewMode == ViewHybridTree && a.TreeFocused {
		currentPath = a.TreePanel.SelectedPath()
	}
	ShowBookmarkDialog(a.Pages, a.Application, a.Bookmarks, currentPath,
		func(path string) {
			a.jumpToBookmark(path)
			a.saveBookmarks()
		},
		func() {
			a.DialogActive = false
			a.saveBookmarks()
			a.restoreFocus()
		},
	)
}

// jumpToBookmark navigates to a bookmarked directory.
func (a *App) jumpToBookmark(path string) {
	if a.ViewMode == ViewHybridTree {
		a.TreePanel.NavigateToPath(path)
		a.RightPanel.Path = path
		a.RightPanel.LoadDir()
	} else {
		a.ActivePanel.Path = path
		a.ActivePanel.LoadDir()
	}
	a.updateStatusBars()
}

// jumpToBookmarkNum jumps to a bookmark by 0-based index.
func (a *App) jumpToBookmarkNum(idx int) {
	path := a.Bookmarks.Get(idx)
	if path != "" {
		a.jumpToBookmark(path)
	}
}

// saveBookmarks persists bookmarks to the config file.
func (a *App) saveBookmarks() {
	a.Config.Bookmarks = a.Bookmarks.Paths()
	_ = SaveConfig(a.Config)
}

// applyConfig applies loaded configuration settings to the app state.
func (a *App) applyConfig() {
	// View mode
	switch a.Config.DefaultViewMode {
	case "dual":
		a.ViewMode = ViewDualPane
		a.TreeFocused = false
	default:
		a.ViewMode = ViewHybridTree
		a.TreeFocused = true
	}

	// Sort mode
	var sortMode SortMode
	switch a.Config.DefaultSortMode {
	case "size":
		sortMode = SortBySize
	case "date":
		sortMode = SortByDate
	case "extension":
		sortMode = SortByExtension
	default:
		sortMode = SortByName
	}
	sortOrder := SortAsc
	if !a.Config.DefaultSortAsc {
		sortOrder = SortDesc
	}

	a.LeftPanel.SortMode = sortMode
	a.LeftPanel.SortOrder = sortOrder
	a.RightPanel.SortMode = sortMode
	a.RightPanel.SortOrder = sortOrder

	// Hidden files
	a.LeftPanel.ShowHidden = a.Config.ShowHidden
	a.RightPanel.ShowHidden = a.Config.ShowHidden
	a.TreePanel.ShowHidden = a.Config.ShowHidden

	// Preview on start
	a.PreviewActive = a.Config.PreviewOnStart

	// Theme
	a.ActiveTheme = GetTheme(ThemeName(a.Config.Theme))
	a.applyThemeColors()
}

// applyThemeColors applies the active theme to all UI elements.
func (a *App) applyThemeColors() {
	tc := a.ActiveTheme
	a.MenuBar.ApplyTheme(tc)

	// Panel border colors
	a.LeftPanel.ActiveBorderColor = tc.PanelBorderActive
	a.LeftPanel.InactiveBorderColor = tc.PanelBorderInactive
	a.RightPanel.ActiveBorderColor = tc.PanelBorderActive
	a.RightPanel.InactiveBorderColor = tc.PanelBorderInactive

	// Tree panel border
	if a.TreeFocused {
		a.TreePanel.TreeView.SetBorderColor(tc.PanelBorderActive)
	} else {
		a.TreePanel.TreeView.SetBorderColor(tc.PanelBorderInactive)
	}

	// Preview wrapper border
	if a.PreviewFocused {
		a.PreviewWrapper.SetBorderColor(tc.PanelBorderActive)
	} else {
		a.PreviewWrapper.SetBorderColor(tc.PanelBorderInactive)
	}

	// Re-apply active state borders
	a.LeftPanel.SetActive(a.ActivePanel == a.LeftPanel)
	a.RightPanel.SetActive(a.ActivePanel == a.RightPanel)
}

// resizeSplit adjusts pane proportions. horizontal=true adjusts left/right split,
// horizontal=false adjusts top/bottom (file list vs preview).
// delta is the percentage change (positive = grow left/top, negative = shrink).
func (a *App) resizeSplit(delta int, horizontal bool) {
	if horizontal {
		a.HSplit = clamp(a.HSplit+delta, 15, 85)
	} else {
		if !a.PreviewActive {
			return
		}
		a.VSplit = clamp(a.VSplit+delta, 15, 85)
	}
	a.buildLayout()
	if a.ViewMode == ViewHybridTree {
		a.Pages.SwitchToPage("hybrid")
	} else {
		a.Pages.SwitchToPage("dual")
	}
	a.restoreFocus()
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// setTheme changes the active theme, applies colors, and saves to config.
func (a *App) setTheme(name ThemeName) {
	a.Config.Theme = string(name)
	a.ActiveTheme = GetTheme(name)
	a.applyThemeColors()
	_ = SaveConfig(a.Config)
}

// buildLayout constructs Pages with both view mode layouts.
func (a *App) buildLayout() {
	// Remove existing pages to allow rebuild
	a.Pages.RemovePage("hybrid")
	a.Pages.RemovePage("dual")
	a.Pages.RemovePage("viewer")
	a.Pages.RemovePage("search")

	// Hybrid tree layout: tree on left, file panel on right
	treeCol := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.TreePanel.TreeView, 0, 1, true).
		AddItem(a.LeftStatus, 2, 0, false)

	var rightCol *tview.Flex
	if a.PreviewActive {
		rightCol = tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(a.RightPanel.Table, 0, a.VSplit, false).
			AddItem(a.PreviewWrapper, 0, 100-a.VSplit, false).
			AddItem(a.RightStatus, 2, 0, false)
	} else {
		rightCol = tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(a.RightPanel.Table, 0, 1, false).
			AddItem(a.RightStatus, 2, 0, false)
	}

	hybridFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(treeCol, 0, a.HSplit, true).
		AddItem(rightCol, 0, 100-a.HSplit, false)

	hybridRoot := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.MenuBar.Bar, 1, 0, false).
		AddItem(hybridFlex, 0, 1, true).
		AddItem(a.FilterInput, 0, 0, false)

	// Dual-pane layout: two file panels side by side
	leftCol := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.LeftPanel.Table, 0, 1, true).
		AddItem(a.LeftStatus, 2, 0, false)

	var rightCol2 *tview.Flex
	if a.PreviewActive {
		rightCol2 = tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(a.RightPanel.Table, 0, a.VSplit, false).
			AddItem(a.PreviewWrapper, 0, 100-a.VSplit, false).
			AddItem(a.RightStatus, 2, 0, false)
	} else {
		rightCol2 = tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(a.RightPanel.Table, 0, 1, false).
			AddItem(a.RightStatus, 2, 0, false)
	}

	dualHSplit := 50 // dual pane is always 50/50 for left/right
	dualFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(leftCol, 0, dualHSplit, true).
		AddItem(rightCol2, 0, 100-dualHSplit, false)

	a.RootFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.MenuBar.Bar, 1, 0, false).
		AddItem(dualFlex, 0, 1, true).
		AddItem(a.FilterInput, 0, 0, false)

	// Search overlay
	searchFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.SearchInput, 1, 0, true).
		AddItem(a.SearchTable, 0, 1, false)

	// Viewer layout: menu bar on top, viewer below
	viewerRoot := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.MenuBar.Bar, 1, 0, false).
		AddItem(a.Viewer.Wrapper, 0, 1, true)

	isHybrid := a.ViewMode == ViewHybridTree
	a.Pages.AddPage("hybrid", hybridRoot, true, isHybrid)
	a.Pages.AddPage("dual", a.RootFlex, true, !isHybrid)
	a.Pages.AddPage("viewer", viewerRoot, true, false)
	a.Pages.AddPage("search", searchFlex, true, false)
}

// setInitialFocus sets focus based on current view mode.
func (a *App) setInitialFocus() {
	if a.ViewMode == ViewHybridTree {
		a.Application.SetFocus(a.TreePanel.TreeView)
		a.TreeFocused = true
	} else {
		a.Application.SetFocus(a.LeftPanel.Table)
	}
}

// createPanel creates a new Panel with the given path.
func (a *App) createPanel(path string) *Panel {
	table := tview.NewTable()
	table.SetBorder(true)
	table.SetSelectable(true, false)
	table.SetSelectedStyle(tcell.StyleDefault.Reverse(true))
	table.SetFixed(0, 0)
	table.SetBorderPadding(0, 0, 1, 1) // left/right inner padding for breathing room

	return &Panel{
		Path:                path,
		Table:               table,
		ShowHidden:          false,
		Filter:              "",
		ActiveBorderColor:   tcell.ColorAqua,
		InactiveBorderColor: tcell.ColorDefault,
	}
}

// Run starts the application.
func (a *App) Run() error {
	return a.Application.Run()
}

// handleKeyEvent is the global InputCapture handler.
func (a *App) handleKeyEvent(event *tcell.EventKey) *tcell.EventKey {
	// Ctrl+C always quits
	if event.Key() == tcell.KeyCtrlC {
		a.Application.Stop()
		return nil
	}

	// Alt+key combinations (menus, resize, clipboard)
	if event.Modifiers()&tcell.ModAlt != 0 {
		switch event.Key() {
		case tcell.KeyRune:
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

	if a.MenuActive {
		return a.handleMenuKey(event)
	}
	if a.ViewerActive {
		return a.handleViewerKey(event)
	}
	if a.SearchMode {
		return a.handleSearchKey(event)
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
	case tcell.KeyCtrlL:
		a.showGoToPathDialog()
		return nil
	case tcell.KeyF3:
		a.enterSearchMode()
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
	case tcell.KeyF9:
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
func (a *App) openViewer(path string) {
	err := a.Viewer.Open(path)
	if err != nil {
		a.setStatusError(fmt.Sprintf("Cannot open: %v", err))
		return
	}
	a.ViewerActive = true
	a.Pages.SwitchToPage("viewer")
	a.Application.SetFocus(a.Viewer.TextView)
}

// closeViewer returns from viewer to the previous view.
func (a *App) closeViewer() {
	a.ViewerActive = false
	if a.ViewMode == ViewHybridTree {
		a.Pages.SwitchToPage("hybrid")
		if a.TreeFocused {
			a.Application.SetFocus(a.TreePanel.TreeView)
		} else {
			a.Application.SetFocus(a.RightPanel.Table)
		}
	} else {
		a.Pages.SwitchToPage("dual")
		a.Application.SetFocus(a.ActivePanel.Table)
	}
}

// restoreToPreview returns from full-screen viewer back to inline preview.
func (a *App) restoreToPreview() {
	a.ViewerActive = false

	// Ensure inline preview is visible
	if !a.PreviewActive {
		a.PreviewActive = true
		a.buildLayout()
	}

	if a.ViewMode == ViewHybridTree {
		a.Pages.SwitchToPage("hybrid")
	} else {
		a.Pages.SwitchToPage("dual")
	}
	a.restoreFocus()
}

// setupSearchHandlers configures the search input and result table behaviours.
func (a *App) setupSearchHandlers() {
	var debounceTimer *time.Timer

	a.SearchInput.SetChangedFunc(func(text string) {
		if debounceTimer != nil {
			debounceTimer.Stop()
		}
		debounceTimer = time.AfterFunc(200*time.Millisecond, func() {
			a.Application.QueueUpdateDraw(func() {
				a.runSearch(text)
			})
		})
	})
}

// enterSearchMode activates the search overlay.
func (a *App) enterSearchMode() {
	a.SearchMode = true
	a.SearchInput.SetText("")
	a.SearchTable.Clear()
	a.Pages.SwitchToPage("search")
	a.Application.SetFocus(a.SearchInput)
}

// exitSearchMode returns from search to the previous view.
func (a *App) exitSearchMode() {
	a.SearchMode = false
	if a.SearchCancel != nil {
		close(a.SearchCancel)
		a.SearchCancel = nil
	}
	if a.ViewMode == ViewHybridTree {
		a.Pages.SwitchToPage("hybrid")
		if a.TreeFocused {
			a.Application.SetFocus(a.TreePanel.TreeView)
		} else {
			a.Application.SetFocus(a.RightPanel.Table)
		}
	} else {
		a.Pages.SwitchToPage("dual")
		a.Application.SetFocus(a.ActivePanel.Table)
	}
}

// runSearch executes a search with the given query.
func (a *App) runSearch(query string) {
	// Cancel previous search
	if a.SearchCancel != nil {
		close(a.SearchCancel)
	}
	a.SearchCancel = make(chan struct{})
	cancelCh := a.SearchCancel

	a.SearchTable.Clear()

	if query == "" {
		return
	}

	rootDir := a.ActivePanel.Path
	if a.ViewMode == ViewHybridTree {
		rootDir = a.TreePanel.RootPath
	}

	resultCh := make(chan SearchResult, 100)
	go Search(SearchOpts{
		RootDir:    rootDir,
		Query:      query,
		MaxResults: 1000,
		ShowHidden: a.ActivePanel.ShowHidden,
	}, resultCh, cancelCh)

	go func() {
		row := 0
		for result := range resultCh {
			r := result
			rowIdx := row
			a.Application.QueueUpdateDraw(func() {
				style := tcell.StyleDefault
				if r.IsDir {
					style = style.Foreground(tcell.ColorBlue).Bold(true)
				}
				a.SearchTable.SetCell(rowIdx, 0,
					tview.NewTableCell(r.RelPath).
						SetStyle(style).
						SetReference(r.Path))
			})
			row++
		}
	}()
}

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

// navigateToSearchResult opens the directory containing the selected search result.
func (a *App) navigateToSearchResult() {
	row, _ := a.SearchTable.GetSelection()
	cell := a.SearchTable.GetCell(row, 0)
	if cell == nil {
		return
	}
	ref := cell.GetReference()
	if ref == nil {
		return
	}
	fullPath := ref.(string)

	// Navigate to the directory containing the file
	dir := filepath.Dir(fullPath)
	baseName := filepath.Base(fullPath)

	a.exitSearchMode()

	// Navigate the active panel to that directory
	a.ActivePanel.Path = dir
	a.ActivePanel.LoadDir()

	// Try to select the file
	for i, e := range a.ActivePanel.Entries {
		if e.Name == baseName {
			a.ActivePanel.Table.Select(i, 0)
			break
		}
	}

	// Sync tree if in hybrid mode
	if a.ViewMode == ViewHybridTree {
		a.TreePanel.NavigateToPath(dir)
	}

	a.updateStatusBars()
}

// handleFilterModeKey handles keys in filter mode.
func (a *App) handleFilterModeKey(event *tcell.EventKey) *tcell.EventKey {
	return event
}

// toggleViewMode switches between hybrid tree and dual-pane views.
func (a *App) toggleViewMode() {
	if a.ViewMode == ViewHybridTree {
		a.ViewMode = ViewDualPane
		a.Pages.SwitchToPage("dual")
		// Sync left panel to tree's current path
		a.LeftPanel.Path = a.TreePanel.SelectedPath()
		a.LeftPanel.LoadDir()
		a.ActivePanel = a.LeftPanel
		a.LeftPanel.SetActive(true)
		a.RightPanel.SetActive(false)
		a.Application.SetFocus(a.LeftPanel.Table)
		a.TreeFocused = false
	} else {
		a.ViewMode = ViewHybridTree
		a.Pages.SwitchToPage("hybrid")
		// Sync tree to left panel's path (expand in-place, don't reset root)
		a.TreePanel.NavigateToPath(a.LeftPanel.Path)
		a.RightPanel.Path = a.LeftPanel.Path
		a.RightPanel.LoadDir()
		a.ActivePanel = a.RightPanel
		a.TreeFocused = true
		a.Application.SetFocus(a.TreePanel.TreeView)
	}
	a.updateStatusBars()
}

// toggleHidden toggles hidden files in active panel and tree.
func (a *App) toggleHidden() {
	if a.ViewMode == ViewHybridTree && a.TreeFocused {
		a.TreePanel.ToggleHidden()
		a.RightPanel.ShowHidden = a.TreePanel.ShowHidden
		a.RightPanel.LoadDir()
	} else {
		a.ActivePanel.ToggleHidden()
	}
	a.updateStatusBars()
}

// refreshActive refreshes the currently active view.
func (a *App) refreshActive() {
	if a.ViewMode == ViewHybridTree && a.TreeFocused {
		a.TreePanel.Refresh()
		a.RightPanel.Refresh()
	} else {
		a.ActivePanel.Refresh()
	}
	a.updateStatusBars()
}

// switchPanel cycles focus: panels → preview (if active) → back.
// switchPanel cycles focus through available panes.
// Hybrid mode: tree → files → preview → tree
// Dual mode:   left → right → preview → left
func (a *App) switchPanel() {
	// Helper to unfocus all panes
	unfocusAll := func() {
		a.TreeFocused = false
		a.PreviewFocused = false
		a.TreePanel.TreeView.SetBorderColor(a.ActiveTheme.PanelBorderInactive)
		a.LeftPanel.SetActive(false)
		a.RightPanel.SetActive(false)
		a.PreviewWrapper.SetBorderColor(a.ActiveTheme.PanelBorderInactive)
	}

	if a.ViewMode == ViewHybridTree {
		// Cycle: tree → files → preview → tree
		if a.PreviewFocused {
			unfocusAll()
			a.TreeFocused = true
			a.TreePanel.TreeView.SetBorderColor(a.ActiveTheme.PanelBorderActive)
			a.Application.SetFocus(a.TreePanel.TreeView)
		} else if a.TreeFocused {
			unfocusAll()
			a.ActivePanel = a.RightPanel
			a.RightPanel.SetActive(true)
			a.Application.SetFocus(a.RightPanel.Table)
		} else if a.PreviewActive {
			unfocusAll()
			a.PreviewFocused = true
			a.PreviewWrapper.SetBorderColor(a.ActiveTheme.PanelBorderActive)
			a.Application.SetFocus(a.PreviewPane)
		} else {
			// No preview — cycle back to tree
			unfocusAll()
			a.TreeFocused = true
			a.ActivePanel = a.RightPanel
			a.TreePanel.TreeView.SetBorderColor(a.ActiveTheme.PanelBorderActive)
			a.Application.SetFocus(a.TreePanel.TreeView)
		}
	} else {
		// Cycle: left → right → preview → left
		if a.PreviewFocused {
			unfocusAll()
			a.ActivePanel = a.LeftPanel
			a.LeftPanel.SetActive(true)
			a.Application.SetFocus(a.LeftPanel.Table)
		} else if a.ActivePanel == a.LeftPanel {
			unfocusAll()
			a.ActivePanel = a.RightPanel
			a.RightPanel.SetActive(true)
			a.Application.SetFocus(a.RightPanel.Table)
		} else if a.PreviewActive {
			unfocusAll()
			a.PreviewFocused = true
			a.PreviewWrapper.SetBorderColor(a.ActiveTheme.PanelBorderActive)
			a.Application.SetFocus(a.PreviewPane)
		} else {
			unfocusAll()
			a.ActivePanel = a.LeftPanel
			a.LeftPanel.SetActive(true)
			a.Application.SetFocus(a.LeftPanel.Table)
		}
	}
}

// switchPanelReverse cycles focus in reverse order.
// Hybrid mode: tree ← files ← preview ← tree
// Dual mode:   left ← right ← preview ← left
func (a *App) switchPanelReverse() {
	unfocusAll := func() {
		a.TreeFocused = false
		a.PreviewFocused = false
		a.TreePanel.TreeView.SetBorderColor(a.ActiveTheme.PanelBorderInactive)
		a.LeftPanel.SetActive(false)
		a.RightPanel.SetActive(false)
		a.PreviewWrapper.SetBorderColor(a.ActiveTheme.PanelBorderInactive)
	}

	if a.ViewMode == ViewHybridTree {
		// Reverse cycle: tree ← preview ← files ← tree
		if a.TreeFocused {
			unfocusAll()
			if a.PreviewActive {
				a.PreviewFocused = true
				a.PreviewWrapper.SetBorderColor(a.ActiveTheme.PanelBorderActive)
				a.Application.SetFocus(a.PreviewPane)
			} else {
				a.ActivePanel = a.RightPanel
				a.RightPanel.SetActive(true)
				a.Application.SetFocus(a.RightPanel.Table)
			}
		} else if a.PreviewFocused {
			unfocusAll()
			a.ActivePanel = a.RightPanel
			a.RightPanel.SetActive(true)
			a.Application.SetFocus(a.RightPanel.Table)
		} else {
			// Files → tree
			unfocusAll()
			a.TreeFocused = true
			a.TreePanel.TreeView.SetBorderColor(a.ActiveTheme.PanelBorderActive)
			a.Application.SetFocus(a.TreePanel.TreeView)
		}
	} else {
		// Reverse cycle: left ← preview ← right ← left
		if a.ActivePanel == a.LeftPanel && !a.PreviewFocused {
			unfocusAll()
			if a.PreviewActive {
				a.PreviewFocused = true
				a.PreviewWrapper.SetBorderColor(a.ActiveTheme.PanelBorderActive)
				a.Application.SetFocus(a.PreviewPane)
			} else {
				a.ActivePanel = a.RightPanel
				a.RightPanel.SetActive(true)
				a.Application.SetFocus(a.RightPanel.Table)
			}
		} else if a.PreviewFocused {
			unfocusAll()
			a.ActivePanel = a.RightPanel
			a.RightPanel.SetActive(true)
			a.Application.SetFocus(a.RightPanel.Table)
		} else {
			unfocusAll()
			a.ActivePanel = a.LeftPanel
			a.LeftPanel.SetActive(true)
			a.Application.SetFocus(a.LeftPanel.Table)
		}
	}
}

// navigateUp handles Backspace/h in normal mode.
// In tree mode: collapse expanded node, or move cursor to parent.
func (a *App) navigateUp() {
	if a.ViewMode == ViewHybridTree && a.TreeFocused {
		if a.TreePanel.SelectedIsExpanded() {
			a.TreePanel.CollapseSelected()
		} else {
			a.TreePanel.MoveToParent()
		}
	} else {
		a.ActivePanel.NavigateUp()
	}
	a.updateStatusBars()
}

// jumpToHome navigates the tree to $HOME.
func (a *App) jumpToHome() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	if a.ViewMode == ViewHybridTree {
		a.TreePanel.SetRootPath(home)
		a.RightPanel.Path = home
		a.RightPanel.LoadDir()
	} else {
		a.ActivePanel.Path = home
		a.ActivePanel.LoadDir()
	}
	a.updateStatusBars()
}

// jumpToRoot sets the tree root to "/" (full filesystem).
func (a *App) jumpToRoot() {
	if a.ViewMode == ViewHybridTree {
		currentPath := a.TreePanel.SelectedPath()
		a.TreePanel.SetRootPath("/")
		a.TreePanel.ExpandToPath(currentPath)
		a.updateStatusBars()
	}
}

// showGoToPathDialog shows an input dialog for manual path entry.
func (a *App) showGoToPathDialog() {
	startPath := ""
	if a.ViewMode == ViewHybridTree && a.TreeFocused {
		startPath = a.TreePanel.SelectedPath()
	} else {
		startPath = a.ActivePanel.Path
	}

	a.DialogActive = true
	ShowInputDialog(a.Pages, a.Application, "Go to Path", "Path: ", startPath, func(value string, cancelled bool) {
		a.DialogActive = false
		if cancelled || value == "" {
			a.restoreFocus()
			return
		}

		// Expand ~ to home dir
		if strings.HasPrefix(value, "~") {
			home, err := os.UserHomeDir()
			if err == nil {
				value = filepath.Join(home, value[1:])
			}
		}

		// Resolve to absolute path
		absPath, err := filepath.Abs(value)
		if err != nil {
			a.setStatusError(fmt.Sprintf("Invalid path: %v", err))
			a.restoreFocus()
			return
		}

		// Check path exists
		info, err := os.Stat(absPath)
		if err != nil {
			a.setStatusError(fmt.Sprintf("Path not found: %s", absPath))
			a.restoreFocus()
			return
		}

		target := absPath
		if !info.IsDir() {
			target = filepath.Dir(absPath)
		}

		if a.ViewMode == ViewHybridTree {
			a.TreePanel.NavigateToPath(target)
			a.RightPanel.Path = target
			a.RightPanel.LoadDir()
		} else {
			a.ActivePanel.Path = target
			a.ActivePanel.LoadDir()
		}
		a.updateStatusBars()
		a.restoreFocus()
	})
}

// handleEnter handles Enter on the selected entry.
func (a *App) handleEnter() {
	if a.ViewMode == ViewHybridTree && a.TreeFocused {
		// Tree handles its own Enter via SetSelectedFunc
		return
	}

	entry := a.ActivePanel.SelectedEntry()
	if entry == nil {
		return
	}

	if entry.Name == ".." {
		a.navigateUp()
		return
	}

	if !entry.IsDir {
		if a.PreviewActive {
			// Already previewing — Enter escalates to full-screen viewer
			a.openViewer(filepath.Join(a.ActivePanel.Path, entry.Name))
		} else {
			// First Enter on a file opens inline preview
			a.PreviewActive = true
			a.buildLayout()
			if a.ViewMode == ViewHybridTree {
				a.Pages.SwitchToPage("hybrid")
			} else {
				a.Pages.SwitchToPage("dual")
			}
			a.updatePreview()
			a.restoreFocus()
		}
		return
	}

	if !entry.Accessible {
		a.setStatusError(fmt.Sprintf("Permission denied: %s",
			filepath.Join(a.ActivePanel.Path, entry.Name)))
		return
	}

	err := a.ActivePanel.TryNavigateInto(entry.Name)
	if err != nil {
		a.setStatusError(fmt.Sprintf("Cannot read directory: %s",
			filepath.Join(a.ActivePanel.Path, entry.Name)))
		return
	}
	a.updateStatusBars()
}

// enterFilterMode switches to filter mode.
func (a *App) enterFilterMode() {
	if a.ViewMode == ViewHybridTree && a.TreeFocused {
		return // Filter only applies to file panels
	}
	a.FilterMode = true
	a.FilterInput.SetText(a.ActivePanel.Filter)

	// Show filter input — find the correct root flex
	if a.ViewMode == ViewHybridTree {
		// hybridRoot is the parent page
		a.resizeFilterInput(1)
	} else {
		a.RootFlex.ResizeItem(a.FilterInput, 1, 0)
	}

	a.Application.SetFocus(a.FilterInput)
}

// exitFilterMode exits filter mode.
func (a *App) exitFilterMode(clearFilter bool) {
	a.FilterMode = false

	if clearFilter {
		a.ActivePanel.ClearFilter()
		a.FilterInput.SetText("")
	}

	if a.ViewMode == ViewHybridTree {
		a.resizeFilterInput(0)
	} else {
		a.RootFlex.ResizeItem(a.FilterInput, 0, 0)
	}

	a.Application.SetFocus(a.ActivePanel.Table)
	a.updateStatusBars()
}

// resizeFilterInput is a helper that resizes the filter input in the hybrid layout.
func (a *App) resizeFilterInput(height int) {
	// The filter input is shared across both layouts, so resize in both
	a.RootFlex.ResizeItem(a.FilterInput, height, 0)
	// Also need to resize in the hybrid root — walk Pages to find it
	// Since both layouts contain the same FilterInput, this works
}

// updateStatusBars refreshes both status bar texts.
func (a *App) updateStatusBars() {
	gitPrefix := ""
	if a.GitRepo != nil && a.GitRepo.Branch != "" {
		gitPrefix = "[" + a.GitRepo.Branch + "] "
	}

	if a.ViewMode == ViewHybridTree {
		a.LeftStatus.SetText(gitPrefix + a.TreePanel.SelectedPath())
		a.RightStatus.SetText(a.RightPanel.StatusText())
	} else {
		a.LeftStatus.SetText(gitPrefix + a.LeftPanel.StatusText())
		a.RightStatus.SetText(a.RightPanel.StatusText())
	}
}

// setStatusError displays an error message in the active panel's status bar.
func (a *App) setStatusError(msg string) {
	if a.ViewMode == ViewHybridTree {
		a.RightStatus.SetText(msg)
	} else if a.ActivePanel == a.LeftPanel {
		a.LeftStatus.SetText(msg)
	} else {
		a.RightStatus.SetText(msg)
	}
	a.ActivePanel.StatusBar = msg
}

// handleCopy copies the selected entry to the other panel's directory.
func (a *App) handleCopy() {
	entry := a.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name == ".." {
		return
	}
	src := filepath.Join(a.ActivePanel.Path, entry.Name)
	dstDir := a.InactivePanel().Path
	dst := filepath.Join(dstDir, entry.Name)

	msg := fmt.Sprintf("Copy %q to %s?", entry.Name, dstDir)
	if PathExists(dst) {
		msg = fmt.Sprintf("Copy %q to %s? (will overwrite existing)", entry.Name, dstDir)
	}

	a.DialogActive = true
	ShowConfirmDialog(a.Pages, "Copy", msg, func(confirmed bool) {
		a.DialogActive = false
		if !confirmed {
			a.restoreFocus()
			return
		}
		var err error
		if entry.IsDir {
			err = CopyDir(src, dst)
		} else {
			err = CopyFile(src, dst)
		}
		if err != nil {
			ShowErrorDialog(a.Pages, fmt.Sprintf("Copy failed: %v", err))
		}
		a.refreshAllPanels()
		a.restoreFocus()
	})
}

// handleMove moves the selected entry to the other panel's directory.
func (a *App) handleMove() {
	entry := a.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name == ".." {
		return
	}
	src := filepath.Join(a.ActivePanel.Path, entry.Name)
	dstDir := a.InactivePanel().Path
	dst := filepath.Join(dstDir, entry.Name)

	msg := fmt.Sprintf("Move %q to %s?", entry.Name, dstDir)
	if PathExists(dst) {
		msg = fmt.Sprintf("Move %q to %s? (will overwrite existing)", entry.Name, dstDir)
	}

	a.DialogActive = true
	ShowConfirmDialog(a.Pages, "Move", msg, func(confirmed bool) {
		a.DialogActive = false
		if !confirmed {
			a.restoreFocus()
			return
		}
		err := MoveFile(src, dst)
		if err != nil {
			ShowErrorDialog(a.Pages, fmt.Sprintf("Move failed: %v", err))
		}
		a.refreshAllPanels()
		a.restoreFocus()
	})
}

// handleDelete offers trash (default) or permanent delete for the selected entry.
func (a *App) handleDelete() {
	entry := a.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name == ".." {
		return
	}
	path := filepath.Join(a.ActivePanel.Path, entry.Name)

	sizeStr := ""
	if entry.IsDir {
		size, _ := CalcDirSize(path)
		sizeStr = fmt.Sprintf(" (%s)", FormatSize(size))
	}

	msg := fmt.Sprintf("Delete %q%s?", entry.Name, sizeStr)

	a.DialogActive = true
	ShowChoiceDialog(a.Pages, "Delete", msg, []string{"Move to Trash", "Permanently Delete", "Cancel"}, func(label string) {
		a.DialogActive = false
		switch label {
		case "Move to Trash":
			err := MoveToTrash(path)
			if err != nil {
				// Fallback to permanent delete if trash fails
				ShowErrorDialog(a.Pages, fmt.Sprintf("Trash failed: %v\nUse permanent delete instead.", err))
			}
		case "Permanently Delete":
			err := DeletePath(path)
			if err != nil {
				ShowErrorDialog(a.Pages, fmt.Sprintf("Delete failed: %v", err))
			}
		}
		a.refreshAllPanels()
		a.restoreFocus()
	})
}

// handleMkdir creates a new directory in the active panel's current directory.
func (a *App) handleMkdir() {
	a.DialogActive = true
	ShowInputDialog(a.Pages, a.Application, "New Directory", "Name: ", "", func(value string, cancelled bool) {
		a.DialogActive = false
		if cancelled || value == "" {
			a.restoreFocus()
			return
		}
		path := filepath.Join(a.ActivePanel.Path, value)
		err := MakeDirSafe(path)
		if err != nil {
			ShowErrorDialog(a.Pages, fmt.Sprintf("Mkdir failed: %v", err))
		}
		a.refreshAllPanels()
		a.restoreFocus()
	})
}

// handleRename renames the selected entry.
func (a *App) handleRename() {
	entry := a.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name == ".." {
		return
	}
	oldPath := filepath.Join(a.ActivePanel.Path, entry.Name)

	a.DialogActive = true
	ShowInputDialog(a.Pages, a.Application, "Rename", "New name: ", entry.Name, func(value string, cancelled bool) {
		a.DialogActive = false
		if cancelled || value == "" || value == entry.Name {
			a.restoreFocus()
			return
		}
		err := RenamePath(oldPath, value)
		if err != nil {
			ShowErrorDialog(a.Pages, fmt.Sprintf("Rename failed: %v", err))
		}
		a.refreshAllPanels()
		a.restoreFocus()
	})
}

// handleGitDiff shows the git diff for the selected file in the viewer.
func (a *App) handleGitDiff() {
	if a.GitRepo == nil {
		a.setStatusError("Not in a git repository")
		return
	}
	entry := a.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name == ".." || entry.IsDir {
		return
	}

	absPath := filepath.Join(a.ActivePanel.Path, entry.Name)
	relPath := a.GitRepo.RelPath(absPath)
	diff, err := a.GitRepo.GetDiff(relPath)
	if err != nil || diff == "" {
		a.setStatusError("No diff available for " + entry.Name)
		return
	}

	a.Viewer.Wrapper.SetTitle(fmt.Sprintf(" git diff: %s ", entry.Name))
	a.Viewer.TextView.SetText(diff)
	a.Viewer.TextView.ScrollToBeginning()
	a.ViewerActive = true
	a.Pages.SwitchToPage("viewer")
	a.Application.SetFocus(a.Viewer.TextView)
}

// handleGitStage toggles git staging for the selected file.
func (a *App) handleGitStage() {
	if a.GitRepo == nil {
		a.setStatusError("Not in a git repository")
		return
	}
	entry := a.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name == ".." {
		return
	}

	absPath := filepath.Join(a.ActivePanel.Path, entry.Name)
	relPath := a.GitRepo.RelPath(absPath)
	err := a.GitRepo.ToggleStaged(relPath)
	if err != nil {
		a.setStatusError(fmt.Sprintf("Git stage error: %v", err))
		return
	}

	a.GitRepo.Refresh()
	a.refreshAllPanels()
}

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
func (a *App) jumpToTop() {
	a.ActivePanel.Table.Select(0, 0)
	a.ActivePanel.Table.ScrollToBeginning()
}

// jumpToBottom moves cursor to the last entry.
func (a *App) jumpToBottom() {
	lastRow := len(a.ActivePanel.Entries) - 1
	if lastRow >= 0 {
		a.ActivePanel.Table.Select(lastRow, 0)
	}
}

// handleYank marks the selected entry for copy (yank).
func (a *App) handleYank() {
	entry := a.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name == ".." {
		return
	}
	path := filepath.Join(a.ActivePanel.Path, entry.Name)
	a.YankBuffer = []string{path}
	a.setStatusError(fmt.Sprintf("Yanked: %s", entry.Name))
}

// handlePaste copies yanked files to the active panel's directory.
func (a *App) handlePaste() {
	if len(a.YankBuffer) == 0 {
		a.setStatusError("Nothing yanked")
		return
	}

	for _, src := range a.YankBuffer {
		dst := filepath.Join(a.ActivePanel.Path, filepath.Base(src))
		info, err := os.Stat(src)
		if err != nil {
			a.setStatusError(fmt.Sprintf("Paste error: %v", err))
			return
		}
		if info.IsDir() {
			err = CopyDir(src, dst)
		} else {
			err = CopyFile(src, dst)
		}
		if err != nil {
			a.setStatusError(fmt.Sprintf("Paste error: %v", err))
			return
		}
	}

	a.refreshAllPanels()
	a.setStatusError(fmt.Sprintf("Pasted %d item(s)", len(a.YankBuffer)))
}

// handleBComp launches Beyond Compare with the selected files from each pane.
// copyPathToClipboard copies the selected file/folder's full path to the system clipboard.
func (a *App) copyPathToClipboard() {
	var path string
	if a.ViewMode == ViewHybridTree && a.TreeFocused {
		path = a.TreePanel.SelectedPath()
	} else {
		entry := a.ActivePanel.SelectedEntry()
		if entry == nil {
			return
		}
		path = filepath.Join(a.ActivePanel.Path, entry.Name)
	}

	err := CopyToClipboard(path)
	if err != nil {
		a.setStatusError("No clipboard tool found (install xclip)")
		return
	}
	a.setStatusError(fmt.Sprintf("Copied: %s", path))
}

func (a *App) handleBComp() {
	if !BCompAvailable() {
		a.setStatusError("bcomp not found in PATH")
		return
	}

	var left, right string
	if a.ViewMode == ViewHybridTree {
		// In hybrid mode, compare selected tree dir with right panel selection
		left = a.TreePanel.SelectedPath()
		entry := a.RightPanel.SelectedEntry()
		if entry != nil && entry.Name != ".." {
			right = filepath.Join(a.RightPanel.Path, entry.Name)
		} else {
			right = a.RightPanel.Path
		}
	} else {
		leftEntry := a.LeftPanel.SelectedEntry()
		rightEntry := a.RightPanel.SelectedEntry()
		if leftEntry != nil && leftEntry.Name != ".." {
			left = filepath.Join(a.LeftPanel.Path, leftEntry.Name)
		} else {
			left = a.LeftPanel.Path
		}
		if rightEntry != nil && rightEntry.Name != ".." {
			right = filepath.Join(a.RightPanel.Path, rightEntry.Name)
		} else {
			right = a.RightPanel.Path
		}
	}

	err := LaunchBComp(a.Application, left, right)
	if err != nil {
		a.setStatusError(fmt.Sprintf("bcomp error: %v", err))
	}
}

// handleOpenEditor opens the selected file in $EDITOR.
func (a *App) handleOpenEditor() {
	entry := a.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name == ".." || entry.IsDir {
		return
	}
	path := filepath.Join(a.ActivePanel.Path, entry.Name)

	err := OpenInEditor(a.Application, path)
	if err != nil {
		a.setStatusError(fmt.Sprintf("Editor error: %v", err))
	}
	a.refreshAllPanels()
}

// refreshAllPanels refreshes both file panels and the tree.
func (a *App) refreshAllPanels() {
	a.LeftPanel.Refresh()
	a.RightPanel.Refresh()
	if a.ViewMode == ViewHybridTree {
		a.TreePanel.Refresh()
	}
	a.updateStatusBars()
}

// restoreFocus returns focus to the appropriate widget after a dialog closes.
func (a *App) restoreFocus() {
	if a.ViewMode == ViewHybridTree && a.TreeFocused {
		a.Application.SetFocus(a.TreePanel.TreeView)
	} else {
		a.Application.SetFocus(a.ActivePanel.Table)
	}
}

// buildMenuBar creates the top-level menus with all available actions.
func (a *App) buildMenuBar() {
	mod := ModifierLabel()
	menus := []Menu{
		{
			Title:  "File",
			Hotkey: 'f',
			Items: []MenuItem{
				{Label: "New Directory", Shortcut: "F7 / n", Action: func() { a.handleMkdir() }},
				{Label: "Copy", Shortcut: "F5 / c", Action: func() { a.handleCopy() }},
				{Label: "Move", Shortcut: "F6 / m", Action: func() { a.handleMove() }},
				{Label: "Rename", Shortcut: "F2 / R", Action: func() { a.handleRename() }},
				{Label: "Delete", Shortcut: "F8 / d", Action: func() { a.handleDelete() }},
				{Label: "Quit", Shortcut: "q / Ctrl+C", Action: func() { a.Application.Stop() }},
			},
		},
		{
			Title:  "View",
			Hotkey: 'v',
			Items: []MenuItem{
				{Label: "Toggle Tree/Dual", Shortcut: "Ctrl+T", Action: func() { a.toggleViewMode() }},
				{Label: "Toggle Hidden Files", Shortcut: ".", Action: func() { a.toggleHidden() }},
				{Label: "Toggle Preview Pane", Shortcut: "t", Action: func() { a.togglePreviewPane() }},
				{Label: "Cycle Sort Mode", Shortcut: "s", Action: func() {
					a.ActivePanel.SortMode = NextSortMode(a.ActivePanel.SortMode)
					a.ActivePanel.Refresh()
					a.updateStatusBars()
				}},
				{Label: "Toggle Sort Order", Shortcut: "S", Action: func() {
					a.ActivePanel.SortOrder = ToggleSortOrder(a.ActivePanel.SortOrder)
					a.ActivePanel.Refresh()
					a.updateStatusBars()
				}},
				{Label: "Refresh", Shortcut: "r", Action: func() { a.refreshActive() }},
			},
		},
		{
			Title:  "Search",
			Hotkey: 's',
			Items: []MenuItem{
				{Label: "Filter", Shortcut: "/", Action: func() { a.enterFilterMode() }},
				{Label: "Recursive Search", Shortcut: "Ctrl+F / F3", Action: func() { a.enterSearchMode() }},
			},
		},
		{
			Title:  "Go",
			Hotkey: 'g',
			Items: []MenuItem{
				{Label: "Go to Path...", Shortcut: "Ctrl+L", Action: func() { a.showGoToPathDialog() }},
				{Label: "Jump to Home", Shortcut: "~", Action: func() { a.jumpToHome() }},
				{Label: "Jump to Root /", Shortcut: "\\", Action: func() { a.jumpToRoot() }},
				{Label: "Bookmarks...", Shortcut: "Ctrl+B", Action: func() { a.showBookmarks() }},
				{Label: "Jump to Bookmark 1", Shortcut: "1", Action: func() { a.jumpToBookmarkNum(0) }},
				{Label: "Jump to Bookmark 2", Shortcut: "2", Action: func() { a.jumpToBookmarkNum(1) }},
				{Label: "Jump to Bookmark 3", Shortcut: "3", Action: func() { a.jumpToBookmarkNum(2) }},
			},
		},
		{
			Title:  "Tools",
			Hotkey: 't',
			Items: []MenuItem{
				{Label: "Open in Editor", Shortcut: "e", Action: func() { a.handleOpenEditor() }},
				{Label: "View File", Shortcut: "Enter (on file)", Action: func() {
					entry := a.ActivePanel.SelectedEntry()
					if entry != nil && !entry.IsDir && entry.Name != ".." {
						a.openViewer(filepath.Join(a.ActivePanel.Path, entry.Name))
					}
				}},
				{Label: "Beyond Compare", Shortcut: "b", Action: func() { a.handleBComp() }},
				{Label: "Copy Path", Shortcut: mod + "+C", Action: func() { a.copyPathToClipboard() }},
				{Label: "Git Diff", Shortcut: "Ctrl+G", Action: func() { a.handleGitDiff() }},
				{Label: "Git Stage/Unstage", Shortcut: "gs", Action: func() { a.handleGitStage() }},
			},
		},
		{
			Title:  "Options",
			Hotkey: 'o',
			Items: []MenuItem{
				{Label: "Theme...", Shortcut: "", Action: func() { a.showThemeDialog() }},
				{Label: "Configuration...", Shortcut: "", Action: func() { a.showConfigDialog() }},
				{Label: "Key Bindings...", Shortcut: "", Action: func() { a.showKeybindingsDialog() }},
				{Label: "About...", Shortcut: "", Action: func() { a.showAboutDialog() }},
			},
		},
	}

	a.MenuBar = NewMenuBar(menus)
	a.MenuBar.OnClose = func() {
		a.deactivateMenuBar()
	}
}

// activateMenuBar opens the menu bar at the first menu.
func (a *App) activateMenuBar() {
	a.activateMenuAt(0)
}

// activateMenuAt opens the menu bar at a specific menu index.
func (a *App) activateMenuAt(index int) {
	if a.MenuActive {
		// Already open — just switch to the requested menu
		a.MenuBar.ActiveMenu = index
		a.MenuBar.openDropdown()
		a.Pages.RemovePage("menu-dropdown")
		a.showMenuDropdown()
		return
	}
	a.MenuActive = true
	a.MenuBar.ActiveMenu = index
	a.MenuBar.openDropdown()
	a.showMenuDropdown()
}

// showMenuDropdown positions the dropdown directly using SetRect and adds it
// as a non-resized page so the file panels remain fully visible beneath.
func (a *App) showMenuDropdown() {
	width := a.MenuBar.DropdownWidth()
	height := a.MenuBar.DropdownHeight()
	offset := a.MenuBar.DropdownOffset()

	a.MenuBar.Dropdown.SetRect(offset, 1, width, height)
	a.Pages.AddPage("menu-dropdown", a.MenuBar.Dropdown, false, true)
	a.Application.SetFocus(a.MenuBar.Dropdown)
}

// deactivateMenuBar closes the menu bar.
func (a *App) deactivateMenuBar() {
	a.MenuActive = false
	a.MenuBar.IsOpen = false
	a.MenuBar.renderBar()
	a.Pages.RemovePage("menu-dropdown")
	a.restoreFocus()
}

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

// showConfigDialog displays a configuration settings dialog.
func (a *App) showThemeDialog() {
	a.DialogActive = true
	themes := AllThemes()
	list := tview.NewList()
	list.SetBorder(true)
	list.SetTitle(" Select Theme ")
	list.SetBorderPadding(0, 0, 1, 1)
	list.SetHighlightFullLine(true)

	// Pre-select current theme
	currentIdx := 0
	for i, name := range themes {
		tc := GetTheme(name)
		current := ""
		if string(name) == a.Config.Theme {
			current = " *"
			currentIdx = i
		}
		shortcut := rune('1' + i)
		list.AddItem(tc.Name+current, string(name), shortcut, nil)
	}
	list.SetCurrentItem(currentIdx)

	closeDialog := func() {
		a.DialogActive = false
		a.Pages.RemovePage("theme-dialog")
		a.restoreFocus()
	}

	list.SetSelectedFunc(func(idx int, _ string, _ string, _ rune) {
		if idx >= 0 && idx < len(themes) {
			a.setTheme(themes[idx])
		}
		closeDialog()
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeDialog()
			return nil
		}
		return event
	})

	// Center the dialog
	width := 30
	height := len(themes)*2 + 4
	wrapper := tview.NewFlex().SetDirection(tview.FlexRow)
	wrapper.AddItem(nil, 0, 1, false)
	inner := tview.NewFlex().SetDirection(tview.FlexColumn)
	inner.AddItem(nil, 0, 1, false)
	inner.AddItem(list, width, 0, true)
	inner.AddItem(nil, 0, 1, false)
	wrapper.AddItem(inner, height, 0, true)
	wrapper.AddItem(nil, 0, 1, false)

	a.Pages.AddPage("theme-dialog", wrapper, true, true)
	a.Application.SetFocus(list)
}

func (a *App) showConfigDialog() {
	a.DialogActive = true
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle(" Configuration ")
	form.SetBorderPadding(1, 1, 2, 2)
	form.SetButtonsAlign(tview.AlignRight)

	cfg := a.Config

	closeDialog := func() {
		a.DialogActive = false
		a.Pages.RemovePage("config-dialog")
		a.restoreFocus()
	}

	form.AddCheckbox("Show hidden files by default", cfg.ShowHidden, func(checked bool) {
		cfg.ShowHidden = checked
	})
	form.AddCheckbox("Preview pane on start", cfg.PreviewOnStart, func(checked bool) {
		cfg.PreviewOnStart = checked
	})
	form.AddCheckbox("Confirm before delete", cfg.ConfirmDelete, func(checked bool) {
		cfg.ConfirmDelete = checked
	})
	form.AddCheckbox("Use trash (soft delete)", cfg.UseTrash, func(checked bool) {
		cfg.UseTrash = checked
	})

	sortOptions := []string{"name", "size", "date", "extension"}
	sortIdx := 0
	for i, s := range sortOptions {
		if s == cfg.DefaultSortMode {
			sortIdx = i
		}
	}
	form.AddDropDown("Default sort mode", sortOptions, sortIdx, func(option string, index int) {
		cfg.DefaultSortMode = option
	})

	viewOptions := []string{"hybrid", "dual"}
	viewIdx := 0
	for i, v := range viewOptions {
		if v == cfg.DefaultViewMode {
			viewIdx = i
		}
	}
	form.AddDropDown("Default view mode", viewOptions, viewIdx, func(option string, index int) {
		cfg.DefaultViewMode = option
	})

	form.AddInputField("Editor command", cfg.EditorCommand, 30, nil, func(text string) {
		cfg.EditorCommand = text
	})

	form.AddButton("Save", func() {
		a.Config = cfg
		_ = SaveConfig(cfg)
		closeDialog()
	})
	form.AddButton("Cancel", func() {
		closeDialog()
	})
	form.SetCancelFunc(func() {
		closeDialog()
	})

	// Center the dialog
	width := 60
	height := 22
	overlay := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, height, 0, true).
			AddItem(nil, 0, 1, false), width, 0, true).
		AddItem(nil, 0, 1, false)

	a.Pages.AddPage("config-dialog", overlay, true, true)
	a.Application.SetFocus(form)
}

// showKeybindingsDialog displays a reference of all keyboard shortcuts.
func (a *App) showKeybindingsDialog() {
	a.DialogActive = true
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	tv.SetBorder(true)
	tv.SetTitle(" Key Bindings ")
	tv.SetBorderPadding(1, 1, 2, 2)

	closeDialog := func() {
		a.DialogActive = false
		a.Pages.RemovePage("keybindings-dialog")
		a.restoreFocus()
	}

	mod := ModifierLabel()
	text := `[yellow]Navigation[-]
  j / Down        Move cursor down
  k / Up          Move cursor up
  h / Backspace   Navigate up / collapse tree node
  l / Enter       Navigate into / expand / open file
  gg              Jump to top
  G               Jump to bottom
  ~               Jump to $HOME
  \               Set tree root to / (full filesystem)
  Ctrl+L          Go to path...
  Tab             Switch active pane (forward)
  Shift+Tab       Switch active pane (backward)

[yellow]View[-]
  Ctrl+T   Toggle hybrid tree / dual-pane
  t        Toggle inline preview pane
  .        Toggle hidden files
  s        Cycle sort mode (name/size/date/ext)
  S        Toggle sort order (asc/desc)
  r        Refresh

[yellow]Search[-]
  /        Filter (type-to-filter)
  Ctrl+F   Recursive filename search
  F3       Recursive filename search

[yellow]File Operations[-]
  F5 / c   Copy to other pane
  F6 / m   Move to other pane
  F7 / n   New directory
  F8 / d   Delete (trash or permanent)
  F2 / R   Rename
  yy       Yank (mark for copy)
  p        Paste yanked files
  dd       Delete (vim-style)

[yellow]Tools[-]
  e        Open in $EDITOR
  b        Beyond Compare
  Ctrl+G   Git diff
  gs       Git stage/unstage

[yellow]Resize[-]
  ` + mod + `+Left/Right  Adjust horizontal split
  ` + mod + `+Up/Down     Adjust vertical split

[yellow]Menu & System[-]
  ` + mod + `+F/V/S/G/T/O  Open menu by hotkey
  F9       Open menu bar
  ` + mod + `+C     Copy path to clipboard
  Ctrl+B   Bookmarks...
  1-9      Jump to bookmark
  q        Quit
  Ctrl+C   Force quit
  Esc      Close overlay / cancel

[gray]Press Esc to close this dialog[-]`

	tv.SetText(text)

	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeDialog()
			return nil
		}
		return event
	})

	width := 55
	height := 40
	overlay := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(tv, height, 0, true).
			AddItem(nil, 0, 1, false), width, 0, true).
		AddItem(nil, 0, 1, false)

	a.Pages.AddPage("keybindings-dialog", overlay, true, true)
	a.Application.SetFocus(tv)
}

// showAboutDialog displays application info.
// showNerdFontWarning displays a warning if no Nerd Font is detected.
func (a *App) showNerdFontWarning() {
	a.DialogActive = true
	msg := "No Nerd Font detected.\n\n" +
		"File icons require a Nerd Font to display correctly.\n" +
		"Install one from: https://www.nerdfonts.com\n\n" +
		"Then set it as your terminal font.\n\n" +
		NerdFontInstallHint()

	modal := tview.NewModal().
		SetText(msg).
		AddButtons([]string{"OK", "Don't remind me"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.DialogActive = false
			if buttonLabel == "Don't remind me" {
				a.Config.NerdFontDismissed = true
				_ = SaveConfig(a.Config)
			}
			a.Pages.RemovePage("nerd-font-warning")
			a.restoreFocus()
		})
	modal.SetTitle("Nerd Font")
	modal.SetBorderColor(tcell.ColorYellow)

	a.Pages.AddPage("nerd-font-warning", modal, true, true)
}

func (a *App) showAboutDialog() {
	a.DialogActive = true
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	tv.SetBorder(true)
	tv.SetTitle(" About Twin Commander ")
	tv.SetBorderPadding(1, 1, 2, 2)

	closeDialog := func() {
		a.DialogActive = false
		a.Pages.RemovePage("about-dialog")
		a.restoreFocus()
	}

	tv.SetText(`[yellow]Twin Commander[-]

A dual-pane terminal file explorer
inspired by Norton & Midnight Commander.

Built with Go, tcell, and tview.

[gray]Keyboard-driven. No mouse. No compromise.[-]

Features: Hybrid Tree View, File Operations,
Git Integration, Vim Keys, Inline Preview,
Recursive Search, Soft Delete/Trash,
Beyond Compare, $EDITOR Integration.

[gray]Press Esc to close[-]`)

	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeDialog()
			return nil
		}
		return event
	})

	width := 50
	height := 20
	overlay := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(tv, height, 0, true).
			AddItem(nil, 0, 1, false), width, 0, true).
		AddItem(nil, 0, 1, false)

	a.Pages.AddPage("about-dialog", overlay, true, true)
	a.Application.SetFocus(tv)
}

// togglePreviewPane toggles the inline preview pane on/off.
func (a *App) togglePreviewPane() {
	if a.PreviewActive {
		// Preview is showing — escalate to full-screen viewer
		entry := a.ActivePanel.SelectedEntry()
		if entry != nil && !entry.IsDir && entry.Name != ".." {
			a.openViewer(filepath.Join(a.ActivePanel.Path, entry.Name))
		}
		return
	}

	// Open inline preview
	a.PreviewActive = true
	a.buildLayout()

	if a.ViewMode == ViewHybridTree {
		a.Pages.SwitchToPage("hybrid")
	} else {
		a.Pages.SwitchToPage("dual")
	}

	a.updatePreview()
	a.restoreFocus()
}

// closePreview hides the inline preview pane.
func (a *App) closePreview() {
	a.PreviewActive = false
	a.PreviewFocused = false
	a.PreviewWrapper.SetBorderColor(a.ActiveTheme.PanelBorderInactive)
	a.buildLayout()

	if a.ViewMode == ViewHybridTree {
		a.Pages.SwitchToPage("hybrid")
	} else {
		a.Pages.SwitchToPage("dual")
	}
	a.restoreFocus()
}

// updatePreview loads the currently selected file into the preview pane.
func (a *App) updatePreview() {
	if !a.PreviewActive {
		return
	}

	entry := a.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name == ".." {
		a.PreviewWrapper.SetTitle(" Preview ")
		a.PreviewPane.SetText("")
		return
	}

	if entry.IsDir {
		a.PreviewWrapper.SetTitle(fmt.Sprintf(" %s/ (directory) ", entry.Name))
		a.PreviewPane.SetText("")
		return
	}

	path := filepath.Join(a.ActivePanel.Path, entry.Name)

	const maxPreviewBytes = 32 * 1024 // 32KB for preview
	data, err := readFileHead(path, maxPreviewBytes)
	if err != nil {
		a.PreviewWrapper.SetTitle(fmt.Sprintf(" %s ", entry.Name))
		a.PreviewPane.SetText(fmt.Sprintf("Cannot read: %v", err))
		return
	}

	if isBinary(data) {
		a.PreviewWrapper.SetTitle(fmt.Sprintf(" %s (binary) ", entry.Name))
		a.PreviewPane.SetText("Binary file — cannot preview")
		return
	}

	// Apply syntax highlighting
	mode := DetectHighlight(path)
	highlighted := HighlightContent(string(data), mode)

	a.PreviewWrapper.SetTitle(fmt.Sprintf(" %s ", entry.Name))
	a.PreviewPane.SetText(highlighted)
	a.PreviewPane.ScrollToBeginning()
}

// InactivePanel returns the panel that is NOT active (used for file operations).
func (a *App) InactivePanel() *Panel {
	if a.ViewMode == ViewHybridTree {
		return a.RightPanel // In hybrid mode, right panel is always the file panel
	}
	if a.ActivePanel == a.LeftPanel {
		return a.RightPanel
	}
	return a.LeftPanel
}
