package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rivo/tview"
)

// handleCopy copies the selected entry to the other panel's directory.
// If multi-select is active, copies all selected files.
func (a *App) handleCopy() {
	// Multi-selection mode
	if a.ActivePanel.Selection != nil && a.ActivePanel.Selection.Count() > 0 {
		paths := a.ActivePanel.Selection.Paths()
		dstDir := a.InactivePanel().Path
		if filepath.Clean(a.ActivePanel.Path) == filepath.Clean(dstDir) {
			a.setStatusError("Cannot copy to same location")
			return
		}
		msg := fmt.Sprintf("Copy %d selected items to %s?", len(paths), dstDir)
		a.DialogActive = true
		ShowConfirmDialog(a.Pages, "Copy", msg, func(confirmed bool) {
			a.DialogActive = false
			if !confirmed {
				a.restoreFocus()
				return
			}
			var errors []string
			for _, src := range paths {
				dst := filepath.Join(dstDir, filepath.Base(src))
				info, err := os.Stat(src)
				if err != nil {
					errors = append(errors, fmt.Sprintf("%s: %v", filepath.Base(src), err))
					continue
				}
				if info.IsDir() {
					if err := CopyDir(src, dst); err != nil {
						errors = append(errors, fmt.Sprintf("%s: %v", filepath.Base(src), err))
					}
				} else {
					if err := CopyFile(src, dst); err != nil {
						errors = append(errors, fmt.Sprintf("%s: %v", filepath.Base(src), err))
					}
				}
			}
			a.ActivePanel.Selection.Clear()
			a.refreshAllPanels()
			if len(errors) > 0 {
				ShowErrorDialog(a.Pages, fmt.Sprintf("Copy errors (%d):\n%s", len(errors), strings.Join(errors, "\n")))
			}
			a.restoreFocus()
		})
		return
	}

	entry := a.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name == ".." {
		return
	}
	src := filepath.Join(a.ActivePanel.Path, entry.Name)
	dstDir := a.InactivePanel().Path

	// Guard: prevent copy to same location
	if filepath.Clean(filepath.Dir(src)) == filepath.Clean(dstDir) {
		a.setStatusError("Cannot copy to same location")
		return
	}

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
// If multi-select is active, moves all selected files.
// handleMove moves the selected entry to the other panel's directory.
// If multi-select is active, moves all selected files.
func (a *App) handleMove() {
	// Multi-selection mode
	if a.ActivePanel.Selection != nil && a.ActivePanel.Selection.Count() > 0 {
		paths := a.ActivePanel.Selection.Paths()
		dstDir := a.InactivePanel().Path
		if filepath.Clean(a.ActivePanel.Path) == filepath.Clean(dstDir) {
			a.setStatusError("Cannot move to same location")
			return
		}
		msg := fmt.Sprintf("Move %d selected items to %s?", len(paths), dstDir)
		a.DialogActive = true
		ShowConfirmDialog(a.Pages, "Move", msg, func(confirmed bool) {
			a.DialogActive = false
			if !confirmed {
				a.restoreFocus()
				return
			}
			var errors []string
			for _, src := range paths {
				dst := filepath.Join(dstDir, filepath.Base(src))
				if err := MoveFile(src, dst); err != nil {
					errors = append(errors, fmt.Sprintf("%s: %v", filepath.Base(src), err))
				}
			}
			a.ActivePanel.Selection.Clear()
			a.refreshAllPanels()
			if len(errors) > 0 {
				ShowErrorDialog(a.Pages, fmt.Sprintf("Move errors (%d):\n%s", len(errors), strings.Join(errors, "\n")))
			}
			a.restoreFocus()
		})
		return
	}

	entry := a.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name == ".." {
		return
	}
	src := filepath.Join(a.ActivePanel.Path, entry.Name)
	dstDir := a.InactivePanel().Path

	// Guard: prevent move to same location
	if filepath.Clean(filepath.Dir(src)) == filepath.Clean(dstDir) {
		a.setStatusError("Cannot move to same location")
		return
	}

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
// If multi-select is active, deletes all selected files.
// handleDelete offers trash (default) or permanent delete for the selected entry.
// If multi-select is active, deletes all selected files.
func (a *App) handleDelete() {
	// Multi-selection mode
	if a.ActivePanel.Selection != nil && a.ActivePanel.Selection.Count() > 0 {
		paths := a.ActivePanel.Selection.Paths()
		msg := fmt.Sprintf("Delete %d selected items?", len(paths))
		a.DialogActive = true
		ShowChoiceDialog(a.Pages, "Delete", msg, []string{"Move to Trash", "Permanently Delete", "Cancel"}, func(label string) {
			a.DialogActive = false
			var errors []string
			switch label {
			case "Move to Trash":
				for _, p := range paths {
					if err := MoveToTrash(p); err != nil {
						errors = append(errors, fmt.Sprintf("%s: %v", filepath.Base(p), err))
					}
				}
			case "Permanently Delete":
				for _, p := range paths {
					if err := DeletePath(p); err != nil {
						errors = append(errors, fmt.Sprintf("%s: %v", filepath.Base(p), err))
					}
				}
			}
			a.ActivePanel.Selection.Clear()
			a.refreshAllPanels()
			if len(errors) > 0 {
				ShowErrorDialog(a.Pages, fmt.Sprintf("Delete errors (%d):\n%s", len(errors), strings.Join(errors, "\n")))
			}
			a.restoreFocus()
		})
		return
	}

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
		// Path traversal validation
		absPath, _ := filepath.Abs(path)
		absParent, _ := filepath.Abs(a.ActivePanel.Path)
		if !strings.HasPrefix(absPath, absParent+string(filepath.Separator)) && absPath != absParent {
			a.setStatusError("Path traversal not allowed")
			a.restoreFocus()
			return
		}
		err := MakeDirSafe(path)
		if err != nil {
			ShowErrorDialog(a.Pages, fmt.Sprintf("Mkdir failed: %v", err))
		}
		a.refreshAllPanels()
		a.restoreFocus()
	})
}

