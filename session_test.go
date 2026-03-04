package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// withTempConfigDir overrides configDir for testing by setting XDG_CONFIG_HOME.
func withTempConfigDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	// configDir() uses os.UserHomeDir() + ".config/twin-commander"
	// We override by setting HOME to a temp dir so configDir returns tmp/.config/twin-commander
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmp)
	t.Cleanup(func() {
		os.Setenv("HOME", origHome)
	})
	return filepath.Join(tmp, ".config", "twin-commander")
}

func TestSaveAndLoadSession_RoundTrip(t *testing.T) {
	withTempConfigDir(t)

	workspaces := []*Workspace{
		{
			Name:          "main",
			LeftPath:      "/tmp",
			RightPath:     "/tmp",
			LeftSortMode:  SortByName,
			RightSortMode: SortBySize,
			HSplit:        40,
			VSplit:        60,
			ViewMode:      ViewHybridTree,
			TreeFocused:   true,
			ActiveIsLeft:  false,
			TreeRootPath:  "/tmp",
			TreeExpandedPaths: map[string]bool{
				"/tmp": true,
			},
		},
	}

	err := SaveSession(workspaces, 0)
	if err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	loaded, err := LoadSession()
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadSession returned nil")
	}

	if len(loaded.Workspaces) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(loaded.Workspaces))
	}

	ws := loaded.Workspaces[0]
	if ws.Name != "main" {
		t.Errorf("expected name 'main', got %q", ws.Name)
	}
	if ws.LeftPath != "/tmp" {
		t.Errorf("expected LeftPath '/tmp', got %q", ws.LeftPath)
	}
	if ws.RightPath != "/tmp" {
		t.Errorf("expected RightPath '/tmp', got %q", ws.RightPath)
	}
	if ws.LeftSortMode != SortByName {
		t.Errorf("expected LeftSortMode SortByName, got %d", ws.LeftSortMode)
	}
	if ws.RightSortMode != SortBySize {
		t.Errorf("expected RightSortMode SortBySize, got %d", ws.RightSortMode)
	}
	if ws.HSplit != 40 {
		t.Errorf("expected HSplit 40, got %d", ws.HSplit)
	}
	if ws.VSplit != 60 {
		t.Errorf("expected VSplit 60, got %d", ws.VSplit)
	}
	if ws.ViewMode != ViewHybridTree {
		t.Errorf("expected ViewMode ViewHybridTree, got %d", ws.ViewMode)
	}
	if !ws.TreeFocused {
		t.Error("expected TreeFocused true")
	}
	if ws.ActiveIsLeft {
		t.Error("expected ActiveIsLeft false")
	}
	if loaded.ActiveIndex != 0 {
		t.Errorf("expected ActiveIndex 0, got %d", loaded.ActiveIndex)
	}
	if !loaded.LastSaved.IsZero() == false {
		// LastSaved should be set
	}
	if loaded.LastSaved.IsZero() {
		t.Error("expected LastSaved to be set")
	}
}

func TestLoadSession_MissingFile(t *testing.T) {
	withTempConfigDir(t)

	loaded, err := LoadSession()
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil for missing file")
	}
}

