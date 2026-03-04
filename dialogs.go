package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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

		// Anchor guard: prevent navigation outside anchor scope
		if a.AnchorActive && !a.isPathInScope(target) {
			a.setStatusError("Path outside anchor scope")
			a.restoreFocus()
			return
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
	form.AddCheckbox("Restore session on start", cfg.SessionRestore, func(checked bool) {
		cfg.SessionRestore = checked
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
	height := 24
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
  ~               Jump to $HOME (preserves tree state)
  \               Jump to / (works in all modes)
  a               Anchor (scope lock)
  -               History back
  =               History forward
  Ctrl+L          Go to path...
  gd              Directory jump (fuzzy)
  gr              Recent directories
  Tab             Switch active pane (forward)
  Shift+Tab       Switch active pane (backward)

[yellow]View[-]
  Ctrl+T   Toggle hybrid tree / dual-pane
  t        Toggle inline preview pane
  .        Toggle hidden files
  s        Cycle sort mode (name/size/date/ext)
  S        Toggle sort order (asc/desc)
  r        Refresh

[yellow]Selection[-]
  Space    Toggle select + move down
  v        Start visual selection
  V/Esc    End visual selection
  *        Invert selection
  +        Select by pattern

[yellow]Search & Filter[-]
  /        Filter (supports glob *.go, regex /pat/)
  Ctrl+F   Recursive filename search
  F3       Recursive filename search
  Ctrl+/   Content search (grep)
  Ctrl+P   Fuzzy finder

[yellow]File Operations[-]
  F5 / c   Copy to other pane (multi-select aware)
  F6 / m   Move to other pane (multi-select aware)
  F7 / n   New directory
  N        New file
  F8 / d   Delete (trash or permanent, multi-select)
  F2 / R   Rename
  %%        Bulk rename (find/replace in selected filenames)
  L        Create symlink to selected entry
  yy       Yank (mark for copy, multi-select)
  p        Paste yanked files
  dd       Delete (vim-style)

[yellow]Tools[-]
  e        Open in $EDITOR (respects config)
  o        Open with system default (xdg-open)
  b        Beyond Compare
  Ctrl+D   File diff (compare across panels)
  D        Disk usage (size breakdown)
  :        Run shell command (%f=file, %d=dir, %s=selected)
  Ctrl+G   Git diff
  gs       Git stage/unstage

[yellow]Resize[-]
  ` + mod + `+Left/Right  Adjust horizontal split
  ` + mod + `+Up/Down     Adjust vertical split

[yellow]Workspaces[-]
  Ctrl+N   New workspace
  Ctrl+W   Close workspace
  ` + mod + `+1-9     Switch to workspace

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

// showFileDiff compares two files and displays a unified diff.
func (a *App) showFileDiff() {
	// Determine the two files to compare
	var leftPath, rightPath string

	if a.ViewMode == ViewHybridTree {
		// In hybrid mode, can't easily pick two files — use right panel selected + prompt
		a.setStatusError("Diff requires dual-pane mode (two panels)")
		return
	}

	leftEntry := a.LeftPanel.SelectedEntry()
	rightEntry := a.RightPanel.SelectedEntry()

	if leftEntry == nil || rightEntry == nil || leftEntry.IsDir || rightEntry.IsDir ||
		leftEntry.Name == ".." || rightEntry.Name == ".." {
		a.setStatusError("Select a file in each panel to diff")
		return
	}

	leftPath = filepath.Join(a.LeftPanel.Path, leftEntry.Name)
	rightPath = filepath.Join(a.RightPanel.Path, rightEntry.Name)

	// Read both files
	leftData, err := os.ReadFile(leftPath)
	if err != nil {
		a.setStatusError(fmt.Sprintf("Cannot read %s: %v", leftEntry.Name, err))
		return
	}
	rightData, err := os.ReadFile(rightPath)
	if err != nil {
		a.setStatusError(fmt.Sprintf("Cannot read %s: %v", rightEntry.Name, err))
		return
	}

	// Check for binary
	if isBinaryContent(leftData) || isBinaryContent(rightData) {
		a.setStatusError("Cannot diff binary files")
		return
	}

	leftLines := strings.Split(string(leftData), "\n")
	rightLines := strings.Split(string(rightData), "\n")

	diff := unifiedDiff(leftPath, rightPath, leftLines, rightLines)
	if diff == "" {
		diff = "Files are identical."
	}

	a.DialogActive = true
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false)
	tv.SetBorder(true)
	tv.SetTitle(fmt.Sprintf(" Diff: %s ↔ %s ", leftEntry.Name, rightEntry.Name))
	tv.SetBorderPadding(0, 0, 1, 1)

	// Colorize diff output
	var colored strings.Builder
	for _, line := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++"):
			fmt.Fprintf(&colored, "[yellow]%s[-]\n", tview.Escape(line))
		case strings.HasPrefix(line, "@@"):
			fmt.Fprintf(&colored, "[aqua]%s[-]\n", tview.Escape(line))
		case strings.HasPrefix(line, "-"):
			fmt.Fprintf(&colored, "[red]%s[-]\n", tview.Escape(line))
		case strings.HasPrefix(line, "+"):
			fmt.Fprintf(&colored, "[green]%s[-]\n", tview.Escape(line))
		default:
			fmt.Fprintf(&colored, "%s\n", tview.Escape(line))
		}
	}
	tv.SetText(colored.String())

	closeDialog := func() {
		a.DialogActive = false
		a.Pages.RemovePage("file-diff")
		a.restoreFocus()
	}

	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			closeDialog()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				closeDialog()
				return nil
			case 'j':
				row, col := tv.GetScrollOffset()
				tv.ScrollTo(row+1, col)
				return nil
			case 'k':
				row, col := tv.GetScrollOffset()
				if row > 0 {
					tv.ScrollTo(row-1, col)
				}
				return nil
			case 'G':
				tv.ScrollToEnd()
				return nil
			case 'g':
				tv.ScrollToBeginning()
				return nil
			}
		}
		return event
	})

	width := 80
	height := 30
	overlay := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(tv, height, 0, true).
			AddItem(nil, 0, 1, false), width, 0, true).
		AddItem(nil, 0, 1, false)

	a.Pages.AddPage("file-diff", overlay, true, true)
	a.Application.SetFocus(tv)
}

