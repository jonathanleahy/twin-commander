package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Favorites manages a persistent list of pinned directories.
type Favorites struct {
	Paths []string `json:"paths"`
	path  string   // file path for persistence
}

// NewFavorites creates a Favorites manager, loading from disk if available.
func NewFavorites() *Favorites {
	f := &Favorites{
		path: favoritesPath(),
	}
	f.load()
	return f
}

// Has returns true if the given path is in favorites.
func (f *Favorites) Has(path string) bool {
	for _, p := range f.Paths {
		if p == path {
			return true
		}
	}
	return false
}

// Toggle adds or removes a path from favorites. Returns true if added.
func (f *Favorites) Toggle(path string) bool {
	if f.Has(path) {
		f.remove(path)
		f.save()
		return false
	}
	f.Paths = append(f.Paths, path)
	f.save()
	return true
}

func (f *Favorites) remove(path string) {
	for i, p := range f.Paths {
		if p == path {
			f.Paths = append(f.Paths[:i], f.Paths[i+1:]...)
			return
		}
	}
}

func (f *Favorites) load() {
	data, err := os.ReadFile(f.path)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, f)
}

func (f *Favorites) save() {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return
	}
	dir := filepath.Dir(f.path)
	_ = os.MkdirAll(dir, 0755)
	_ = os.WriteFile(f.path, data, 0644)
}

func favoritesPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configDir, "twin-commander", "favorites.json")
}
