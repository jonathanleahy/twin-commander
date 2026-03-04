package main

import "testing"

func TestWorkspaceManager_NewHasOne(t *testing.T) {
	wm := NewWorkspaceManager()
	if wm.Count() != 1 {
		t.Errorf("expected 1 workspace, got %d", wm.Count())
	}
	if wm.Active != 0 {
		t.Errorf("expected active 0, got %d", wm.Active)
	}
}

func TestWorkspaceManager_AddCreates(t *testing.T) {
	wm := NewWorkspaceManager()
	idx := wm.AddWorkspace("test")
	if wm.Count() != 2 {
		t.Errorf("expected 2 workspaces, got %d", wm.Count())
	}
	if idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}
	if wm.Workspaces[1].Name != "test" {
		t.Errorf("expected name 'test', got %q", wm.Workspaces[1].Name)
	}
}

func TestWorkspaceManager_CantRemoveLast(t *testing.T) {
	wm := NewWorkspaceManager()
	ok := wm.RemoveWorkspace(0)
	if ok {
		t.Error("should not be able to remove last workspace")
	}
	if wm.Count() != 1 {
		t.Error("workspace count should still be 1")
	}
}

func TestWorkspaceManager_RemoveAdjustsActive(t *testing.T) {
	wm := NewWorkspaceManager()
	wm.AddWorkspace("two")
	wm.AddWorkspace("three")
	wm.Active = 2 // select third workspace

	ok := wm.RemoveWorkspace(2) // remove third
	if !ok {
		t.Fatal("should be able to remove non-last workspace")
	}
	if wm.Count() != 2 {
		t.Errorf("expected 2 workspaces, got %d", wm.Count())
	}
	if wm.Active != 1 {
		t.Errorf("expected active adjusted to 1, got %d", wm.Active)
	}
}

func TestWorkspaceManager_RemoveMiddle(t *testing.T) {
	wm := NewWorkspaceManager()
	wm.AddWorkspace("two")
	wm.AddWorkspace("three")
	wm.Active = 2 // third is active

	ok := wm.RemoveWorkspace(1) // remove middle
	if !ok {
		t.Fatal("should be able to remove middle workspace")
	}
	if wm.Count() != 2 {
		t.Errorf("expected 2 workspaces, got %d", wm.Count())
	}
	// Active was 2, after removing index 1, it should adjust to 1
	if wm.Active != 1 {
		t.Errorf("expected active 1 after removing middle, got %d", wm.Active)
	}
}

func TestWorkspaceManager_Current(t *testing.T) {
	wm := NewWorkspaceManager()
	wm.AddWorkspace("second")
	wm.Active = 1
	ws := wm.Current()
	if ws == nil {
		t.Fatal("expected non-nil workspace")
	}
	if ws.Name != "second" {
		t.Errorf("expected name 'second', got %q", ws.Name)
	}
}

func TestWorkspaceManager_AddAutoName(t *testing.T) {
	wm := NewWorkspaceManager()
	wm.AddWorkspace("")
	if wm.Workspaces[1].Name != "2" {
		t.Errorf("expected auto-name '2', got %q", wm.Workspaces[1].Name)
	}
}

func TestWorkspace_AnchorRoundTrip(t *testing.T) {
	ws := &Workspace{
		Name:         "test",
		AnchorPath:   "/home/user/project",
		AnchorActive: true,
	}
	if ws.AnchorPath != "/home/user/project" {
		t.Errorf("expected anchor path '/home/user/project', got %q", ws.AnchorPath)
	}
	if !ws.AnchorActive {
		t.Error("expected anchor active to be true")
	}

	// Simulate clearing anchor
	ws.AnchorActive = false
	ws.AnchorPath = ""
	if ws.AnchorActive {
		t.Error("expected anchor active to be false after clear")
	}
	if ws.AnchorPath != "" {
		t.Errorf("expected empty anchor path after clear, got %q", ws.AnchorPath)
	}
}

func TestWorkspaceManager_RenderTabBar(t *testing.T) {
	wm := NewWorkspaceManager()
	wm.AddWorkspace("B")
	wm.Active = 0
	wm.renderTabBar()
	text := wm.TabBar.GetText(true)
	if text == "" {
		t.Error("tab bar should have text")
	}
}