// isBinaryContent checks if data appears to be binary (contains null bytes).
func isBinaryContent(data []byte) bool {
	limit := 8000
	if len(data) < limit {
		limit = len(data)
	}
	for i := 0; i < limit; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}

// unifiedDiff produces a simple unified diff between two sets of lines.
func unifiedDiff(leftName, rightName string, left, right []string) string {
	// Simple LCS-based diff
	m, n := len(left), len(right)

	// Build LCS table (memory-efficient for reasonable file sizes)
	if m > 10000 || n > 10000 {
		return "Files too large for inline diff (>10000 lines)"
	}

	lcs := make([][]int, m+1)
	for i := range lcs {
		lcs[i] = make([]int, n+1)
	}
	for i := m - 1; i >= 0; i-- {
		for j := n - 1; j >= 0; j-- {
			if left[i] == right[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}

	// Generate diff hunks
	type diffLine struct {
		op   byte // ' ', '+', '-'
		text string
		oldN int
		newN int
	}
	var lines []diffLine
	i, j := 0, 0
	oldLine, newLine := 1, 1
	for i < m || j < n {
		if i < m && j < n && left[i] == right[j] {
			lines = append(lines, diffLine{' ', left[i], oldLine, newLine})
			i++
			j++
			oldLine++
			newLine++
		} else if j < n && (i >= m || lcs[i][j+1] >= lcs[i+1][j]) {
			lines = append(lines, diffLine{'+', right[j], 0, newLine})
			j++
			newLine++
		} else if i < m {
			lines = append(lines, diffLine{'-', left[i], oldLine, 0})
			i++
			oldLine++
		}
	}

	// Group into hunks with context
	const contextLines = 3
	var result strings.Builder
	fmt.Fprintf(&result, "--- %s\n+++ %s\n", leftName, rightName)

	// Find changed regions
	type hunk struct {
		start, end int
	}
	var hunks []hunk
	inChange := false
	changeStart := 0
	for idx, l := range lines {
		if l.op != ' ' {
			if !inChange {
				changeStart = idx
				inChange = true
			}
		} else {
			if inChange {
				hunks = append(hunks, hunk{changeStart, idx})
				inChange = false
			}
		}
	}
	if inChange {
		hunks = append(hunks, hunk{changeStart, len(lines)})
	}

	// Merge overlapping hunks and add context
	for _, h := range hunks {
		start := h.start - contextLines
		if start < 0 {
			start = 0
		}
		end := h.end + contextLines
		if end > len(lines) {
			end = len(lines)
		}

		// Find old/new line numbers at start
		oldStart, newStart := 1, 1
		for idx := 0; idx < start; idx++ {
			if lines[idx].op != '+' {
				oldStart++
			}
			if lines[idx].op != '-' {
				newStart++
			}
		}

		oldCount, newCount := 0, 0
		for idx := start; idx < end; idx++ {
			if lines[idx].op != '+' {
				oldCount++
			}
			if lines[idx].op != '-' {
				newCount++
			}
		}

		fmt.Fprintf(&result, "@@ -%d,%d +%d,%d @@\n", oldStart, oldCount, newStart, newCount)
		for idx := start; idx < end; idx++ {
			fmt.Fprintf(&result, "%c%s\n", lines[idx].op, lines[idx].text)
		}
	}

	return result.String()
}

// showDiskUsage displays a sorted breakdown of subdirectory sizes.
func (a *App) showDiskUsage() {
	dir := a.ActivePanel.Path
	if a.ViewMode == ViewHybridTree && a.TreeFocused {
		dir = a.TreePanel.SelectedPath()
	}

	entries := a.ActivePanel.Entries
	if a.ViewMode == ViewHybridTree {
		entries = a.RightPanel.Entries
	}

	// Collect directory sizes from cache
	type dirSize struct {
		Name string
		Size int64
	}
	var dirs []dirSize
	var totalSize int64

	for _, e := range entries {
		if !e.IsDir || e.Name == ".." {
			continue
		}
		path := filepath.Join(dir, e.Name)
		size, ok := a.DirSizeCache.Get(path)
		if !ok {
			// Calculate synchronously for dirs not yet cached
			cancel := make(chan struct{})
			size = calcDirSizeWithCancel(path, cancel)
		}
		dirs = append(dirs, dirSize{Name: e.Name, Size: size})
		totalSize += size
	}

	// Also count files in current directory
	for _, e := range entries {
		if e.IsDir || e.Name == ".." {
			continue
		}
		totalSize += e.Size
	}

	if len(dirs) == 0 {
		a.setStatusError("No subdirectories")
		return
	}

	// Sort by size descending
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Size > dirs[j].Size
	})

	a.DialogActive = true
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	tv.SetBorder(true)
	tv.SetTitle(fmt.Sprintf(" Disk Usage: %s ", filepath.Base(dir)))
	tv.SetBorderPadding(1, 1, 2, 2)

	var text strings.Builder
	barWidth := 30

	for _, d := range dirs {
		pct := float64(0)
		if totalSize > 0 {
			pct = float64(d.Size) / float64(totalSize) * 100
		}
		filled := int(float64(barWidth) * pct / 100)
		if filled > barWidth {
			filled = barWidth
		}

		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		fmt.Fprintf(&text, "[yellow]%8s[white]  %s  [blue]%s/[-] (%.1f%%)\n",
			FormatSize(d.Size), bar, d.Name, pct)
	}

	// Add total
	fmt.Fprintf(&text, "\n[green]Total: %s[-]", FormatSize(totalSize))

	tv.SetText(text.String())

	closeDialog := func() {
		a.DialogActive = false
		a.Pages.RemovePage("disk-usage")
		a.restoreFocus()
	}

	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			closeDialog()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				closeDialog()
				return nil
			case 'j':
				row, col := tv.GetScrollOffset()
				tv.ScrollTo(row+1, col)
				return nil
			case 'k':
				row, col := tv.GetScrollOffset()
				if row > 0 {
					tv.ScrollTo(row-1, col)
				}
				return nil
			}
		}
		return event
	})

	height := len(dirs) + 6
	if height > 30 {
		height = 30
	}
	width := 70
	overlay := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(tv, height, 0, true).
			AddItem(nil, 0, 1, false), width, 0, true).
		AddItem(nil, 0, 1, false)

	a.Pages.AddPage("disk-usage", overlay, true, true)
	a.Application.SetFocus(tv)
}

