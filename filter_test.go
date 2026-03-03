package main

import (
	"testing"
)

// TS-29: Filter Is Case-Insensitive
func TestFilterEntries_CaseInsensitive(t *testing.T) {
	entries := []FileEntry{
		{Name: "..", IsDir: true, Accessible: true},
		{Name: "README.md", Accessible: true},
		{Name: "readme.txt", Accessible: true},
		{Name: "Info.MD", Accessible: true},
		{Name: "config.json", Accessible: true},
	}
	result := FilterEntries(entries, "md")
	// Should match README.md, readme.txt (no), Info.MD — wait, "readme.txt" doesn't contain "md"
	// Actually: README.md contains "md", Info.MD contains "md", readme.txt does not contain "md"
	// Plus .. is always included
	expected := map[string]bool{"..": true, "README.md": true, "Info.MD": true}
	if len(result) != len(expected) {
		t.Fatalf("FilterEntries with 'md' returned %d entries, want %d", len(result), len(expected))
	}
	for _, e := range result {
		if !expected[e.Name] {
			t.Errorf("unexpected entry %q in filtered results", e.Name)
		}
	}
}

// EC-10: Filter Matches Nothing
func TestFilterEntries_MatchesNothing(t *testing.T) {
	entries := []FileEntry{
		{Name: "..", IsDir: true, Accessible: true},
		{Name: "docs", IsDir: true, Accessible: true},
		{Name: "src", IsDir: true, Accessible: true},
		{Name: "README.md", Accessible: true},
	}
	result := FilterEntries(entries, "xyz123")
	// Only .. should remain
	if len(result) != 1 {
		t.Fatalf("FilterEntries with 'xyz123' returned %d entries, want 1", len(result))
	}
	if result[0].Name != ".." {
		t.Errorf("expected '..' entry, got %q", result[0].Name)
	}
}

// TS-43: .. Entry Never Filtered Out
func TestFilterEntries_DotDotNeverFiltered(t *testing.T) {
	entries := []FileEntry{
		{Name: "..", IsDir: true, Accessible: true},
		{Name: "file.txt", Accessible: true},
	}
	result := FilterEntries(entries, "zzz")
	found := false
	for _, e := range result {
		if e.Name == ".." {
			found = true
			break
		}
	}
	if !found {
		t.Error("'..' entry should never be filtered out")
	}
}

// Test empty query returns all entries
func TestFilterEntries_EmptyQuery(t *testing.T) {
	entries := []FileEntry{
		{Name: "..", IsDir: true, Accessible: true},
		{Name: "docs", IsDir: true, Accessible: true},
		{Name: "README.md", Accessible: true},
	}
	result := FilterEntries(entries, "")
	if len(result) != len(entries) {
		t.Errorf("empty query should return all %d entries, got %d", len(entries), len(result))
	}
}

// Test substring matching
func TestFilterEntries_SubstringMatch(t *testing.T) {
	entries := []FileEntry{
		{Name: "..", IsDir: true, Accessible: true},
		{Name: "notes.txt", Accessible: true},
		{Name: "README.TXT", Accessible: true},
		{Name: "txtfile", Accessible: true},
		{Name: "config.json", Accessible: true},
	}
	result := FilterEntries(entries, "txt")
	expected := map[string]bool{"..": true, "notes.txt": true, "README.TXT": true, "txtfile": true}
	if len(result) != len(expected) {
		t.Fatalf("FilterEntries with 'txt' returned %d entries, want %d", len(result), len(expected))
	}
	for _, e := range result {
		if !expected[e.Name] {
			t.Errorf("unexpected entry %q in filtered results", e.Name)
		}
	}
}

// TestFilterGlob tests glob pattern matching with *
func TestFilterGlob(t *testing.T) {
	entries := []FileEntry{
		{Name: "..", IsDir: true, Accessible: true},
		{Name: "main.go", Accessible: true},
		{Name: "utils.go", Accessible: true},
		{Name: "README.md", Accessible: true},
	}
	result := FilterEntries(entries, "*.go")
	expected := map[string]bool{"..": true, "main.go": true, "utils.go": true}
	if len(result) != len(expected) {
		t.Fatalf("FilterEntries with '*.go' returned %d entries, want %d", len(result), len(expected))
	}
	for _, e := range result {
		if !expected[e.Name] {
			t.Errorf("unexpected entry %q in filtered results", e.Name)
		}
	}
}

