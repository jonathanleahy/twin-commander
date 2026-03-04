package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadEntriesRecursive(t *testing.T) {
	dir := t.TempDir()

	// Create nested structure
	os.MkdirAll(filepath.Join(dir, "sub1", "deep"), 0755)
	os.MkdirAll(filepath.Join(dir, "sub2"), 0755)
	os.WriteFile(filepath.Join(dir, "root.txt"), []byte("r"), 0644)
	os.WriteFile(filepath.Join(dir, "sub1", "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "sub1", "deep", "b.txt"), []byte("b"), 0644)
	os.WriteFile(filepath.Join(dir, "sub2", "c.txt"), []byte("c"), 0644)

	entries, err := ReadEntriesRecursive(dir, true)
	if err != nil {
		t.Fatalf("ReadEntriesRecursive failed: %v", err)
	}

	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	// Check that relative paths are used
	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name] = true
	}

	expected := []string{"root.txt", filepath.Join("sub1", "a.txt"), filepath.Join("sub1", "deep", "b.txt"), filepath.Join("sub2", "c.txt")}
	for _, exp := range expected {
		if !names[exp] {
			t.Errorf("expected entry %q not found", exp)
		}
	}
}

func TestReadEntriesRecursive_SkipsHidden(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, ".hidden"), 0755)
	os.WriteFile(filepath.Join(dir, ".hidden", "secret.txt"), []byte("s"), 0644)
	os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("v"), 0644)

	entries, err := ReadEntriesRecursive(dir, false)
	if err != nil {
		t.Fatalf("ReadEntriesRecursive failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (hidden skipped), got %d", len(entries))
	}
	if entries[0].Name != "visible.txt" {
		t.Errorf("expected visible.txt, got %s", entries[0].Name)
	}
}

func TestReadEntriesRecursive_NoDirs(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, "emptydir"), 0755)
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("f"), 0644)

	entries, err := ReadEntriesRecursive(dir, true)
	if err != nil {
		t.Fatalf("ReadEntriesRecursive failed: %v", err)
	}

	for _, e := range entries {
		if e.IsDir {
			t.Errorf("expected no directories, got %s", e.Name)
		}
	}
}
