package main

import (
	"fmt"
	"os"
	"path/filepath"
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
  F8 / d   Delete (trash or permanent, multi-select)
  F2 / R   Rename
  yy       Yank (mark for copy, multi-select)
  p        Paste yanked files
  dd       Delete (vim-style)

[yellow]Tools[-]
  e        Open in $EDITOR (respects config)
  o        Open with system default (xdg-open)
  b        Beyond Compare
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

// togglePreviewPane toggles the inline preview pane on/off.
