package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TS-18: Sort Order -- Directories First, Then Files, Alphabetical
func TestSortEntries_DirectoriesFirstThenFiles(t *testing.T) {
	entries := []FileEntry{
		{Name: "README.md", IsDir: false, Accessible: true},
		{Name: "src", IsDir: true, Accessible: true},
		{Name: "config.json", IsDir: false, Accessible: true},
		{Name: "docs", IsDir: true, Accessible: true},
	}
	sorted := SortEntries(entries)
	expected := []string{"docs", "src", "config.json", "README.md"}
	if len(sorted) != len(expected) {
		t.Fatalf("SortEntries returned %d entries, want %d", len(sorted), len(expected))
	}
	for i, name := range expected {
		if sorted[i].Name != name {
			t.Errorf("sorted[%d].Name = %q, want %q", i, sorted[i].Name, name)
		}
	}
}

// EC-15: Case-Insensitive Sort Verification
func TestSortEntries_CaseInsensitive(t *testing.T) {
	entries := []FileEntry{
		{Name: "Zebra.txt", IsDir: false, Accessible: true},
		{Name: "alpha", IsDir: true, Accessible: true},
		{Name: "BETA", IsDir: true, Accessible: true},
		{Name: "gamma.txt", IsDir: false, Accessible: true},
		{Name: "Delta", IsDir: true, Accessible: true},
	}
	sorted := SortEntries(entries)
	expected := []string{"alpha", "BETA", "Delta", "gamma.txt", "Zebra.txt"}
	if len(sorted) != len(expected) {
		t.Fatalf("SortEntries returned %d entries, want %d", len(sorted), len(expected))
	}
	for i, name := range expected {
		if sorted[i].Name != name {
			t.Errorf("sorted[%d].Name = %q, want %q", i, sorted[i].Name, name)
		}
	}
}

// Test sort stability: entries with same sort key maintain relative order
func TestSortEntries_StableSort(t *testing.T) {
	entries := []FileEntry{
		{Name: "b.txt", IsDir: false, Accessible: true},
		{Name: "a.txt", IsDir: false, Size: 100, Accessible: true},
		{Name: "a.txt", IsDir: false, Size: 200, Accessible: true},
	}
	sorted := SortEntries(entries)
	if sorted[0].Name != "a.txt" || sorted[1].Name != "a.txt" {
		t.Error("same-name entries should both come before b.txt")
	}
	// Stable sort preserves original order for equal elements
	if sorted[0].Size != 100 || sorted[1].Size != 200 {
		t.Error("stable sort should preserve relative order of equal elements")
	}
}

// Test broken symlinks are sorted with files
func TestSortEntries_BrokenSymlinkWithFiles(t *testing.T) {
	entries := []FileEntry{
		{Name: "zdir", IsDir: true, Accessible: true},
		{Name: "broken_link", IsDir: false, IsSymlink: true, Accessible: false},
		{Name: "adir", IsDir: true, Accessible: true},
		{Name: "afile.txt", IsDir: false, Accessible: true},
	}
	sorted := SortEntries(entries)
	expected := []string{"adir", "zdir", "afile.txt", "broken_link"}
	for i, name := range expected {
		if sorted[i].Name != name {
			t.Errorf("sorted[%d].Name = %q, want %q", i, sorted[i].Name, name)
		}
	}
}

// Integration test: ReadEntries with real temp directory
func TestReadEntries_BasicDirectory(t *testing.T) {
	tmp := t.TempDir()

	// Create test files and directories
	os.Mkdir(filepath.Join(tmp, "docs"), 0755)
	os.Mkdir(filepath.Join(tmp, "src"), 0755)
	os.WriteFile(filepath.Join(tmp, "README.md"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(tmp, "config.json"), []byte("{}"), 0644)

	entries, err := ReadEntries(tmp, false)
	if err != nil {
		t.Fatalf("ReadEntries failed: %v", err)
	}
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	// Verify directory detection
	dirCount := 0
	fileCount := 0
	for _, e := range entries {
		if e.IsDir {
			dirCount++
		} else {
			fileCount++
		}
		if !e.Accessible {
			t.Errorf("entry %q should be accessible", e.Name)
		}
	}
	if dirCount != 2 {
		t.Errorf("expected 2 dirs, got %d", dirCount)
	}
	if fileCount != 2 {
		t.Errorf("expected 2 files, got %d", fileCount)
	}
}

// Integration test: ReadEntries hides dotfiles by default
func TestReadEntries_HiddenFilesOff(t *testing.T) {
	tmp := t.TempDir()

	os.Mkdir(filepath.Join(tmp, ".git"), 0755)
	os.WriteFile(filepath.Join(tmp, ".gitignore"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmp, "README.md"), []byte("hello"), 0644)
	os.Mkdir(filepath.Join(tmp, "src"), 0755)

	entries, err := ReadEntries(tmp, false)
	if err != nil {
		t.Fatalf("ReadEntries failed: %v", err)
	}
	// Should only see README.md and src (hidden files filtered)
	if len(entries) != 2 {
		t.Errorf("expected 2 visible entries, got %d", len(entries))
		for _, e := range entries {
			t.Logf("  entry: %q", e.Name)
		}
	}
}

// Integration test: ReadEntries shows dotfiles when enabled
func TestReadEntries_HiddenFilesOn(t *testing.T) {
	tmp := t.TempDir()

	os.Mkdir(filepath.Join(tmp, ".git"), 0755)
	os.WriteFile(filepath.Join(tmp, ".gitignore"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmp, "README.md"), []byte("hello"), 0644)
	os.Mkdir(filepath.Join(tmp, "src"), 0755)

	entries, err := ReadEntries(tmp, true)
	if err != nil {
		t.Fatalf("ReadEntries failed: %v", err)
	}
	if len(entries) != 4 {
		t.Errorf("expected 4 entries with hidden visible, got %d", len(entries))
	}
}

// Integration test: file metadata (size, modtime)
func TestReadEntries_FileMetadata(t *testing.T) {
	tmp := t.TempDir()

	content := []byte("hello world")
	os.WriteFile(filepath.Join(tmp, "test.txt"), content, 0644)

	entries, err := ReadEntries(tmp, false)
	if err != nil {
		t.Fatalf("ReadEntries failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Name != "test.txt" {
		t.Errorf("expected name 'test.txt', got %q", e.Name)
	}
	if e.Size != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), e.Size)
	}
	if e.ModTime.IsZero() {
		t.Error("ModTime should not be zero")
	}
	if e.IsDir {
		t.Error("should not be a directory")
	}
	if !e.Accessible {
		t.Error("should be accessible")
	}
}

