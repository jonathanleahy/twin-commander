package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// BookmarkManager holds a list of bookmarked directory paths.
type BookmarkManager struct {
	paths []string
}

// NewBookmarkManager creates a bookmark manager with initial paths.
func NewBookmarkManager(paths []string) *BookmarkManager {
	if paths == nil {
		paths = []string{}
	}
	return &BookmarkManager{paths: paths}
}

// Get returns the bookmark at the given index, or "" if out of range.
func (bm *BookmarkManager) Get(idx int) string {
	if idx < 0 || idx >= len(bm.paths) {
		return ""
	}
	return bm.paths[idx]
}

// Add adds a path to the bookmark list (if not already present).
func (bm *BookmarkManager) Add(path string) {
	for _, p := range bm.paths {
		if p == path {
			return
		}
	}
	bm.paths = append(bm.paths, path)
}

// Remove removes a bookmark by index.
func (bm *BookmarkManager) Remove(idx int) {
	if idx < 0 || idx >= len(bm.paths) {
		return
	}
	bm.paths = append(bm.paths[:idx], bm.paths[idx+1:]...)
}

// Paths returns all bookmarked paths.
func (bm *BookmarkManager) Paths() []string {
	result := make([]string, len(bm.paths))
	copy(result, bm.paths)
	return result
}

// ShowBookmarkDialog displays a dialog for managing and selecting bookmarks.
func ShowBookmarkDialog(
	pages *tview.Pages,
	app *tview.Application,
	bm *BookmarkManager,
	currentPath string,
	onSelect func(string),
	onClose func(),
) {
	list := tview.NewList()
	list.SetBorder(true)
	list.SetTitle(" Bookmarks ")
	list.SetBorderPadding(0, 0, 1, 1)
	list.ShowSecondaryText(false)
	list.SetHighlightFullLine(true)

	rebuildList := func() {
		list.Clear()
		for i, p := range bm.Paths() {
			shortcut := rune('1' + i)
			if i > 8 {
				shortcut = 0
			}
			list.AddItem(fmt.Sprintf("%d. %s", i+1, p), "", shortcut, nil)
		}
		// Add current directory option
		list.AddItem(fmt.Sprintf("+ Add current (%s)", currentPath), "", 'a', nil)
	}
	rebuildList()

	closeDialog := func() {
		pages.RemovePage("bookmark-dialog")
		if onClose != nil {
			onClose()
		}
	}

	list.SetSelectedFunc(func(idx int, _ string, _ string, _ rune) {
		paths := bm.Paths()
		if idx < len(paths) {
			closeDialog()
			if onSelect != nil {
				onSelect(paths[idx])
			}
		} else {
			// Add current path
			bm.Add(currentPath)
			rebuildList()
		}
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeDialog()
			return nil
		}
		if event.Key() == tcell.KeyRune && event.Rune() == 'x' {
			// Remove selected bookmark
			idx := list.GetCurrentItem()
			paths := bm.Paths()
			if idx >= 0 && idx < len(paths) {
				bm.Remove(idx)
				rebuildList()
			}
			return nil
		}
		return event
	})

	// Center the dialog
	width := 60
	height := len(bm.Paths()) + 5
	if height < 8 {
		height = 8
	}
	if height > 20 {
		height = 20
	}

	overlay := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(list, height, 0, true).
			AddItem(nil, 0, 1, false), width, 0, true).
		AddItem(nil, 0, 1, false)

	pages.AddPage("bookmark-dialog", overlay, true, true)
	app.SetFocus(list)
}