// showRecentDirs displays a list of recently visited directories.
func (a *App) showRecentDirs() {
	panel := a.ActivePanel
	if a.ViewMode == ViewHybridTree {
		panel = a.RightPanel
	}
	if panel.History == nil {
		return
	}

	dirs := panel.History.Recent(20)
	if len(dirs) == 0 {
		a.setStatusError("No recent directories")
		return
	}

	a.DialogActive = true
	list := tview.NewList()
	list.SetBorder(true)
	list.SetTitle(" Recent Directories ")
	list.SetBorderPadding(0, 0, 1, 1)
	list.SetHighlightFullLine(true)
	list.SetSecondaryTextColor(tcell.ColorGray)

	for i, dir := range dirs {
		shortcut := rune(0)
		if i < 9 {
			shortcut = rune('1' + i)
		}
		list.AddItem(dir, "", shortcut, nil)
	}

	closeDialog := func() {
		a.DialogActive = false
		a.Pages.RemovePage("recent-dirs")
		a.restoreFocus()
	}

	navigateTo := func(idx int) {
		if idx < 0 || idx >= len(dirs) {
			closeDialog()
			return
		}
		target := dirs[idx]
		closeDialog()
		if a.AnchorActive && !a.isPathInScope(target) {
			a.setStatusError("Path outside anchor scope")
			return
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

	list.SetSelectedFunc(func(idx int, _ string, _ string, _ rune) {
		navigateTo(idx)
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			closeDialog()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k':
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			case 'q':
				closeDialog()
				return nil
			}
		}
		return event
	})

	height := len(dirs) + 4
	if height > 24 {
		height = 24
	}
	width := 60
	overlay := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(list, height, 0, true).
			AddItem(nil, 0, 1, false), width, 0, true).
		AddItem(nil, 0, 1, false)

	a.Pages.AddPage("recent-dirs", overlay, true, true)
	a.Application.SetFocus(list)
}

// togglePreviewPane toggles the inline preview pane on/off.
