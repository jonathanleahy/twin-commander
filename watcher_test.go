package main

import (
	"testing"
)

func TestFileWatcher_UpdateWatchedDirs(t *testing.T) {
	app := &App{}
	fw, err := NewFileWatcher(app)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Close()

	dir := t.TempDir()
	fw.UpdateWatchedDirs([]string{dir})

	if !fw.watched[dir] {
		t.Error("expected dir to be watched")
	}

	// Update to empty — should unwatch
	fw.UpdateWatchedDirs([]string{})
	if fw.watched[dir] {
		t.Error("expected dir to be unwatched")
	}
}

func TestFileWatcher_UpdateSwapDirs(t *testing.T) {
	app := &App{}
	fw, err := NewFileWatcher(app)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Close()

	dir1 := t.TempDir()
	dir2 := t.TempDir()

	fw.UpdateWatchedDirs([]string{dir1})
	if !fw.watched[dir1] {
		t.Error("expected dir1 to be watched")
	}

	// Swap to dir2
	fw.UpdateWatchedDirs([]string{dir2})
	if fw.watched[dir1] {
		t.Error("expected dir1 to be unwatched after swap")
	}
	if !fw.watched[dir2] {
		t.Error("expected dir2 to be watched after swap")
	}
}
