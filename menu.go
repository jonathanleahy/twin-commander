package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// MenuItem represents a single item in a dropdown menu.
type MenuItem struct {
	Label    string
	Shortcut string
	Action   func()
	Disabled bool
}

// Menu represents a top-level menu with a dropdown of items.
type Menu struct {
	Title  string
	Hotkey rune
	Items  []MenuItem
}

// MenuBar manages the top menu bar and its dropdown menus.
type MenuBar struct {
	Bar        *tview.TextView
	Dropdown   *tview.List
	Menus      []Menu
	ActiveMenu int
	IsOpen     bool
	OnClose    func()
	theme      ThemeColors
}

// NewMenuBar creates a menu bar from the given menus.
func NewMenuBar(menus []Menu) *MenuBar {
	mb := &MenuBar{
		Bar:        tview.NewTextView().SetDynamicColors(true),
		Dropdown:   tview.NewList().ShowSecondaryText(false),
		Menus:      menus,
		ActiveMenu: -1,
	}
	mb.Bar.SetTextAlign(tview.AlignLeft)
	mb.Dropdown.SetBorder(true)
	mb.Dropdown.SetBorderPadding(0, 0, 1, 1)
	mb.renderBar()
	return mb
}

// MenuForHotkey returns the menu index matching the given hotkey rune, or -1.
func (mb *MenuBar) MenuForHotkey(r rune) int {
	lower := unicode.ToLower(r)
	for i, m := range mb.Menus {
		if unicode.ToLower(m.Hotkey) == lower {
			return i
		}
	}
	return -1
}

// ApplyTheme applies color scheme to the menu bar.
func (mb *MenuBar) ApplyTheme(tc ThemeColors) {
	mb.theme = tc
	mb.Bar.SetBackgroundColor(tc.MenuBarBg)
	mb.Dropdown.SetBackgroundColor(tc.DropdownBg)
	mb.Dropdown.SetMainTextColor(tc.DropdownFg)
	mb.Dropdown.SetSelectedBackgroundColor(tc.DropdownSelected)
	mb.Dropdown.SetSelectedTextColor(tc.DropdownFg)
	mb.renderBar()
}

// renderBar redraws the menu bar text.
func (mb *MenuBar) renderBar() {
	var parts []string
	for i, m := range mb.Menus {
		label := m.Title
		if mb.IsOpen && i == mb.ActiveMenu {
			label = fmt.Sprintf("[::r] %s [::-]", label)
		} else {
			// Highlight hotkey
			label = highlightHotkey(label, m.Hotkey, mb.theme)
		}
		parts = append(parts, " "+label+" ")
	}
	mb.Bar.SetText(strings.Join(parts, ""))
}

// highlightHotkey underlines the hotkey character in the label.
func highlightHotkey(label string, hotkey rune, tc ThemeColors) string {
	lower := unicode.ToLower(hotkey)
	for i, r := range label {
		if unicode.ToLower(r) == lower {
			return label[:i] + "[yellow]" + string(r) + "[-]" + label[i+len(string(r)):]
		}
	}
	return label
}

// openDropdown populates the dropdown list for the active menu.
func (mb *MenuBar) openDropdown() {
	mb.IsOpen = true
	mb.Dropdown.Clear()
	if mb.ActiveMenu < 0 || mb.ActiveMenu >= len(mb.Menus) {
		return
	}
	menu := mb.Menus[mb.ActiveMenu]
	for _, item := range menu.Items {
		label := item.Label
		if item.Shortcut != "" {
			label = fmt.Sprintf("%-25s %s", item.Label, item.Shortcut)
		}
		style := tcell.StyleDefault
		if item.Disabled {
			style = style.Foreground(tcell.ColorDarkGray)
		}
		_ = style // tview List doesn't use tcell.Style directly; disabled items are visual-only
		mb.Dropdown.AddItem(label, "", 0, nil)
	}
	mb.Dropdown.SetCurrentItem(0)
	mb.Dropdown.SetTitle(fmt.Sprintf(" %s ", menu.Title))
	mb.renderBar()
}

// DropdownWidth returns the width for the dropdown.
func (mb *MenuBar) DropdownWidth() int {
	if mb.ActiveMenu < 0 || mb.ActiveMenu >= len(mb.Menus) {
		return 30
	}
	maxLen := 0
	for _, item := range mb.Menus[mb.ActiveMenu].Items {
		l := len(item.Label)
		if item.Shortcut != "" {
			l = 25 + 1 + len(item.Shortcut)
		}
		if l > maxLen {
			maxLen = l
		}
	}
	return maxLen + 6 // padding + border
}

// DropdownHeight returns the height for the dropdown.
func (mb *MenuBar) DropdownHeight() int {
	if mb.ActiveMenu < 0 || mb.ActiveMenu >= len(mb.Menus) {
		return 5
	}
	return len(mb.Menus[mb.ActiveMenu].Items) + 2 // items + border
}

// DropdownOffset returns the horizontal offset for the dropdown.
func (mb *MenuBar) DropdownOffset() int {
	offset := 0
	for i := 0; i < mb.ActiveMenu && i < len(mb.Menus); i++ {
		offset += len(mb.Menus[i].Title) + 3 // " Title "
	}
	return offset
}

// MenuIndexAtX returns which menu title was clicked based on x coordinate, or -1.
func (mb *MenuBar) MenuIndexAtX(x int) int {
	offset := 0
	for i, m := range mb.Menus {
		width := len(m.Title) + 3 // " Title "
		if x >= offset && x < offset+width {
			return i
		}
		offset += width
	}
	return -1
}
