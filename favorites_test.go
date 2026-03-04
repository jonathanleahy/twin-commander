package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFavorites_ToggleAddRemove(t *testing.T) {
	dir := t.TempDir()
	f := &Favorites{path: filepath.Join(dir, "fav.json")}

	added := f.Toggle("/home/user/docs")
	if !added {
		t.Error("expected Toggle to return true (added)")
	}
	if !f.Has("/home/user/docs") {
		t.Error("expected Has to return true after add")
	}

	removed := f.Toggle("/home/user/docs")
	if removed {
		t.Error("expected Toggle to return false (removed)")
	}
	if f.Has("/home/user/docs") {
		t.Error("expected Has to return false after remove")
	}
}

func TestFavorites_Persistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fav.json")

	f1 := &Favorites{path: path}
	f1.Toggle("/home/user/projects")
	f1.Toggle("/home/user/docs")

	// Load a new instance from same file
	f2 := &Favorites{path: path}
	f2.load()

	if len(f2.Paths) != 2 {
		t.Fatalf("expected 2 favorites after reload, got %d", len(f2.Paths))
	}
	if !f2.Has("/home/user/projects") {
		t.Error("expected /home/user/projects in reloaded favorites")
	}
}

func TestFavorites_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fav.json")

	// Non-existent file should not error
	f := &Favorites{path: path}
	f.load()
	if len(f.Paths) != 0 {
		t.Error("expected empty paths for non-existent file")
	}
}

func TestFavorites_SaveCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "fav.json")

	f := &Favorites{path: path}
	f.Toggle("/tmp")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected favorites file to be created")
	}
}
