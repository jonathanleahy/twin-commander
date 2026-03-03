package main

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

// Workspace captures the full state of a single workspace (both panels, view mode, etc.).
type Workspace struct {
	Name             string
	LeftPath         string
	LeftShowHidden   bool
	LeftSortMode     SortMode
	LeftSortOrder    SortOrder
	RightPath        string
	RightShowHidden  bool
	RightSortMode    SortMode
	RightSortOrder   SortOrder
	ViewMode         ViewMode
	TreeFocused      bool
	PreviewActive    bool
	ActiveIsLeft     bool
	HSplit           int
	VSplit           int
	TreeRootPath     string
	TreeExpandedPaths map[string]bool
}

// WorkspaceManager manages multiple workspaces with a tab bar.
type WorkspaceManager struct {
	Workspaces []*Workspace
	Active     int
	TabBar     *tview.TextView
}

// NewWorkspaceManager creates a manager with one default workspace.
func NewWorkspaceManager() *WorkspaceManager {
	wm := &WorkspaceManager{
		Workspaces: []*Workspace{
			{Name: "1"},
		},
		Active: 0,
		TabBar: tview.NewTextView().SetDynamicColors(true),
	}
	wm.TabBar.SetTextAlign(tview.AlignLeft)
	wm.renderTabBar()
	return wm
}

// renderTabBar draws the tab bar text with the active workspace highlighted.
func (wm *WorkspaceManager) renderTabBar() {
	var parts []string
	for i, ws := range wm.Workspaces {
		label := ws.Name
		if label == "" {
			label = fmt.Sprintf("%d", i+1)
		}
		if i == wm.Active {
			parts = append(parts, fmt.Sprintf(" [::r] %s [::-] ", label))
		} else {
			parts = append(parts, fmt.Sprintf("  %s  ", label))
		}
	}
	wm.TabBar.SetText(strings.Join(parts, ""))
}

// AddWorkspace creates a new workspace and returns its index.
func (wm *WorkspaceManager) AddWorkspace(name string) int {
	if name == "" {
		name = fmt.Sprintf("%d", len(wm.Workspaces)+1)
	}
	wm.Workspaces = append(wm.Workspaces, &Workspace{Name: name})
	wm.renderTabBar()
	return len(wm.Workspaces) - 1
}

// RemoveWorkspace removes a workspace by index. Returns false if it's the last one.
func (wm *WorkspaceManager) RemoveWorkspace(index int) bool {
	if len(wm.Workspaces) <= 1 {
		return false
	}
	if index < 0 || index >= len(wm.Workspaces) {
		return false
	}
	wm.Workspaces = append(wm.Workspaces[:index], wm.Workspaces[index+1:]...)
	if wm.Active >= len(wm.Workspaces) {
		wm.Active = len(wm.Workspaces) - 1
	} else if wm.Active > index {
		wm.Active--
	}
	wm.renderTabBar()
	return true
}

// Count returns the number of workspaces.
func (wm *WorkspaceManager) Count() int {
	return len(wm.Workspaces)
}

// Current returns the active workspace.
func (wm *WorkspaceManager) Current() *Workspace {
	if wm.Active >= 0 && wm.Active < len(wm.Workspaces) {
		return wm.Workspaces[wm.Active]
	}
	return nil
}
