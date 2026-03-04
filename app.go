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

func NewApp(startPath string) *App {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine working directory: %v\n", err)
		os.Exit(1)
	}
	// Use explicit start path if provided
	if startPath != "" {
		abs, err := filepath.Abs(startPath)
		if err == nil {
			if info, serr := os.Stat(abs); serr == nil && info.IsDir() {
				cwd = abs
			}
		}
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

	// Update preview pane and visual selection when cursor moves
	app.RightPanel.Table.SetSelectionChangedFunc(func(row, column int) {
		if app.RightPanel.Selection.IsVisual() {
			app.RightPanel.Selection.UpdateVisual(row, app.RightPanel.Entries, app.RightPanel.Path)
			app.RightPanel.renderTable()
			app.updateStatusBars()
		}
		app.updatePreview()
	})
	app.LeftPanel.Table.SetSelectionChangedFunc(func(row, column int) {
		if app.LeftPanel.Selection.IsVisual() {
			app.LeftPanel.Selection.UpdateVisual(row, app.LeftPanel.Entries, app.LeftPanel.Path)
			app.LeftPanel.renderTable()
			app.updateStatusBars()
		}
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

	// Create fuzzy finder UI
	app.FuzzyInput = tview.NewInputField().
		SetLabel(" Find: ").
		SetFieldWidth(0)
	app.FuzzyTable = tview.NewTable().
		SetSelectable(true, false)
	app.FuzzyTable.SetBorder(true)
	app.FuzzyTable.SetTitle("Fuzzy Results")
	app.FuzzyTable.SetBorderPadding(0, 0, 1, 1)

	app.setupFuzzyHandlers()

	// Create directory jump UI
	app.GoDirInput = tview.NewInputField().
		SetLabel(" Dir: ").
		SetFieldWidth(0)
	app.GoDirTable = tview.NewTable().
		SetSelectable(true, false)
	app.GoDirTable.SetBorder(true)
	app.GoDirTable.SetTitle("Directory Jump")
	app.GoDirTable.SetBorderPadding(0, 0, 1, 1)

	app.setupGoDirHandlers()

	// Create inline preview pane with scrollbar (before applyConfig which themes it)
	app.PreviewPane = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	app.PreviewWrapper, _ = WrapWithScrollBar(app.PreviewPane)
	app.PreviewWrapper.SetBorder(true)
	app.PreviewWrapper.SetTitle(" Preview ")
	app.PreviewWrapper.SetBorderPadding(0, 0, 1, 1)

	// Create workspace manager
	app.WorkspaceMgr = NewWorkspaceManager()

	// Create directory size cache
	app.DirSizeCache = NewDirSizeCache(func(path string, size int64) {
		app.Application.QueueUpdateDraw(func() {
			app.updateDirSizeInTable(app.LeftPanel, path, size)
			app.updateDirSizeInTable(app.RightPanel, path, size)
		})
	})
	app.LeftPanel.DirSizeCache = app.DirSizeCache
	app.RightPanel.DirSizeCache = app.DirSizeCache

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

	// Restore session if enabled (and no explicit start path)
	if startPath == "" {
		app.loadSessionIfEnabled()
	}

	// Detect git repo
	app.GitRepo = DetectGitRepo(cwd)
	app.LeftPanel.GitRepo = app.GitRepo
	app.RightPanel.GitRepo = app.GitRepo

	app.updateStatusBars()

	// Set up global key handler
	app.Application.SetInputCapture(app.handleKeyEvent)

	// Enable mouse support — click to focus panels, select files, open menus
	app.Application.EnableMouse(true)
	app.Application.SetMouseCapture(app.handleMouseEvent)

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

func (a *App) jumpToBookmark(path string) {
	if a.AnchorActive && !a.isPathInScope(path) {
		a.setStatusError("Bookmark outside anchor scope")
		return
	}
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

// toggleAnchor locks or unlocks scope to the current directory.
func (a *App) toggleAnchor() {
	if a.AnchorActive {
		a.AnchorActive = false
		a.AnchorPath = ""
		// Restore tree root per config preference
		if a.ViewMode == ViewHybridTree {
			root := ""
			if a.Config.TreeRoot == "/" {
				root = "/"
			} else {
				if home, err := os.UserHomeDir(); err == nil {
					root = home
				} else {
					root = "/"
				}
			}
			a.TreePanel.SetRootPath(root)
		}
		a.setStatusError("Anchor released")
	} else {
		// Determine current directory
		dir := a.ActivePanel.Path
		if a.ViewMode == ViewHybridTree && a.TreeFocused {
			dir = a.TreePanel.SelectedPath()
		}
		a.AnchorPath = filepath.Clean(dir)
		a.AnchorActive = true
		// Re-root tree to anchor path in hybrid mode
		if a.ViewMode == ViewHybridTree {
			a.TreePanel.SetRootPath(a.AnchorPath)
		}
		a.setStatusError("⚓ Anchored to " + a.AnchorPath)
	}
	a.updateStatusBars()
}

// anchoredRoot returns the search root directory, respecting anchor scope.
func (a *App) anchoredRoot() string {
	if a.AnchorActive {
		return a.AnchorPath
	}
	if a.ViewMode == ViewHybridTree {
		return a.TreePanel.RootPath
	}
	return a.ActivePanel.Path
}

// isPathInScope returns true if the path is within the anchor scope (or anchor is inactive).
func (a *App) isPathInScope(path string) bool {
	if !a.AnchorActive {
		return true
	}
	return strings.HasPrefix(filepath.Clean(path), filepath.Clean(a.AnchorPath))
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

	// Theme tab bar
	if a.WorkspaceMgr != nil {
		a.WorkspaceMgr.TabBar.SetBackgroundColor(tc.MenuBarBg)
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
	a.Pages.RemovePage("fuzzy")

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
		AddItem(a.MenuBar.Bar, 1, 0, false)
	if a.WorkspaceMgr != nil && a.WorkspaceMgr.Count() > 1 {
		hybridRoot.AddItem(a.WorkspaceMgr.TabBar, 1, 0, false)
	}
	hybridRoot.AddItem(hybridFlex, 0, 1, true).
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
		AddItem(a.MenuBar.Bar, 1, 0, false)
	if a.WorkspaceMgr != nil && a.WorkspaceMgr.Count() > 1 {
		a.RootFlex.AddItem(a.WorkspaceMgr.TabBar, 1, 0, false)
	}
	a.RootFlex.AddItem(dualFlex, 0, 1, true).
		AddItem(a.FilterInput, 0, 0, false)

	// Search overlay
	searchFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.SearchInput, 1, 0, true).
		AddItem(a.SearchTable, 0, 1, false)

	// Fuzzy finder overlay
	fuzzyFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.FuzzyInput, 1, 0, true).
		AddItem(a.FuzzyTable, 0, 1, false)

	// Directory jump overlay
	godirFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.GoDirInput, 1, 0, true).
		AddItem(a.GoDirTable, 0, 1, false)

	// Viewer layout: menu bar on top, viewer below
	viewerRoot := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.MenuBar.Bar, 1, 0, false).
		AddItem(a.Viewer.Wrapper, 0, 1, true)

	isHybrid := a.ViewMode == ViewHybridTree
	a.Pages.AddPage("hybrid", hybridRoot, true, isHybrid)
	a.Pages.AddPage("dual", a.RootFlex, true, !isHybrid)
	a.Pages.AddPage("viewer", viewerRoot, true, false)
	a.Pages.AddPage("search", searchFlex, true, false)
	a.Pages.AddPage("fuzzy", fuzzyFlex, true, false)
	a.Pages.AddPage("godir", godirFlex, true, false)
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
		Selection:           NewSelection(),
		History:             NewHistory(100),
	}
}

