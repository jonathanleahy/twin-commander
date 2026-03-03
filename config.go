package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds user-configurable settings, persisted to ~/.config/twin-commander/config.json.
type Config struct {
	Theme              string   `json:"theme"`
	ShowHidden         bool     `json:"show_hidden"`
	PreviewOnStart     bool     `json:"preview_on_start"`
	ConfirmDelete      bool     `json:"confirm_delete"`
	UseTrash           bool     `json:"use_trash"`
	DefaultSortMode    string   `json:"default_sort_mode"`
	DefaultSortAsc     bool     `json:"default_sort_asc"`
	DefaultViewMode    string   `json:"default_view_mode"`
	EditorCommand      string   `json:"editor_command"`
	Bookmarks          []string `json:"bookmarks"`
	NerdFontDismissed  bool     `json:"nerd_font_dismissed"`
}

func configDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config", "twin-commander")
}

func configPath() string {
	return filepath.Join(configDir(), "config.json")
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Theme:           string(ThemeDefault),
		ShowHidden:      false,
		PreviewOnStart:  false,
		ConfirmDelete:   true,
		UseTrash:        true,
		DefaultSortMode: "name",
		DefaultSortAsc:  true,
		DefaultViewMode: "hybrid",
		EditorCommand:   "",
	}
}

// LoadConfig reads configuration from disk, returning defaults on any error.
func LoadConfig() Config {
	cfg := DefaultConfig()
	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, &cfg)
	return cfg
}

// SaveConfig writes the configuration to disk.
func SaveConfig(cfg Config) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0o644)
}