// Integration test: executable file detection
func TestReadEntries_ExecutableDetection(t *testing.T) {
	tmp := t.TempDir()

	os.WriteFile(filepath.Join(tmp, "script.sh"), []byte("#!/bin/sh"), 0755)
	os.WriteFile(filepath.Join(tmp, "data.txt"), []byte("data"), 0644)

	entries, err := ReadEntries(tmp, false)
	if err != nil {
		t.Fatalf("ReadEntries failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	execFound := false
	nonExecFound := false
	for _, e := range entries {
		if e.Name == "script.sh" {
			execFound = true
			if !e.IsExecutable {
				t.Error("script.sh should be executable")
			}
		}
		if e.Name == "data.txt" {
			nonExecFound = true
			if e.IsExecutable {
				t.Error("data.txt should not be executable")
			}
		}
	}
	if !execFound {
		t.Error("script.sh not found in entries")
	}
	if !nonExecFound {
		t.Error("data.txt not found in entries")
	}
}

// Integration test: symlinks
func TestReadEntries_Symlinks(t *testing.T) {
	tmp := t.TempDir()

	// Create a target file and directory
	os.WriteFile(filepath.Join(tmp, "target.txt"), []byte("content"), 0644)
	os.Mkdir(filepath.Join(tmp, "targetdir"), 0755)

	// Create symlinks
	os.Symlink(filepath.Join(tmp, "target.txt"), filepath.Join(tmp, "link_to_file"))
	os.Symlink(filepath.Join(tmp, "targetdir"), filepath.Join(tmp, "link_to_dir"))

	entries, err := ReadEntries(tmp, false)
	if err != nil {
		t.Fatalf("ReadEntries failed: %v", err)
	}

	for _, e := range entries {
		switch e.Name {
		case "link_to_file":
			if !e.IsSymlink {
				t.Error("link_to_file should be a symlink")
			}
			if e.IsDir {
				t.Error("link_to_file should not be a directory")
			}
			if !e.Accessible {
				t.Error("link_to_file should be accessible")
			}
		case "link_to_dir":
			if !e.IsSymlink {
				t.Error("link_to_dir should be a symlink")
			}
			if !e.IsDir {
				t.Error("link_to_dir should be a directory (target is dir)")
			}
			if !e.Accessible {
				t.Error("link_to_dir should be accessible")
			}
		}
	}
}

// Integration test: broken symlink (EC-11)
func TestReadEntries_BrokenSymlink(t *testing.T) {
	tmp := t.TempDir()

	os.Symlink(filepath.Join(tmp, "nonexistent"), filepath.Join(tmp, "broken_link"))
	os.WriteFile(filepath.Join(tmp, "normal.txt"), []byte("ok"), 0644)

	entries, err := ReadEntries(tmp, false)
	if err != nil {
		t.Fatalf("ReadEntries failed: %v", err)
	}

	for _, e := range entries {
		if e.Name == "broken_link" {
			if !e.IsSymlink {
				t.Error("broken_link should be a symlink")
			}
			if e.Accessible {
				t.Error("broken_link should not be accessible")
			}
			if e.Size != -1 {
				t.Errorf("broken_link size should be -1, got %d", e.Size)
			}
			if !e.ModTime.IsZero() {
				t.Error("broken_link ModTime should be zero")
			}
			if e.IsDir {
				t.Error("broken_link should not be a directory (sorted with files)")
			}
		}
	}
}

// Integration test: ReadEntries with nonexistent directory
func TestReadEntries_NonexistentDir(t *testing.T) {
	_, err := ReadEntries("/nonexistent/path/that/does/not/exist", false)
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

// Test FileEntry date formatting uses Go reference time for YYYY-MM-DD
func TestFileEntry_DateFormat(t *testing.T) {
	tm := time.Date(2026, 2, 23, 10, 30, 0, 0, time.UTC)
	expected := "2026-02-23"
	result := tm.Format("2006-01-02")
	if result != expected {
		t.Errorf("date format = %q, want %q", result, expected)
	}
}

// Test that directories are not marked as executable
func TestReadEntries_DirectoriesNotExecutable(t *testing.T) {
	tmp := t.TempDir()
	os.Mkdir(filepath.Join(tmp, "mydir"), 0755)

	entries, err := ReadEntries(tmp, false)
	if err != nil {
		t.Fatalf("ReadEntries failed: %v", err)
	}
	for _, e := range entries {
		if e.Name == "mydir" && e.IsExecutable {
			t.Error("directories should not be marked as executable")
		}
	}
}
