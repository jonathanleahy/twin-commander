package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindDuplicates_NoDups(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("world"), 0644)

	groups := FindDuplicates(dir, true)
	if len(groups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(groups))
	}
}

func TestFindDuplicates_WithDups(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("same content"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("same content"), 0644)
	os.WriteFile(filepath.Join(dir, "c.txt"), []byte("different"), 0644)

	groups := FindDuplicates(dir, true)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if len(groups[0].Paths) != 2 {
		t.Errorf("expected 2 paths in group, got %d", len(groups[0].Paths))
	}
}

func TestFindDuplicates_InSubdirs(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("dup content"), 0644)
	os.WriteFile(filepath.Join(dir, "sub", "b.txt"), []byte("dup content"), 0644)

	groups := FindDuplicates(dir, true)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
}

func TestFindDuplicates_SkipsEmpty(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte{}, 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte{}, 0644)

	groups := FindDuplicates(dir, true)
	if len(groups) != 0 {
		t.Errorf("expected 0 groups (empty files skipped), got %d", len(groups))
	}
}

func TestFormatDuplicates_Empty(t *testing.T) {
	result := FormatDuplicates(nil, "/tmp")
	if result != "No duplicates found." {
		t.Errorf("unexpected: %s", result)
	}
}
