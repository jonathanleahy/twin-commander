package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestIsArchive(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"file.zip", true},
		{"file.ZIP", true},
		{"file.tar.gz", true},
		{"file.tgz", true},
		{"file.tar.bz2", true},
		{"file.tar", true},
		{"file.txt", false},
		{"file.go", false},
		{"archive.zip.bak", false},
	}
	for _, tt := range tests {
		if got := IsArchive(tt.name); got != tt.want {
			t.Errorf("IsArchive(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestListArchive_Zip(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "test.zip")

	// Create a small zip file
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	fw, _ := w.Create("hello.txt")
	fw.Write([]byte("hello world"))
	fw2, _ := w.Create("subdir/nested.txt")
	fw2.Write([]byte("nested content"))
	w.Close()
	f.Close()

	entries, err := ListArchive(zipPath)
	if err != nil {
		t.Fatalf("ListArchive failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Path != "hello.txt" {
		t.Errorf("expected hello.txt, got %s", entries[0].Path)
	}
	if entries[1].Path != "subdir/nested.txt" {
		t.Errorf("expected subdir/nested.txt, got %s", entries[1].Path)
	}
}

func TestFormatArchiveListing(t *testing.T) {
	entries := []ArchiveEntry{
		{Path: "file1.txt", Size: 1024},
		{Path: "dir", Dir: true},
	}
	result := FormatArchiveListing(entries, "test.zip")
	if len(result) == 0 {
		t.Error("expected non-empty listing")
	}
	if !containsStr(result, "file1.txt") {
		t.Error("expected file1.txt in listing")
	}
	if !containsStr(result, "dir/") {
		t.Error("expected dir/ in listing")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && findSubstring(s, sub))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