// Run starts the application.
func (a *App) Run() error {
	return a.Application.Run()
}

// saveSessionOnQuit persists the current workspace state before exit.
func (a *App) saveSessionOnQuit() {
	if !a.Config.SessionRestore {
		return
	}
	a.saveWorkspaceState()
	_ = SaveSession(a.WorkspaceMgr.Workspaces, a.WorkspaceMgr.Active)
}

// loadSessionIfEnabled restores workspaces from the previous session.
func (a *App) loadSessionIfEnabled() {
	if !a.Config.SessionRestore {
		return
	}
	session, err := LoadSession()
	if err != nil || session == nil || len(session.Workspaces) == 0 {
		return
	}

	// Replace default workspace(s) with saved ones
	a.WorkspaceMgr.Workspaces = make([]*Workspace, len(session.Workspaces))
	for i := range session.Workspaces {
		ws := session.Workspaces[i]
		a.WorkspaceMgr.Workspaces[i] = &ws
	}
	a.WorkspaceMgr.Active = session.ActiveIndex
	a.WorkspaceMgr.renderTabBar()

	// Restore the active workspace state
	a.restoreWorkspaceState()
}

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

	rootDir := a.anchoredRoot()

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
	// Anchor guard: don't navigate above anchor path
	if a.AnchorActive {
		if a.ViewMode == ViewHybridTree && a.TreeFocused {
			if filepath.Clean(a.TreePanel.SelectedPath()) == filepath.Clean(a.AnchorPath) {
				return
			}
		} else if filepath.Clean(a.ActivePanel.Path) == filepath.Clean(a.AnchorPath) {
			return
		}
	}

	if a.ViewMode == ViewHybridTree && a.TreeFocused {
		if a.TreePanel.SelectedIsExpanded() {
			a.TreePanel.CollapseSelected()
		} else {
			a.TreePanel.MoveToParent()
		}
		// Sync right panel to tree's current selection
		a.syncRightPanelToTree()
	} else {
		a.ActivePanel.NavigateUp()
		a.syncTreeToRightPanel()
	}
	a.updateStatusBars()
}