// TestFilterGlob_QuestionMark tests glob pattern matching with ?
func TestFilterGlob_QuestionMark(t *testing.T) {
	entries := []FileEntry{
		{Name: "..", IsDir: true, Accessible: true},
		{Name: "a.txt", Accessible: true},
		{Name: "ab.txt", Accessible: true},
		{Name: "b.txt", Accessible: true},
	}
	result := FilterEntries(entries, "?.txt")
	expected := map[string]bool{"..": true, "a.txt": true, "b.txt": true}
	if len(result) != len(expected) {
		t.Fatalf("FilterEntries with '?.txt' returned %d entries, want %d", len(result), len(expected))
	}
	for _, e := range result {
		if !expected[e.Name] {
			t.Errorf("unexpected entry %q in filtered results", e.Name)
		}
	}
}

// TestFilterRegex tests regex pattern matching
func TestFilterRegex(t *testing.T) {
	entries := []FileEntry{
		{Name: "..", IsDir: true, Accessible: true},
		{Name: "test_foo.go", Accessible: true},
		{Name: "test_bar.go", Accessible: true},
		{Name: "mytest.go", Accessible: true},
		{Name: "README.md", Accessible: true},
	}
	result := FilterEntries(entries, `/^test.*\.go$/`)
	expected := map[string]bool{"..": true, "test_foo.go": true, "test_bar.go": true}
	if len(result) != len(expected) {
		t.Fatalf("FilterEntries with regex returned %d entries, want %d", len(result), len(expected))
	}
	for _, e := range result {
		if !expected[e.Name] {
			t.Errorf("unexpected entry %q in filtered results", e.Name)
		}
	}
}

// TestFilterNegation tests negation with ! prefix
func TestFilterNegation(t *testing.T) {
	entries := []FileEntry{
		{Name: "..", IsDir: true, Accessible: true},
		{Name: "main.go", Accessible: true},
		{Name: "temp.tmp", Accessible: true},
		{Name: "cache.tmp", Accessible: true},
		{Name: "README.md", Accessible: true},
	}
	result := FilterEntries(entries, "!*.tmp")
	expected := map[string]bool{"..": true, "main.go": true, "README.md": true}
	if len(result) != len(expected) {
		t.Fatalf("FilterEntries with '!*.tmp' returned %d entries, want %d", len(result), len(expected))
	}
	for _, e := range result {
		if !expected[e.Name] {
			t.Errorf("unexpected entry %q in filtered results", e.Name)
		}
	}
}

// TestFilterMultiTerm tests multi-term OR matching (space-separated substring)
func TestFilterMultiTerm(t *testing.T) {
	entries := []FileEntry{
		{Name: "..", IsDir: true, Accessible: true},
		{Name: "main.go", Accessible: true},
		{Name: "notes.txt", Accessible: true},
		{Name: "README.md", Accessible: true},
	}
	result := FilterEntries(entries, "go txt")
	expected := map[string]bool{"..": true, "main.go": true, "notes.txt": true}
	if len(result) != len(expected) {
		t.Fatalf("FilterEntries with 'go txt' returned %d entries, want %d", len(result), len(expected))
	}
	for _, e := range result {
		if !expected[e.Name] {
			t.Errorf("unexpected entry %q in filtered results", e.Name)
		}
	}
}

// TestFilterRegex_Invalid tests that invalid regex falls back to substring matching
func TestFilterRegex_Invalid(t *testing.T) {
	entries := []FileEntry{
		{Name: "..", IsDir: true, Accessible: true},
		{Name: "[invalid", Accessible: true},
		{Name: "valid.go", Accessible: true},
	}
	// /[invalid/ is an invalid regex, falls back to substring match on "/[invalid/"
	// Neither entry name contains "/[invalid/" so only ".." is returned
	result := FilterEntries(entries, "/[invalid/")
	if len(result) != 1 {
		t.Fatalf("FilterEntries with invalid regex returned %d entries, want 1 (only ..)", len(result))
	}
	if result[0].Name != ".." {
		t.Errorf("expected '..' entry, got %q", result[0].Name)
	}
}