func TestLoadSession_CorruptJSON(t *testing.T) {
	cfgDir := withTempConfigDir(t)
	os.MkdirAll(cfgDir, 0o755)

	err := os.WriteFile(filepath.Join(cfgDir, "session.json"), []byte("not json{{{"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadSession()
	if err == nil {
		t.Fatal("expected error for corrupt JSON")
	}
	if loaded != nil {
		t.Fatal("expected nil result for corrupt JSON")
	}
}

func TestLoadSession_InvalidPathsFallback(t *testing.T) {
	cfgDir := withTempConfigDir(t)
	os.MkdirAll(cfgDir, 0o755)

	data := SessionData{
		Workspaces: []Workspace{
			{
				Name:      "test",
				LeftPath:  "/nonexistent/path/that/does/not/exist",
				RightPath: "/another/fake/path",
				HSplit:    5,  // out of range, should be corrected
				VSplit:    95, // out of range, should be corrected
			},
		},
		ActiveIndex: 0,
	}
	raw, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(filepath.Join(cfgDir, "session.json"), raw, 0o644)

	loaded, err := LoadSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected non-nil session")
	}

	ws := loaded.Workspaces[0]
	// Paths should fall back to home dir (or cwd)
	if ws.LeftPath == "/nonexistent/path/that/does/not/exist" {
		t.Error("expected LeftPath to be corrected from nonexistent path")
	}
	if ws.RightPath == "/another/fake/path" {
		t.Error("expected RightPath to be corrected from nonexistent path")
	}
	// HSplit/VSplit should be corrected
	if ws.HSplit != 33 {
		t.Errorf("expected HSplit corrected to 33, got %d", ws.HSplit)
	}
	if ws.VSplit != 50 {
		t.Errorf("expected VSplit corrected to 50, got %d", ws.VSplit)
	}
}

func TestSaveAndLoadSession_MultipleWorkspaces(t *testing.T) {
	withTempConfigDir(t)

	workspaces := []*Workspace{
		{Name: "ws1", LeftPath: "/tmp", RightPath: "/tmp", HSplit: 33, VSplit: 50},
		{Name: "ws2", LeftPath: "/tmp", RightPath: "/tmp", HSplit: 50, VSplit: 50},
		{Name: "ws3", LeftPath: "/tmp", RightPath: "/tmp", HSplit: 25, VSplit: 75},
	}

	err := SaveSession(workspaces, 1)
	if err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	loaded, err := LoadSession()
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}

	if len(loaded.Workspaces) != 3 {
		t.Fatalf("expected 3 workspaces, got %d", len(loaded.Workspaces))
	}
	if loaded.ActiveIndex != 1 {
		t.Errorf("expected ActiveIndex 1, got %d", loaded.ActiveIndex)
	}
	if loaded.Workspaces[0].Name != "ws1" {
		t.Errorf("expected ws1, got %q", loaded.Workspaces[0].Name)
	}
	if loaded.Workspaces[2].HSplit != 25 {
		t.Errorf("expected HSplit 25, got %d", loaded.Workspaces[2].HSplit)
	}
}

func TestLoadSession_InvalidActiveIndex(t *testing.T) {
	cfgDir := withTempConfigDir(t)
	os.MkdirAll(cfgDir, 0o755)

	data := SessionData{
		Workspaces: []Workspace{
			{Name: "only", LeftPath: "/tmp", RightPath: "/tmp", HSplit: 33, VSplit: 50},
		},
		ActiveIndex: 99, // out of range
	}
	raw, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(filepath.Join(cfgDir, "session.json"), raw, 0o644)

	loaded, err := LoadSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded.ActiveIndex != 0 {
		t.Errorf("expected ActiveIndex corrected to 0, got %d", loaded.ActiveIndex)
	}
}

func TestSaveSession_EmptyWorkspaces(t *testing.T) {
	withTempConfigDir(t)

	err := SaveSession([]*Workspace{}, 0)
	if err != nil {
		t.Fatalf("SaveSession with empty workspaces failed: %v", err)
	}

	loaded, err := LoadSession()
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}
	if len(loaded.Workspaces) != 0 {
		t.Errorf("expected 0 workspaces, got %d", len(loaded.Workspaces))
	}
}

func TestLoadSession_PrunesNonexistentExpandedPaths(t *testing.T) {
	cfgDir := withTempConfigDir(t)
	os.MkdirAll(cfgDir, 0o755)

	data := SessionData{
		Workspaces: []Workspace{
			{
				Name:      "test",
				LeftPath:  "/tmp",
				RightPath: "/tmp",
				HSplit:    33,
				VSplit:    50,
				TreeExpandedPaths: map[string]bool{
					"/tmp":                              true,
					"/nonexistent/should/be/pruned":     true,
					"/also/nonexistent/should/be/pruned": true,
				},
			},
		},
		ActiveIndex: 0,
	}
	raw, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(filepath.Join(cfgDir, "session.json"), raw, 0o644)

	loaded, err := LoadSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ws := loaded.Workspaces[0]
	if len(ws.TreeExpandedPaths) != 1 {
		t.Errorf("expected 1 expanded path after pruning, got %d", len(ws.TreeExpandedPaths))
	}
	if !ws.TreeExpandedPaths["/tmp"] {
		t.Error("expected /tmp to remain in expanded paths")
	}
}

func TestDirExists(t *testing.T) {
	tmp := t.TempDir()

	if !dirExists(tmp) {
		t.Error("expected temp dir to exist")
	}
	if dirExists(filepath.Join(tmp, "nonexistent")) {
		t.Error("expected nonexistent to not exist")
	}

	// Create a file — should return false (not a dir)
	f := filepath.Join(tmp, "file.txt")
	os.WriteFile(f, []byte("hi"), 0o644)
	if dirExists(f) {
		t.Error("expected file to not be treated as dir")
	}
}
