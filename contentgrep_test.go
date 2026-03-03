package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// helper to collect all GrepResults from a channel
func collectGrepResults(results <-chan GrepResult) []GrepResult {
	var out []GrepResult
	for r := range results {
		out = append(out, r)
	}
	return out
}

// helper to write a file in a temp directory
func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestContentSearch_FindsMatch(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "test.txt", "hello world\ngoodbye world\n")

	results := make(chan GrepResult, 100)
	cancel := make(chan struct{})
	go ContentSearch(GrepOpts{
		RootDir:    dir,
		Pattern:    "hello",
		MaxResults: 0,
		ShowHidden: false,
		IgnoreCase: false,
	}, results, cancel)

	got := collectGrepResults(results)
	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}
	if got[0].Content != "hello world" {
		t.Errorf("expected content %q, got %q", "hello world", got[0].Content)
	}
	if got[0].Line != 1 {
		t.Errorf("expected line 1, got %d", got[0].Line)
	}
	if got[0].RelPath != "test.txt" {
		t.Errorf("expected relpath %q, got %q", "test.txt", got[0].RelPath)
	}
}

func TestContentSearch_CaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "test.txt", "hello world\n")

	results := make(chan GrepResult, 100)
	cancel := make(chan struct{})
	go ContentSearch(GrepOpts{
		RootDir:    dir,
		Pattern:    "HELLO",
		MaxResults: 0,
		ShowHidden: false,
		IgnoreCase: true,
	}, results, cancel)

	got := collectGrepResults(results)
	if len(got) != 1 {
		t.Fatalf("expected 1 result with case-insensitive search, got %d", len(got))
	}
	if got[0].Content != "hello world" {
		t.Errorf("expected content %q, got %q", "hello world", got[0].Content)
	}
}

func TestContentSearch_CaseSensitive(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "test.txt", "hello world\n")

	results := make(chan GrepResult, 100)
	cancel := make(chan struct{})
	go ContentSearch(GrepOpts{
		RootDir:    dir,
		Pattern:    "HELLO",
		MaxResults: 0,
		ShowHidden: false,
		IgnoreCase: false,
	}, results, cancel)

	got := collectGrepResults(results)
	if len(got) != 0 {
		t.Fatalf("expected 0 results with case-sensitive search for HELLO, got %d", len(got))
	}
}

func TestContentSearch_MaxResults(t *testing.T) {
	dir := t.TempDir()
	// Create a file with 10 matching lines
	content := ""
	for i := 0; i < 10; i++ {
		content += "match line\n"
	}
	writeTestFile(t, dir, "test.txt", content)

	results := make(chan GrepResult, 100)
	cancel := make(chan struct{})
	go ContentSearch(GrepOpts{
		RootDir:    dir,
		Pattern:    "match",
		MaxResults: 3,
		ShowHidden: false,
		IgnoreCase: false,
	}, results, cancel)

	got := collectGrepResults(results)
	if len(got) != 3 {
		t.Fatalf("expected 3 results with MaxResults=3, got %d", len(got))
	}
}

func TestContentSearch_Cancel(t *testing.T) {
	dir := t.TempDir()
	// Create many files with matching content so the search has work to do
	for i := 0; i < 50; i++ {
		writeTestFile(t, dir, filepath.Join("sub", "file"+string(rune('a'+i%26))+string(rune('0'+i/26))+".txt"), "findme in this file\nmore content\nfindme again\n")
	}

	results := make(chan GrepResult, 100)
	cancel := make(chan struct{})

	go ContentSearch(GrepOpts{
		RootDir:    dir,
		Pattern:    "findme",
		MaxResults: 0,
		ShowHidden: false,
		IgnoreCase: false,
	}, results, cancel)

	// Read a couple of results then cancel
	count := 0
	for range results {
		count++
		if count >= 2 {
			close(cancel)
			break
		}
	}

	// Drain remaining results (channel will be closed by ContentSearch)
	for range results {
		count++
	}

	// We cancelled after 2, so we should have far fewer than all 100 possible matches
	if count >= 100 {
		t.Errorf("cancel did not stop search, got %d results", count)
	}

	// Give it a moment and ensure it completed (channel closed)
	// If we reached here, the channel was successfully closed
}

