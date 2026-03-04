package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// SessionData holds the persisted session state.
type SessionData struct {
	Workspaces  []Workspace `json:"workspaces"`
	ActiveIndex int         `json:"active_index"`
	LastSaved   time.Time   `json:"last_saved"`
}

// sessionPath returns the path to the session file.
func sessionPath() string {
	return filepath.Join(configDir(), "session.json")
}

// SaveSession persists the current workspace state to disk.
func SaveSession(workspaces []*Workspace, activeIndex int) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Convert pointer slice to value slice for serialization
	ws := make([]Workspace, len(workspaces))
	for i, w := range workspaces {
		if w != nil {
			ws[i] = *w
		}
	}

	data := SessionData{
		Workspaces:  ws,
		ActiveIndex: activeIndex,
		LastSaved:   time.Now(),
	}

	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sessionPath(), raw, 0o644)
}

// LoadSession reads the persisted session from disk.
// Returns nil, nil if the file does not exist.
func LoadSession() (*SessionData, error) {
	raw, err := os.ReadFile(sessionPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var data SessionData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	// Validate paths exist — fall back to home dir for missing paths
	home, _ := os.UserHomeDir()
	if home == "" {
		home, _ = os.Getwd()
	}
	for i := range data.Workspaces {
		ws := &data.Workspaces[i]
		if ws.LeftPath == "" || !dirExists(ws.LeftPath) {
			ws.LeftPath = home
		}
		if ws.RightPath == "" || !dirExists(ws.RightPath) {
			ws.RightPath = home
		}
		if ws.HSplit < 10 || ws.HSplit > 90 {
			ws.HSplit = 33
		}
		if ws.VSplit < 10 || ws.VSplit > 90 {
			ws.VSplit = 50
		}
		// Prune non-existent expanded paths
		if ws.TreeExpandedPaths != nil {
			for p := range ws.TreeExpandedPaths {
				if !dirExists(p) {
					delete(ws.TreeExpandedPaths, p)
				}
			}
		}
	}

	// Validate active index
	if data.ActiveIndex < 0 || data.ActiveIndex >= len(data.Workspaces) {
		data.ActiveIndex = 0
	}

	return &data, nil
}

// dirExists checks if a path exists and is a directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
