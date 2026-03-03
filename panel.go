package main

import (
	"fmt"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Panel represents one side of the dual-pane file explorer.
type Panel struct {
	Path                string
	Entries             []FileEntry
	ShowHidden          bool
	Filter              string
	Table               *tview.Table
	StatusBar           string
	SortMode            SortMode
	SortOrder           SortOrder
	GitRepo             *GitRepo    // Set by App for git-aware rendering
	ActiveBorderColor   tcell.Color // Theme-driven active border
	InactiveBorderColor tcell.Color // Theme-driven inactive border
	Selection           *Selection  // Multi-file selection model
	History             *History    // Directory navigation history
}

// LoadDir reads the directory, sorts, filters, and renders entries to the Table.
func (p *Panel) LoadDir() {
	entries, err := ReadEntries(p.Path, p.ShowHidden)
	if err != nil {
		p.StatusBar = fmt.Sprintf("Cannot read directory: %s", p.Path)
		return
	}

	entries = SortEntriesBy(entries, p.SortMode, p.SortOrder)

	// Prepend ".." if not at root
	if p.Path != "/" {
		dotdot := FileEntry{Name: "..", IsDir: true, Accessible: true}
		entries = append([]FileEntry{dotdot}, entries...)
	}

	// Apply filter
	if p.Filter != "" {
		entries = FilterEntries(entries, p.Filter)
	}

	p.Entries = entries
	p.renderTable()
	p.updateStatusBar()
	p.Table.SetTitle(p.Path)
}

// renderTable draws all entries into the tview.Table.
func (p *Panel) renderTable() {
	p.Table.Clear()

	for i, e := range p.Entries {
		nameText := e.Name
		nameStyle := tcell.StyleDefault

		if e.Name == ".." {
			// .. entry: blue+bold, no "/" suffix
			nameText = iconDotDot + ".."
			nameStyle = tcell.StyleDefault.Foreground(tcell.ColorBlue).Bold(true)
		} else if !e.Accessible {
			// Inaccessible: dark gray. Still add "/" suffix for dirs.
			nameStyle = tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
			if e.IsDir {
				nameText = FileIcon(e.Name, true, false, false) + e.Name + "/"
			} else {
				nameText = FileIcon(e.Name, false, false, false) + e.Name
			}
		} else if e.IsSymlink {
			// Symlinks: purple. Add "/" suffix if target is directory.
			nameStyle = tcell.StyleDefault.Foreground(tcell.ColorPurple)
			if e.IsDir {
				nameText = FileIcon(e.Name, false, true, false) + e.Name + "/"
			} else {
				nameText = FileIcon(e.Name, false, true, false) + e.Name
			}
		} else if e.IsDir {
			// Regular directories: blue+bold with "/" suffix
			nameText = FileIcon(e.Name, true, false, false) + e.Name + "/"
			nameStyle = tcell.StyleDefault.Foreground(tcell.ColorBlue).Bold(true)
		} else if e.IsExecutable {
			nameText = FileIcon(e.Name, false, false, true) + e.Name
			nameStyle = tcell.StyleDefault.Foreground(tcell.ColorGreen)
		} else {
			nameText = FileIcon(e.Name, false, false, false) + e.Name
		}

		// Overlay git status color
		if p.GitRepo != nil && e.Name != ".." {
			if e.IsDir {
				relDir := p.GitRepo.RelPath(filepath.Join(p.Path, e.Name))
				status := p.GitRepo.GetDirStatus(relDir)
				if color, hasColor := GitStatusColor(status); hasColor {
					nameStyle = tcell.StyleDefault.Foreground(color).Bold(true)
				}
			} else {
				relPath := p.GitRepo.RelPath(filepath.Join(p.Path, e.Name))
				status := p.GitRepo.GetFileStatus(relPath)
				if color, hasColor := GitStatusColor(status); hasColor {
					nameStyle = tcell.StyleDefault.Foreground(color)
				}
			}
		}

		// Selection marker
		if p.Selection != nil && e.Name != ".." {
			fullPath := filepath.Join(p.Path, e.Name)
			if p.Selection.IsSelected(fullPath) {
				nameText = ">" + nameText
				nameStyle = nameStyle.Background(tcell.ColorDarkGoldenrod)
			}
		}

		// Name column
		nameCell := tview.NewTableCell(nameText).
			SetStyle(nameStyle).
			SetExpansion(1).
			SetAlign(tview.AlignLeft)
		p.Table.SetCell(i, 0, nameCell)

		// Size column
		var sizeText string
		if e.Name == ".." {
			sizeText = ""
		} else if !e.Accessible {
			sizeText = "---"
		} else {
			sizeText = FormatSize(e.Size)
		}
		sizeCell := tview.NewTableCell(sizeText).
			SetAlign(tview.AlignRight)
		p.Table.SetCell(i, 1, sizeCell)

		// Permissions column
		var permText string
		if e.Name == ".." {
			permText = ""
		} else if !e.Accessible {
			permText = "---"
		} else {
			permText = FormatPermissions(e.Mode)
		}
		permCell := tview.NewTableCell(permText).
			SetAlign(tview.AlignLeft)
		p.Table.SetCell(i, 2, permCell)

		// Date column
		var dateText string
		if e.Name == ".." {
			dateText = ""
		} else if !e.Accessible {
			dateText = "---"
		} else {
			dateText = e.ModTime.Format("2006-01-02")
		}
		dateCell := tview.NewTableCell(dateText).
			SetAlign(tview.AlignLeft)
		p.Table.SetCell(i, 3, dateCell)
	}

	p.Table.ScrollToBeginning()
}

// updateStatusBar computes the status bar text.
func (p *Panel) updateStatusBar() {
	count := 0
	var totalSize int64
	for _, e := range p.Entries {
		if e.Name == ".." {
			continue
		}
		count++
		if e.Accessible && e.Size > 0 {
			totalSize += e.Size
		}
	}

	text := fmt.Sprintf("%d items, %s", count, FormatSize(totalSize))
	if p.ShowHidden {
		text = "[H] " + text
	}
	if p.SortMode != SortByName || p.SortOrder != SortAsc {
		text += fmt.Sprintf(" [%s %s]", SortModeLabel(p.SortMode), SortOrderArrow(p.SortOrder))
	}
	if p.Selection != nil && p.Selection.Count() > 0 {
		// Calculate selected total size
		var selSize int64
		for _, path := range p.Selection.Paths() {
			for _, e := range p.Entries {
				if filepath.Join(p.Path, e.Name) == path && e.Accessible {
					selSize += e.Size
				}
			}
		}
		text += fmt.Sprintf(" | %d selected (%s)", p.Selection.Count(), FormatSize(selSize))
	}
	p.StatusBar = text
}

// StatusText returns the current status bar text.
func (p *Panel) StatusText() string {
	return p.StatusBar
}

// NavigateInto changes the panel directory to a subdirectory.
func (p *Panel) NavigateInto(name string) {
	oldPath := p.Path
	p.Path = filepath.Join(p.Path, name)
	p.Filter = ""
	if p.History != nil {
		p.History.Push(oldPath)
	}
	p.LoadDir()
	p.Table.Select(0, 0)
}

// TryNavigateInto attempts to navigate into a subdirectory.
// Returns an error if the directory cannot be read.
func (p *Panel) TryNavigateInto(name string) error {
	targetPath := filepath.Join(p.Path, name)
	_, err := ReadEntries(targetPath, p.ShowHidden)
	if err != nil {
		return err
	}
	p.NavigateInto(name)
	return nil
}

// NavigateUp moves to the parent directory.
// Returns the basename of the previous directory for cursor positioning.
func (p *Panel) NavigateUp() string {
	if p.Path == "/" {
		return ""
	}
	oldPath := p.Path
	prevName := filepath.Base(p.Path)
	p.Path = filepath.Dir(p.Path)
	p.Filter = ""
	if p.History != nil {
		p.History.Push(oldPath)
	}
	p.LoadDir()

	// Position cursor on the previous directory
	for i, e := range p.Entries {
		if e.Name == prevName {
			p.Table.Select(i, 0)
			return prevName
		}
	}
	// If not found, select row 0
	p.Table.Select(0, 0)
	return prevName
}

// ToggleHidden flips the ShowHidden state and reloads.
func (p *Panel) ToggleHidden() {
	p.ShowHidden = !p.ShowHidden

	// Save currently selected entry name
	selectedName := ""
	row, _ := p.Table.GetSelection()
	if row >= 0 && row < len(p.Entries) {
		selectedName = p.Entries[row].Name
	}

	p.LoadDir()

	// Restore cursor
	p.restoreCursor(selectedName)
}

// Refresh reloads the current directory.
func (p *Panel) Refresh() {
	selectedName := ""
	row, _ := p.Table.GetSelection()
	if row >= 0 && row < len(p.Entries) {
		selectedName = p.Entries[row].Name
	}

	p.LoadDir()
	p.restoreCursor(selectedName)
}

// SetFilter sets the filter query and reloads.
func (p *Panel) SetFilter(query string) {
	p.Filter = query
	p.LoadDir()
}

// ClearFilter removes the filter and reloads.
func (p *Panel) ClearFilter() {
	p.Filter = ""
	p.LoadDir()
}

// SetActive sets the panel's border color based on active state.
func (p *Panel) SetActive(active bool) {
	if active {
		p.Table.SetBorderColor(p.ActiveBorderColor)
	} else {
		p.Table.SetBorderColor(p.InactiveBorderColor)
	}
}

// restoreCursor positions the cursor on the named entry, or nearest valid row.
func (p *Panel) restoreCursor(name string) {
	if name == "" {
		p.Table.Select(0, 0)
		return
	}
	for i, e := range p.Entries {
		if e.Name == name {
			p.Table.Select(i, 0)
			return
		}
	}
	// Entry no longer exists: select nearest valid position
	row, _ := p.Table.GetSelection()
	if row >= len(p.Entries) && len(p.Entries) > 0 {
		p.Table.Select(len(p.Entries)-1, 0)
	} else if len(p.Entries) > 0 {
		p.Table.Select(0, 0)
	}
}

// SelectedEntry returns the currently selected entry, or nil if none.
func (p *Panel) SelectedEntry() *FileEntry {
	row, _ := p.Table.GetSelection()
	if row >= 0 && row < len(p.Entries) {
		return &p.Entries[row]
	}
	return nil
}