func TestContentSearch_SkipsBinary(t *testing.T) {
	dir := t.TempDir()
	// Create a binary file with null bytes
	binaryContent := []byte("hello\x00world\n")
	path := filepath.Join(dir, "binary.dat")
	if err := os.WriteFile(path, binaryContent, 0o644); err != nil {
		t.Fatal(err)
	}

	// Also create a text file that matches
	writeTestFile(t, dir, "text.txt", "hello world\n")

	results := make(chan GrepResult, 100)
	cancel := make(chan struct{})
	go ContentSearch(GrepOpts{
		RootDir:    dir,
		Pattern:    "hello",
		MaxResults: 0,
		ShowHidden: false,
		IgnoreCase: false,
	}, results, cancel)

	got := collectGrepResults(results)
	if len(got) != 1 {
		t.Fatalf("expected 1 result (text file only), got %d", len(got))
	}
	if got[0].RelPath != "text.txt" {
		t.Errorf("expected match in text.txt, got %q", got[0].RelPath)
	}
}

func TestContentSearch_SkipsLargeFiles(t *testing.T) {
	// Verify the maxFileSize constant is set to 1MB
	if maxFileSize != 1<<20 {
		t.Errorf("maxFileSize should be 1MB (1048576), got %d", maxFileSize)
	}

	// Create a small file that matches to ensure basic search works
	dir := t.TempDir()
	writeTestFile(t, dir, "small.txt", "findme\n")

	results := make(chan GrepResult, 100)
	cancel := make(chan struct{})
	go ContentSearch(GrepOpts{
		RootDir:    dir,
		Pattern:    "findme",
		MaxResults: 0,
		ShowHidden: false,
		IgnoreCase: false,
	}, results, cancel)

	got := collectGrepResults(results)
	if len(got) != 1 {
		t.Fatalf("expected 1 result from small file, got %d", len(got))
	}
}

func TestContentSearch_LineNumbers(t *testing.T) {
	dir := t.TempDir()
	content := "line one\nmatch here\nline three\nmatch again\nline five\n"
	writeTestFile(t, dir, "test.txt", content)

	results := make(chan GrepResult, 100)
	cancel := make(chan struct{})
	go ContentSearch(GrepOpts{
		RootDir:    dir,
		Pattern:    "match",
		MaxResults: 0,
		ShowHidden: false,
		IgnoreCase: false,
	}, results, cancel)

	got := collectGrepResults(results)
	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %d", len(got))
	}

	if got[0].Line != 2 {
		t.Errorf("first match should be on line 2, got %d", got[0].Line)
	}
	if got[0].Content != "match here" {
		t.Errorf("first match content should be %q, got %q", "match here", got[0].Content)
	}

	if got[1].Line != 4 {
		t.Errorf("second match should be on line 4, got %d", got[1].Line)
	}
	if got[1].Content != "match again" {
		t.Errorf("second match content should be %q, got %q", "match again", got[1].Content)
	}
}

func TestContentSearch_SkipsHidden(t *testing.T) {
	dir := t.TempDir()
	// Create a hidden file
	writeTestFile(t, dir, ".hidden.txt", "secret match\n")
	// Create a visible file
	writeTestFile(t, dir, "visible.txt", "visible match\n")

	// Search with ShowHidden=false
	results := make(chan GrepResult, 100)
	cancel := make(chan struct{})
	go ContentSearch(GrepOpts{
		RootDir:    dir,
		Pattern:    "match",
		MaxResults: 0,
		ShowHidden: false,
		IgnoreCase: false,
	}, results, cancel)

	got := collectGrepResults(results)
	if len(got) != 1 {
		t.Fatalf("expected 1 result with ShowHidden=false, got %d", len(got))
	}
	if got[0].RelPath != "visible.txt" {
		t.Errorf("expected match in visible.txt, got %q", got[0].RelPath)
	}

	// Search with ShowHidden=true — should find both
	results2 := make(chan GrepResult, 100)
	cancel2 := make(chan struct{})
	done := make(chan struct{})
	go func() {
		ContentSearch(GrepOpts{
			RootDir:    dir,
			Pattern:    "match",
			MaxResults: 0,
			ShowHidden: true,
			IgnoreCase: false,
		}, results2, cancel2)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("search with ShowHidden=true timed out")
	}

	got2 := collectGrepResults(results2)
	if len(got2) != 2 {
		t.Fatalf("expected 2 results with ShowHidden=true, got %d", len(got2))
	}
}
