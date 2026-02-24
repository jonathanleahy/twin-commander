package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App is the application controller.
type App struct {
	Application *tview.Application
	LeftPanel   *Panel
	RightPanel  *Panel
	ActivePanel *Panel
	FilterMode  bool
	FilterInput *tview.InputField
	LeftStatus  *tview.TextView
	RightStatus *tview.TextView
	RootFlex    *tview.Flex
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
		FilterMode:  false,
	}

	// Create panels
	app.LeftPanel = app.createPanel(cwd)
	app.RightPanel = app.createPanel(cwd)
	app.ActivePanel = app.LeftPanel

	// Set initial active state
	app.LeftPanel.SetActive(true)
	app.RightPanel.SetActive(false)

	// Create status bars
	app.LeftStatus = tview.NewTextView().SetDynamicColors(false)
	app.RightStatus = tview.NewTextView().SetDynamicColors(false)

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
			// Keep filter, exit filter mode
			app.exitFilterMode(false)
		case tcell.KeyEscape:
			// Clear filter, exit filter mode
			app.exitFilterMode(true)
		}
	})

	// Build layout
	// Left panel column: panel table + status bar
	leftCol := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(app.LeftPanel.Table, 0, 1, true).
		AddItem(app.LeftStatus, 1, 0, false)

	// Right panel column: panel table + status bar
	rightCol := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(app.RightPanel.Table, 0, 1, true).
		AddItem(app.RightStatus, 1, 0, false)

	// Horizontal flex: left and right columns side by side
	panelsFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(leftCol, 0, 1, true).
		AddItem(rightCol, 0, 1, false)

	// Vertical flex: panels + filter input at bottom
	app.RootFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(panelsFlex, 0, 1, true).
		AddItem(app.FilterInput, 0, 0, false) // Hidden initially (height 0)

	// Load initial directories
	app.LeftPanel.LoadDir()
	app.RightPanel.LoadDir()
	app.updateStatusBars()

	// Set up global key handler
	app.Application.SetInputCapture(app.handleKeyEvent)

	app.Application.SetRoot(app.RootFlex, true)
	app.Application.SetFocus(app.LeftPanel.Table)

	return app
}

// createPanel creates a new Panel with the given path.
func (a *App) createPanel(path string) *Panel {
	table := tview.NewTable()
	table.SetBorder(true)
	table.SetSelectable(true, false)
	table.SetSelectedStyle(tcell.StyleDefault.Reverse(true))
	table.SetFixed(0, 0)

	return &Panel{
		Path:       path,
		Table:      table,
		ShowHidden: false,
		Filter:     "",
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

	if a.FilterMode {
		return a.handleFilterModeKey(event)
	}
	return a.handleNormalModeKey(event)
}

// handleNormalModeKey handles keys in normal (non-filter) mode.
func (a *App) handleNormalModeKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'q':
			a.Application.Stop()
			return nil
		case '.':
			a.ActivePanel.ToggleHidden()
			a.updateStatusBars()
			return nil
		case 'r':
			a.ActivePanel.Refresh()
			a.updateStatusBars()
			return nil
		case '/':
			a.enterFilterMode()
			return nil
		}
	case tcell.KeyTab:
		a.switchPanel()
		return nil
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		a.navigateUp()
		return nil
	case tcell.KeyEnter:
		a.handleEnter()
		return nil
	case tcell.KeyUp:
		return event // Let tview handle cursor movement
	case tcell.KeyDown:
		return event // Let tview handle cursor movement
	}

	return event
}

// handleFilterModeKey handles keys in filter mode.
func (a *App) handleFilterModeKey(event *tcell.EventKey) *tcell.EventKey {
	// Escape and Enter are handled by SetDoneFunc
	// All other keys go to the InputField
	return event
}

// switchPanel toggles the active panel.
func (a *App) switchPanel() {
	if a.ActivePanel == a.LeftPanel {
		a.ActivePanel = a.RightPanel
		a.LeftPanel.SetActive(false)
		a.RightPanel.SetActive(true)
		a.Application.SetFocus(a.RightPanel.Table)
	} else {
		a.ActivePanel = a.LeftPanel
		a.RightPanel.SetActive(false)
		a.LeftPanel.SetActive(true)
		a.Application.SetFocus(a.LeftPanel.Table)
	}
}

// navigateUp handles Backspace in normal mode.
func (a *App) navigateUp() {
	a.ActivePanel.NavigateUp()
	a.updateStatusBars()
}

// handleEnter handles Enter on the selected entry.
func (a *App) handleEnter() {
	entry := a.ActivePanel.SelectedEntry()
	if entry == nil {
		return
	}

	if entry.Name == ".." {
		a.navigateUp()
		return
	}

	if !entry.IsDir {
		return // No-op on files
	}

	if !entry.Accessible {
		a.setStatusError(fmt.Sprintf("Permission denied: %s",
			filepath.Join(a.ActivePanel.Path, entry.Name)))
		return
	}

	// Try to navigate into the directory
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
	a.FilterMode = true
	// Pre-fill with existing filter if any
	a.FilterInput.SetText(a.ActivePanel.Filter)

	// Show filter input (set height to 1)
	a.RootFlex.ResizeItem(a.FilterInput, 1, 0)

	a.Application.SetFocus(a.FilterInput)
}

// exitFilterMode exits filter mode.
func (a *App) exitFilterMode(clearFilter bool) {
	a.FilterMode = false

	if clearFilter {
		a.ActivePanel.ClearFilter()
		a.FilterInput.SetText("")
	}

	// Hide filter input
	a.RootFlex.ResizeItem(a.FilterInput, 0, 0)

	a.Application.SetFocus(a.ActivePanel.Table)
	a.updateStatusBars()
}

// updateStatusBars refreshes both status bar texts.
func (a *App) updateStatusBars() {
	a.LeftStatus.SetText(a.LeftPanel.StatusText())
	a.RightStatus.SetText(a.RightPanel.StatusText())
}

// setStatusError displays an error message in the active panel's status bar.
func (a *App) setStatusError(msg string) {
	if a.ActivePanel == a.LeftPanel {
		a.LeftStatus.SetText(msg)
	} else {
		a.RightStatus.SetText(msg)
	}
	// Store in panel's StatusBar so next action restores it
	a.ActivePanel.StatusBar = msg
}