// handleMkfile creates a new empty file in the active panel directory.
func (a *App) handleMkfile() {
	a.DialogActive = true
	ShowInputDialog(a.Pages, a.Application, "New File", "Name: ", "", func(value string, cancelled bool) {
		a.DialogActive = false
		if cancelled || value == "" {
			a.restoreFocus()
			return
		}
		path := filepath.Join(a.ActivePanel.Path, value)
		// Path traversal validation
		absPath, _ := filepath.Abs(path)
		absParent, _ := filepath.Abs(a.ActivePanel.Path)
		if !strings.HasPrefix(absPath, absParent+string(filepath.Separator)) && absPath != absParent {
			a.setStatusError("Path traversal not allowed")
			a.restoreFocus()
			return
		}
		// Check if file already exists
		if _, err := os.Stat(path); err == nil {
			a.setStatusError(fmt.Sprintf("File already exists: %s", value))
			a.restoreFocus()
			return
		}
		err := os.WriteFile(path, []byte{}, 0644)
		if err != nil {
			ShowErrorDialog(a.Pages, fmt.Sprintf("Create file failed: %v", err))
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
		// Path traversal validation
		newPath := filepath.Join(filepath.Dir(oldPath), value)
		absNew, _ := filepath.Abs(newPath)
		absParent, _ := filepath.Abs(a.ActivePanel.Path)
		if !strings.HasPrefix(absNew, absParent+string(filepath.Separator)) && absNew != absParent {
			a.setStatusError("Path traversal not allowed")
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
// handleYank marks the selected entry (or selection) for copy (yank).
func (a *App) handleYank() {
	// Multi-selection yank
	if a.ActivePanel.Selection != nil && a.ActivePanel.Selection.Count() > 0 {
		a.YankBuffer = a.ActivePanel.Selection.Paths()
		a.setStatusError(fmt.Sprintf("Yanked %d items", len(a.YankBuffer)))
		return
	}
	entry := a.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name == ".." {
		return
	}
	path := filepath.Join(a.ActivePanel.Path, entry.Name)
	a.YankBuffer = []string{path}
	a.setStatusError(fmt.Sprintf("Yanked: %s", entry.Name))
}

// handlePaste copies yanked files to the active panel's directory.
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
// handleOpenEditor opens the selected file in $EDITOR.
func (a *App) handleOpenEditor() {
	entry := a.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name == ".." || entry.IsDir {
		return
	}
	path := filepath.Join(a.ActivePanel.Path, entry.Name)

	err := OpenInEditor(a.Application, path, a.Config.EditorCommand)
	if err != nil {
		a.setStatusError(fmt.Sprintf("Editor error: %v", err))
	}
	a.refreshAllPanels()
}

// syncTreeToRightPanel updates the tree to match the right panel's current path.
// Called after the right panel navigates (Enter, Backspace, history, etc.) so the
// tree stays consistent with what the right panel is showing.
// handleSelectionToggle toggles selection on the current entry and moves cursor down.
func (a *App) handleSelectionToggle() {
	entry := a.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name == ".." {
		return
	}
	path := filepath.Join(a.ActivePanel.Path, entry.Name)
	a.ActivePanel.Selection.Toggle(path)
	a.ActivePanel.renderTable()
	a.updateStatusBars()
	// Move cursor down
	row, _ := a.ActivePanel.Table.GetSelection()
	if row < len(a.ActivePanel.Entries)-1 {
		a.ActivePanel.Table.Select(row+1, 0)
	}
}

// handleVisualStart enters visual selection mode.
// handleVisualStart enters visual selection mode.
func (a *App) handleVisualStart() {
	row, _ := a.ActivePanel.Table.GetSelection()
	a.ActivePanel.Selection.StartVisual(row)
	a.ActivePanel.Selection.UpdateVisual(row, a.ActivePanel.Entries, a.ActivePanel.Path)
	a.ActivePanel.renderTable()
	a.updateStatusBars()
}

// handleVisualEnd exits visual selection mode.
// handleVisualEnd exits visual selection mode.
func (a *App) handleVisualEnd() {
	if a.ActivePanel.Selection.IsVisual() {
		a.ActivePanel.Selection.EndVisual()
		a.ActivePanel.renderTable()
		a.updateStatusBars()
	}
}

// handleSelectionInvert inverts the selection against current entries.
// handleSelectionInvert inverts the selection against current entries.
func (a *App) handleSelectionInvert() {
	a.ActivePanel.Selection.InvertFromEntries(a.ActivePanel.Entries, a.ActivePanel.Path)
	a.ActivePanel.renderTable()
	a.updateStatusBars()
}

// handleSelectionPattern shows a dialog to select files matching a pattern.
// handleSelectionPattern shows a dialog to select files matching a pattern.
func (a *App) handleSelectionPattern() {
	a.DialogActive = true
	ShowInputDialog(a.Pages, a.Application, "Select Pattern", "Pattern: ", "", func(value string, cancelled bool) {
		a.DialogActive = false
		if cancelled || value == "" {
			a.restoreFocus()
			return
		}
		a.ActivePanel.Selection.MatchPattern(value, a.ActivePanel.Entries, a.ActivePanel.Path)
		a.ActivePanel.renderTable()
		a.updateStatusBars()
		a.restoreFocus()
	})
}

// --- History handlers ---

// handleHistoryBack navigates to the previous directory in history.
// handleHistoryBack navigates to the previous directory in history.
func (a *App) handleHistoryBack() {
	dir, ok := a.ActivePanel.History.Back(a.ActivePanel.Path)
	if !ok {
		a.setStatusError("No history to go back to")
		return
	}
	a.ActivePanel.Path = dir
	a.ActivePanel.LoadDir()
	if a.ViewMode == ViewHybridTree {
		a.TreePanel.NavigateToPath(dir)
	}
	a.updateStatusBars()
}

// handleHistoryForward navigates to the next directory in history.
// handleHistoryForward navigates to the next directory in history.
func (a *App) handleHistoryForward() {
	dir, ok := a.ActivePanel.History.Forward(a.ActivePanel.Path)
	if !ok {
		a.setStatusError("No forward history")
		return
	}
	a.ActivePanel.Path = dir
	a.ActivePanel.LoadDir()
	if a.ViewMode == ViewHybridTree {
		a.TreePanel.NavigateToPath(dir)
	}
	a.updateStatusBars()
}

// --- Open with default application ---

// handleOpenDefault opens the selected file with the system default application.
// handleOpenDefault opens the selected file with the system default application.
func (a *App) handleOpenDefault() {
	entry := a.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name == ".." {
		return
	}
	path := filepath.Join(a.ActivePanel.Path, entry.Name)
	err := OpenWithDefault(path)
	if err != nil {
		a.setStatusError(fmt.Sprintf("Open failed: %v", err))
	}
}

// --- Command bar ---

// enterCommandMode shows the command input bar at the bottom.
// enterCommandMode shows the command input bar at the bottom.
func (a *App) enterCommandMode() {
	a.DialogActive = true
	ShowInputDialog(a.Pages, a.Application, "Command", "$ ", "", func(value string, cancelled bool) {
		a.DialogActive = false
		if cancelled || value == "" {
			a.restoreFocus()
			return
		}

		// Expand variables
		selectedFile := ""
		entry := a.ActivePanel.SelectedEntry()
		if entry != nil && entry.Name != ".." {
			selectedFile = filepath.Join(a.ActivePanel.Path, entry.Name)
		}
		var selectedFiles []string
		if a.ActivePanel.Selection != nil && a.ActivePanel.Selection.Count() > 0 {
			selectedFiles = a.ActivePanel.Selection.Paths()
		}
		expanded := ExpandVariables(value, selectedFile, a.ActivePanel.Path, selectedFiles)

		// Run command
		result := RunCommand(expanded, a.ActivePanel.Path)
		if result.Output != "" {
			// Show output in viewer
			a.Viewer.Wrapper.SetTitle(fmt.Sprintf(" $ %s ", value))
			a.Viewer.TextView.SetText(result.Output)
			a.Viewer.TextView.ScrollToBeginning()
			a.ViewerActive = true
			a.Pages.SwitchToPage("viewer")
			a.Application.SetFocus(a.Viewer.TextView)
		} else {
			a.refreshAllPanels()
			a.restoreFocus()
		}
	})
}

// --- Chmod dialog ---

// handleChmod shows a dialog to change file permissions.
// handleChmod shows a dialog to change file permissions.
func (a *App) handleChmod() {
	entry := a.ActivePanel.SelectedEntry()
	if entry == nil || entry.Name == ".." {
		return
	}
	path := filepath.Join(a.ActivePanel.Path, entry.Name)
	currentMode := fmt.Sprintf("%o", entry.Mode.Perm())

	a.DialogActive = true
	ShowInputDialog(a.Pages, a.Application, "Chmod", "Octal mode: ", currentMode, func(value string, cancelled bool) {
		a.DialogActive = false
		if cancelled || value == "" {
			a.restoreFocus()
			return
		}
		mode, err := ParseOctalMode(value)
		if err != nil {
			a.setStatusError(fmt.Sprintf("Invalid mode: %v", err))
			a.restoreFocus()
			return
		}
		err = ChmodPath(path, mode)
		if err != nil {
			ShowErrorDialog(a.Pages, fmt.Sprintf("Chmod failed: %v", err))
		}
		a.refreshAllPanels()
		a.restoreFocus()
	})
}

// --- Content search (grep) ---

// enterContentSearch opens the content search overlay.
// enterContentSearch opens the content search overlay.
func (a *App) enterContentSearch() {
	a.DialogActive = true
	ShowInputDialog(a.Pages, a.Application, "Content Search", "Pattern: ", "", func(value string, cancelled bool) {
		a.DialogActive = false
		if cancelled || value == "" {
			a.restoreFocus()
			return
		}

		rootDir := a.ActivePanel.Path
		if a.ViewMode == ViewHybridTree {
			rootDir = a.TreePanel.RootPath
		}

		// Set up search
		a.SearchMode = true
		a.SearchInput.SetText("Content: " + value)
		a.SearchTable.Clear()
		a.Pages.SwitchToPage("search")
		a.Application.SetFocus(a.SearchTable)

		// Cancel any previous search
		if a.SearchCancel != nil {
			close(a.SearchCancel)
		}
		a.SearchCancel = make(chan struct{})
		cancelCh := a.SearchCancel

		resultCh := make(chan GrepResult, 100)
		go ContentSearch(GrepOpts{
			RootDir:    rootDir,
			Pattern:    value,
			MaxResults: 1000,
			ShowHidden: a.ActivePanel.ShowHidden,
			IgnoreCase: true,
		}, resultCh, cancelCh)

		go func() {
			row := 0
			for result := range resultCh {
				r := result
				rowIdx := row
				a.Application.QueueUpdateDraw(func() {
					text := fmt.Sprintf("%s:%d: %s", r.RelPath, r.Line, r.Content)
					a.SearchTable.SetCell(rowIdx, 0,
						tview.NewTableCell(text).
							SetReference(r.Path))
				})
				row++
			}
		}()
	})
}

// --- Progress callback for file operations ---

// CopyDirWithProgress copies a directory tree, calling progressFn with (done, total) counts.
// CopyDirWithProgress copies a directory tree, calling progressFn with (done, total) counts.
func CopyDirWithProgress(src, dst string, progressFn func(done, total int)) error {
	// Count files first
	total := 0
	_ = filepath.Walk(src, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			total++
		}
		return nil
	})

	done := 0
	return copyDirProgress(src, dst, &done, total, progressFn)
}

func copyDirProgress(src, dst string, done *int, total int, progressFn func(done, total int)) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.Type()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(srcPath)
			if err != nil {
				return err
			}
			if err := os.Symlink(linkTarget, dstPath); err != nil {
				return err
			}
		} else if entry.IsDir() {
			if err := copyDirProgress(srcPath, dstPath, done, total, progressFn); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
			*done++
			if progressFn != nil {
				progressFn(*done, total)
			}
		}
	}
	return nil
}

// InactivePanel returns the panel that is NOT active (used for file operations).
// InactivePanel returns the panel that is NOT active (used for file operations).
func (a *App) InactivePanel() *Panel {
	if a.ViewMode == ViewHybridTree {
		// In hybrid mode, LeftPanel serves as the "other" directory for file ops.
		// It tracks the last directory from dual-pane mode or can be set explicitly.
		return a.LeftPanel
	}
	if a.ActivePanel == a.LeftPanel {
		return a.RightPanel
	}
	return a.LeftPanel
}