// jumpToHome navigates to $HOME, preserving tree state.
func (a *App) jumpToHome() {
	target := ""
	if a.AnchorActive {
		target = a.AnchorPath
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}
		target = home
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
}

// jumpToRoot navigates to "/" (filesystem root).
func (a *App) jumpToRoot() {
	target := "/"
	if a.AnchorActive {
		target = a.AnchorPath
	}
	if a.ViewMode == ViewHybridTree {
		a.TreePanel.SetRootPath(target)
		a.TreePanel.ExpandToPath(target)
		a.RightPanel.Path = target
		a.RightPanel.LoadDir()
	} else {
		a.ActivePanel.Path = target
		a.ActivePanel.LoadDir()
	}
	a.updateStatusBars()
}

func (a *App) handleEnter() {
	if a.ViewMode == ViewHybridTree && a.TreeFocused {
		// Trigger the tree's selected func (expand dir / open file)
		node := a.TreePanel.TreeView.GetCurrentNode()
		if node != nil {
			ref := node.GetReference()
			if ref != nil {
				path := ref.(string)
				info, err := os.Stat(path)
				if err == nil {
					if info.IsDir() {
						a.TreePanel.ToggleExpand(node, path)
						a.syncRightPanelToTree()
					} else if a.TreePanel.OnFileSelect != nil {
						a.TreePanel.OnFileSelect(path)
					}
				}
			}
		}
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
	a.syncTreeToRightPanel()
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
	prefix := ""
	if a.AnchorActive {
		prefix = "⚓ "
	}
	if a.GitRepo != nil && a.GitRepo.Branch != "" {
		prefix += "[" + a.GitRepo.Branch + "] "
	}

	if a.ViewMode == ViewHybridTree {
		a.LeftStatus.SetText(prefix + a.TreePanel.SelectedPath())
		a.RightStatus.SetText(a.RightPanel.StatusText())
	} else {
		a.LeftStatus.SetText(prefix + a.LeftPanel.StatusText())
		a.RightStatus.SetText(a.RightPanel.StatusText())
	}
}

// setStatusError displays a message in the active panel's status bar.
// The message auto-clears after 3 seconds, restoring the normal status text.
func (a *App) setStatusError(msg string) {
	if a.ViewMode == ViewHybridTree {
		a.RightStatus.SetText(msg)
	} else if a.ActivePanel == a.LeftPanel {
		a.LeftStatus.SetText(msg)
	} else {
		a.RightStatus.SetText(msg)
	}
	a.ActivePanel.StatusBar = msg

	// Cancel any existing auto-clear timer
	if a.statusTimer != nil {
		a.statusTimer.Stop()
	}

	// Auto-clear after 3 seconds
	a.statusTimer = time.AfterFunc(3*time.Second, func() {
		a.Application.QueueUpdateDraw(func() {
			a.updateStatusBars()
		})
	})
}

func (a *App) jumpToTop() {
	if a.ViewMode == ViewHybridTree && a.TreeFocused {
		root := a.TreePanel.TreeView.GetRoot()
		if root != nil {
			a.TreePanel.TreeView.SetCurrentNode(root)
		}
		a.syncRightPanelToTree()
		return
	}
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

func (a *App) syncTreeToRightPanel() {
	if a.ViewMode != ViewHybridTree {
		return
	}
	a.TreePanel.NavigateToPath(a.RightPanel.Path)
	a.updateStatusBars()
}

// syncRightPanelToTree syncs the right panel to the tree's currently selected directory.
func (a *App) syncRightPanelToTree() {
	if a.ViewMode != ViewHybridTree {
		return
	}
	path := a.TreePanel.SelectedPath()
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if info.IsDir() {
		if path != a.RightPanel.Path {
			a.RightPanel.Path = path
			a.RightPanel.LoadDir()
			a.updateStatusBars()
		}
	} else {
		dir := filepath.Dir(path)
		if dir != a.RightPanel.Path {
			a.RightPanel.Path = dir
			a.RightPanel.LoadDir()
		}
		baseName := filepath.Base(path)
		for i, e := range a.RightPanel.Entries {
			if e.Name == baseName {
				a.RightPanel.Table.Select(i, 0)
				break
			}
		}
		a.updatePreview()
		a.updateStatusBars()
	}
}

// refreshAllPanels refreshes both file panels and the tree.
// Also invalidates directory size cache so sizes are recalculated.
func (a *App) refreshAllPanels() {
	a.invalidateDirSizes()
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

// focusTree switches focus to the tree panel (hybrid mode).
func (a *App) focusTree() {
	a.TreeFocused = true
	a.PreviewFocused = false
	a.TreePanel.TreeView.SetBorderColor(a.ActiveTheme.PanelBorderActive)
	a.RightPanel.SetActive(false)
	a.LeftPanel.SetActive(false)
	a.PreviewWrapper.SetBorderColor(a.ActiveTheme.PanelBorderInactive)
	a.Application.SetFocus(a.TreePanel.TreeView)
}

// focusLeftPanel switches focus to the left panel (dual-pane mode).
func (a *App) focusLeftPanel() {
	a.TreeFocused = false
	a.PreviewFocused = false
	a.ActivePanel = a.LeftPanel
	a.LeftPanel.SetActive(true)
	a.RightPanel.SetActive(false)
	a.TreePanel.TreeView.SetBorderColor(a.ActiveTheme.PanelBorderInactive)
	a.PreviewWrapper.SetBorderColor(a.ActiveTheme.PanelBorderInactive)
	a.Application.SetFocus(a.LeftPanel.Table)
}

// focusRightPanel switches focus to the right panel.
func (a *App) focusRightPanel() {
	a.TreeFocused = false
	a.PreviewFocused = false
	a.ActivePanel = a.RightPanel
	a.RightPanel.SetActive(true)
	a.LeftPanel.SetActive(false)
	a.TreePanel.TreeView.SetBorderColor(a.ActiveTheme.PanelBorderInactive)
	a.PreviewWrapper.SetBorderColor(a.ActiveTheme.PanelBorderInactive)
	a.Application.SetFocus(a.RightPanel.Table)
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
				{Label: "New File", Shortcut: "N", Action: func() { a.handleMkfile() }},
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
				{Label: "New Workspace", Shortcut: "Ctrl+N", Action: func() { a.createWorkspace() }},
				{Label: "Close Workspace", Shortcut: "Ctrl+W", Action: func() { a.closeWorkspace() }},
			},
		},
		{
			Title:  "Search",
			Hotkey: 's',
			Items: []MenuItem{
				{Label: "Filter (glob/regex)", Shortcut: "/", Action: func() { a.enterFilterMode() }},
				{Label: "Recursive Search", Shortcut: "Ctrl+F / F3", Action: func() { a.enterSearchMode() }},
				{Label: "Content Search", Shortcut: "Ctrl+/", Action: func() { a.enterContentSearch() }},
				{Label: "Fuzzy Finder", Shortcut: "Ctrl+P", Action: func() { a.enterFuzzyMode() }},
			},
		},
		{
			Title:  "Go",
			Hotkey: 'g',
			Items: []MenuItem{
				{Label: "Anchor", Shortcut: "a", Action: func() { a.toggleAnchor() }},
				{Label: "Go to Path...", Shortcut: "Ctrl+L", Action: func() { a.showGoToPathDialog() }},
				{Label: "Directory Jump", Shortcut: "gd", Action: func() { a.enterGoDirMode() }},
				{Label: "Jump to Home", Shortcut: "~", Action: func() { a.jumpToHome() }},
				{Label: "Jump to Root /", Shortcut: "\\", Action: func() { a.jumpToRoot() }},
				{Label: "History Back", Shortcut: "-", Action: func() { a.handleHistoryBack() }},
				{Label: "History Forward", Shortcut: "=", Action: func() { a.handleHistoryForward() }},
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
				{Label: "Open with Default", Shortcut: "o", Action: func() { a.handleOpenDefault() }},
				{Label: "View File", Shortcut: "Enter (on file)", Action: func() {
					entry := a.ActivePanel.SelectedEntry()
					if entry != nil && !entry.IsDir && entry.Name != ".." {
						a.openViewer(filepath.Join(a.ActivePanel.Path, entry.Name))
					}
				}},
				{Label: "Shell Command", Shortcut: ":", Action: func() { a.enterCommandMode() }},
				{Label: "Change Permissions", Shortcut: "", Action: func() { a.handleChmod() }},
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

	// Set up mouse click handler for dropdown items
	menuIdx := a.MenuBar.ActiveMenu
	a.MenuBar.Dropdown.SetSelectedFunc(func(idx int, _ string, _ string, _ rune) {
		if menuIdx >= 0 && menuIdx < len(a.MenuBar.Menus) {
			items := a.MenuBar.Menus[menuIdx].Items
			if idx >= 0 && idx < len(items) && !items[idx].Disabled {
				a.deactivateMenuBar()
				if items[idx].Action != nil {
					items[idx].Action()
				}
			}
		}
	})

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

// --- Directory size cache ---

// updateDirSizeInTable updates the size column for a directory entry when its
// size calculation completes asynchronously.
func (a *App) updateDirSizeInTable(panel *Panel, dirPath string, size int64) {
	dirName := filepath.Base(dirPath)
	parentDir := filepath.Dir(dirPath)
	if panel.Path != parentDir {
		return
	}
	for i, e := range panel.Entries {
		if e.IsDir && e.Name == dirName {
			panel.Table.GetCell(i, 1).SetText(FormatSize(size))
			break
		}
	}
}

// --- Workspace management ---

// saveWorkspaceState captures the current app state into the active workspace.
func (a *App) saveWorkspaceState() {
	ws := a.WorkspaceMgr.Current()
	if ws == nil {
		return
	}
	ws.LeftPath = a.LeftPanel.Path
	ws.LeftShowHidden = a.LeftPanel.ShowHidden
	ws.LeftSortMode = a.LeftPanel.SortMode
	ws.LeftSortOrder = a.LeftPanel.SortOrder
	ws.RightPath = a.RightPanel.Path
	ws.RightShowHidden = a.RightPanel.ShowHidden
	ws.RightSortMode = a.RightPanel.SortMode
	ws.RightSortOrder = a.RightPanel.SortOrder
	ws.ViewMode = a.ViewMode
	ws.TreeFocused = a.TreeFocused
	ws.PreviewActive = a.PreviewActive
	ws.ActiveIsLeft = (a.ActivePanel == a.LeftPanel)
	ws.HSplit = a.HSplit
	ws.VSplit = a.VSplit
	ws.TreeRootPath = a.TreePanel.RootPath
	ws.TreeExpandedPaths = a.TreePanel.ExpandedPaths()
	ws.AnchorPath = a.AnchorPath
	ws.AnchorActive = a.AnchorActive
}

// restoreWorkspaceState applies a saved workspace state to the app.
func (a *App) restoreWorkspaceState() {
	ws := a.WorkspaceMgr.Current()
	if ws == nil {
		return
	}

	// Restore panel paths and settings
	a.LeftPanel.Path = ws.LeftPath
	a.LeftPanel.ShowHidden = ws.LeftShowHidden
	a.LeftPanel.SortMode = ws.LeftSortMode
	a.LeftPanel.SortOrder = ws.LeftSortOrder
	a.RightPanel.Path = ws.RightPath
	a.RightPanel.ShowHidden = ws.RightShowHidden
	a.RightPanel.SortMode = ws.RightSortMode
	a.RightPanel.SortOrder = ws.RightSortOrder

	// Restore view mode and splits
	a.ViewMode = ws.ViewMode
	a.TreeFocused = ws.TreeFocused
	a.PreviewActive = ws.PreviewActive
	a.HSplit = ws.HSplit
	a.VSplit = ws.VSplit
	a.AnchorPath = ws.AnchorPath
	a.AnchorActive = ws.AnchorActive

	// Restore tree state
	if ws.TreeRootPath != "" {
		a.TreePanel.SetRootPath(ws.TreeRootPath)
		if ws.TreeExpandedPaths != nil {
			for path := range ws.TreeExpandedPaths {
				a.TreePanel.ExpandToPath(path)
			}
		}
	}

	// Reload directories
	a.LeftPanel.LoadDir()
	a.RightPanel.LoadDir()

	// Restore active panel
	if ws.ActiveIsLeft {
		a.ActivePanel = a.LeftPanel
		a.LeftPanel.SetActive(true)
		a.RightPanel.SetActive(false)
	} else {
		a.ActivePanel = a.RightPanel
		a.RightPanel.SetActive(true)
		a.LeftPanel.SetActive(false)
	}

	// Rebuild layout and switch to correct page
	a.buildLayout()
	if a.ViewMode == ViewHybridTree {
		a.Pages.SwitchToPage("hybrid")
	} else {
		a.Pages.SwitchToPage("dual")
	}
	a.restoreFocus()
	a.updateTabBarVisibility()
	a.updateStatusBars()
}

// switchWorkspace saves current state, switches to the target workspace, and restores.
func (a *App) switchWorkspace(index int) {
	if index < 0 || index >= a.WorkspaceMgr.Count() {
		return
	}
	if index == a.WorkspaceMgr.Active {
		return
	}
	a.saveWorkspaceState()
	a.WorkspaceMgr.Active = index
	a.WorkspaceMgr.renderTabBar()
	a.restoreWorkspaceState()
}

// createWorkspace saves current state, creates a new workspace, and switches to it.
func (a *App) createWorkspace() {
	a.saveWorkspaceState()

	// New workspace inherits current paths
	idx := a.WorkspaceMgr.AddWorkspace("")
	ws := a.WorkspaceMgr.Workspaces[idx]
	ws.LeftPath = a.LeftPanel.Path
	ws.RightPath = a.RightPanel.Path
	ws.ViewMode = a.ViewMode
	ws.TreeFocused = a.TreeFocused
	ws.HSplit = a.HSplit
	ws.VSplit = a.VSplit
	ws.TreeRootPath = a.TreePanel.RootPath
	ws.TreeExpandedPaths = a.TreePanel.ExpandedPaths()
	ws.AnchorPath = a.AnchorPath
	ws.AnchorActive = a.AnchorActive

	a.WorkspaceMgr.Active = idx
	a.WorkspaceMgr.renderTabBar()
	a.updateTabBarVisibility()
	a.updateStatusBars()
}

// closeWorkspace removes the current workspace.
func (a *App) closeWorkspace() {
	if a.WorkspaceMgr.Count() <= 1 {
		a.setStatusError("Cannot close last workspace")
		return
	}
	ok := a.WorkspaceMgr.RemoveWorkspace(a.WorkspaceMgr.Active)
	if !ok {
		return
	}
	a.restoreWorkspaceState()
}

// updateTabBarVisibility shows the tab bar only when there are multiple workspaces.
func (a *App) updateTabBarVisibility() {
	// Rebuild layout to include/exclude tab bar
	a.buildLayout()
	if a.ViewMode == ViewHybridTree {
		a.Pages.SwitchToPage("hybrid")
	} else {
		a.Pages.SwitchToPage("dual")
	}
}

// invalidateDirSizes clears the directory size cache after file operations.
func (a *App) invalidateDirSizes() {
	if a.DirSizeCache != nil {
		a.DirSizeCache.InvalidateAll()
	}
}

// --- Fuzzy finder ---

// setupFuzzyHandlers configures the fuzzy input with debounced search.
func (a *App) setupFuzzyHandlers() {
	var debounceTimer *time.Timer

	a.FuzzyInput.SetChangedFunc(func(text string) {
		if debounceTimer != nil {
			debounceTimer.Stop()
		}
		debounceTimer = time.AfterFunc(150*time.Millisecond, func() {
			a.Application.QueueUpdateDraw(func() {
				a.runFuzzySearch(text)
			})
		})
	})
}

// enterFuzzyMode activates the fuzzy finder overlay.
func (a *App) enterFuzzyMode() {
	a.FuzzyMode = true
	a.FuzzyInput.SetText("")
	a.FuzzyTable.Clear()
	a.Pages.SwitchToPage("fuzzy")
	a.Application.SetFocus(a.FuzzyInput)
}

// exitFuzzyMode returns from fuzzy finder to the previous view.
func (a *App) exitFuzzyMode() {
	a.FuzzyMode = false
	if a.FuzzyCancel != nil {
		close(a.FuzzyCancel)
		a.FuzzyCancel = nil
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

// runFuzzySearch executes a fuzzy search with the given query.
func (a *App) runFuzzySearch(query string) {
	// Cancel previous search
	if a.FuzzyCancel != nil {
		close(a.FuzzyCancel)
	}
	a.FuzzyCancel = make(chan struct{})
	cancelCh := a.FuzzyCancel

	a.FuzzyTable.Clear()

	if query == "" {
		return
	}

	rootDir := a.anchoredRoot()

	resultCh := make(chan FuzzyResult, 100)
	go FuzzySearch(FuzzySearchOpts{
		RootDir:    rootDir,
		Pattern:    query,
		MaxResults: 200,
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
				a.FuzzyTable.SetCell(rowIdx, 0,
					tview.NewTableCell(r.RelPath).
						SetStyle(style).
						SetReference(r.Path))
			})
			row++
		}
	}()
}

// navigateToFuzzyResult opens the selected fuzzy search result.
func (a *App) navigateToFuzzyResult() {
	row, _ := a.FuzzyTable.GetSelection()
	cell := a.FuzzyTable.GetCell(row, 0)
	if cell == nil {
		return
	}
	ref := cell.GetReference()
	if ref == nil {
		return
	}
	fullPath := ref.(string)

	// Determine target directory and file name
	info, err := os.Stat(fullPath)
	if err != nil {
		a.exitFuzzyMode()
		return
	}

	var dir, baseName string
	if info.IsDir() {
		dir = fullPath
		baseName = ""
	} else {
		dir = filepath.Dir(fullPath)
		baseName = filepath.Base(fullPath)
	}

	a.exitFuzzyMode()

	// Navigate the active panel to the directory
	if a.ViewMode == ViewHybridTree {
		a.TreePanel.NavigateToPath(dir)
		a.RightPanel.Path = dir
		a.RightPanel.LoadDir()
	} else {
		a.ActivePanel.Path = dir
		a.ActivePanel.LoadDir()
	}

	// Select the file if applicable
	if baseName != "" {
		panel := a.ActivePanel
		if a.ViewMode == ViewHybridTree {
			panel = a.RightPanel
		}
		for i, e := range panel.Entries {
			if e.Name == baseName {
				panel.Table.Select(i, 0)
				break
			}
		}
	}

	a.updateStatusBars()
}

// setupGoDirHandlers configures the directory jump input with debounced search.
func (a *App) setupGoDirHandlers() {
	var debounceTimer *time.Timer

	a.GoDirInput.SetChangedFunc(func(text string) {
		if debounceTimer != nil {
			debounceTimer.Stop()
		}
		debounceTimer = time.AfterFunc(150*time.Millisecond, func() {
			a.Application.QueueUpdateDraw(func() {
				a.runGoDirSearch(text)
			})
		})
	})
}

// enterGoDirMode activates the directory jump overlay.
func (a *App) enterGoDirMode() {
	a.GoDirMode = true
	a.GoDirInput.SetText("")
	a.GoDirTable.Clear()
	a.Pages.SwitchToPage("godir")
	a.Application.SetFocus(a.GoDirInput)
}

// exitGoDirMode returns from directory jump to the previous view.
func (a *App) exitGoDirMode() {
	a.GoDirMode = false
	if a.GoDirCancel != nil {
		close(a.GoDirCancel)
		a.GoDirCancel = nil
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

// runGoDirSearch executes a directory-only fuzzy search with the given query.
func (a *App) runGoDirSearch(query string) {
	if a.GoDirCancel != nil {
		close(a.GoDirCancel)
	}
	a.GoDirCancel = make(chan struct{})
	cancelCh := a.GoDirCancel

	a.GoDirTable.Clear()

	if query == "" {
		return
	}

	rootDir := a.anchoredRoot()

	resultCh := make(chan FuzzyResult, 100)
	go FuzzySearch(FuzzySearchOpts{
		RootDir:    rootDir,
		Pattern:    query,
		MaxResults: 200,
		ShowHidden: a.ActivePanel.ShowHidden,
		DirsOnly:   true,
	}, resultCh, cancelCh)

	go func() {
		row := 0
		for result := range resultCh {
			r := result
			rowIdx := row
			a.Application.QueueUpdateDraw(func() {
				style := tcell.StyleDefault.Foreground(tcell.ColorBlue).Bold(true)
				a.GoDirTable.SetCell(rowIdx, 0,
					tview.NewTableCell(r.RelPath).
						SetStyle(style).
						SetReference(r.Path))
			})
			row++
		}
	}()
}

// navigateToGoDirResult opens the selected directory from the directory jump results.
func (a *App) navigateToGoDirResult() {
	row, _ := a.GoDirTable.GetSelection()
	cell := a.GoDirTable.GetCell(row, 0)
	if cell == nil {
		return
	}
	ref := cell.GetReference()
	if ref == nil {
		return
	}
	fullPath := ref.(string)

	a.exitGoDirMode()

	if a.ViewMode == ViewHybridTree {
		a.TreePanel.NavigateToPath(fullPath)
		a.RightPanel.Path = fullPath
		a.RightPanel.LoadDir()
	} else {
		a.ActivePanel.Path = fullPath
		a.ActivePanel.LoadDir()
	}

	a.updateStatusBars()
}

// --- Selection handlers ---

