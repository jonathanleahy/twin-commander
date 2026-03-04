package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFuzzyScore_BasicMatch(t *testing.T) {
	score, matched := FuzzyScore("abc", "abc")
	if !matched {
		t.Fatal("expected match")
	}
	if score <= 0 {
		t.Errorf("expected positive score for exact match, got %d", score)
	}
}

func TestFuzzyScore_NoMatch(t *testing.T) {
	_, matched := FuzzyScore("xyz", "abc")
	if matched {
		t.Error("expected no match")
	}
}

func TestFuzzyScore_CaseInsensitive(t *testing.T) {
	_, matched := FuzzyScore("ABC", "abcdef")
	if !matched {
		t.Error("expected case-insensitive match")
	}
}

func TestFuzzyScore_EmptyPattern(t *testing.T) {
	score, matched := FuzzyScore("", "anything")
	if !matched {
		t.Error("empty pattern should match everything")
	}
	if score != 0 {
		t.Errorf("empty pattern should score 0, got %d", score)
	}
}

func TestFuzzyScore_ContiguousBeatsScattered(t *testing.T) {
	// "abc" contiguous in "abcdef" should beat "abc" scattered in "axbxcx"
	contiguousScore, _ := FuzzyScore("abc", "abcdef")
	scatteredScore, _ := FuzzyScore("abc", "axbxcx")
	if contiguousScore <= scatteredScore {
		t.Errorf("contiguous (%d) should beat scattered (%d)",
			contiguousScore, scatteredScore)
	}
}

func TestFuzzyScore_PrefixBeatsMiddle(t *testing.T) {
	// Match at start of filename should beat match in middle
	prefixScore, _ := FuzzyScore("main", "main.go")
	middleScore, _ := FuzzyScore("main", "src/domain/main.go")
	// Prefix gets filename prefix bonus AND shorter path
	if prefixScore <= middleScore {
		t.Errorf("prefix (%d) should beat middle (%d)", prefixScore, middleScore)
	}
}

func TestFuzzyScore_WordBoundary(t *testing.T) {
	// Match after word boundary should score higher
	boundaryScore, _ := FuzzyScore("test", "my_test_file")
	noBonus, _ := FuzzyScore("test", "mytestfile")
	if boundaryScore <= noBonus {
		t.Errorf("word boundary (%d) should beat no boundary (%d)",
			boundaryScore, noBonus)
	}
}

func TestFuzzyScore_OrderMatters(t *testing.T) {
	// Pattern chars must appear in order
	_, matched := FuzzyScore("cba", "abc")
	if matched {
		t.Error("pattern chars must appear in order")
	}
}

func TestFuzzySearch_EndToEnd(t *testing.T) {
	dir := t.TempDir()

	// Create file structure
	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "src", "app.go"), []byte("package src"), 0644)
	os.WriteFile(filepath.Join(dir, "src", "util.go"), []byte("package src"), 0644)
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# readme"), 0644)

	results := make(chan FuzzyResult, 100)
	cancel := make(chan struct{})

	go FuzzySearch(FuzzySearchOpts{
		RootDir:    dir,
		Pattern:    "main",
		MaxResults: 10,
		ShowHidden: false,
	}, results, cancel)

	var found []FuzzyResult
	for r := range results {
		found = append(found, r)
	}

	if len(found) == 0 {
		t.Fatal("expected at least one result")
	}

	// First result should be main.go (best match)
	if found[0].RelPath != "main.go" {
		t.Errorf("expected first result to be 'main.go', got %q", found[0].RelPath)
	}
}

func TestFuzzySearch_Cancellation(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)

	results := make(chan FuzzyResult, 100)
	cancel := make(chan struct{})
	close(cancel) // Cancel immediately

	FuzzySearch(FuzzySearchOpts{
		RootDir:    dir,
		Pattern:    "file",
		MaxResults: 10,
		ShowHidden: false,
	}, results, cancel)

	// Should complete without hanging
	count := 0
	for range results {
		count++
	}
	// May or may not find results depending on timing, but should not hang
}

func TestFuzzySearch_HiddenFileSkipping(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".hidden"), []byte("secret"), 0644)
	os.WriteFile(filepath.Join(dir, "visible"), []byte("hello"), 0644)

	// Without ShowHidden
	results := make(chan FuzzyResult, 100)
	cancel := make(chan struct{})
	go FuzzySearch(FuzzySearchOpts{
		RootDir:    dir,
		Pattern:    "hid",
		MaxResults: 10,
		ShowHidden: false,
	}, results, cancel)

	var found []FuzzyResult
	for r := range results {
		found = append(found, r)
	}
	if len(found) != 0 {
		t.Errorf("expected 0 results with ShowHidden=false, got %d", len(found))
	}

	// With ShowHidden
	results2 := make(chan FuzzyResult, 100)
	cancel2 := make(chan struct{})
	go FuzzySearch(FuzzySearchOpts{
		RootDir:    dir,
		Pattern:    "hid",
		MaxResults: 10,
		ShowHidden: true,
	}, results2, cancel2)

	var found2 []FuzzyResult
	for r := range results2 {
		found2 = append(found2, r)
	}
	if len(found2) == 0 {
		t.Error("expected results with ShowHidden=true")
	}
}

func TestFuzzySearch_EmptyPattern(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)

	results := make(chan FuzzyResult, 100)
	cancel := make(chan struct{})
	go FuzzySearch(FuzzySearchOpts{
		RootDir:    dir,
		Pattern:    "",
		MaxResults: 10,
		ShowHidden: false,
	}, results, cancel)

	var found []FuzzyResult
	for r := range results {
		found = append(found, r)
	}
	if len(found) != 0 {
		t.Errorf("expected 0 results for empty pattern, got %d", len(found))
	}
}

func TestFuzzySearch_MaxResults(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 20; i++ {
		os.WriteFile(filepath.Join(dir, filepath.Base(t.TempDir())+".txt"), []byte("data"), 0644)
	}
	// Create files that will match "t" pattern
	for i := 0; i < 20; i++ {
		name := "test" + string(rune('a'+i)) + ".txt"
		os.WriteFile(filepath.Join(dir, name), []byte("data"), 0644)
	}

	results := make(chan FuzzyResult, 100)
	cancel := make(chan struct{})
	go FuzzySearch(FuzzySearchOpts{
		RootDir:    dir,
		Pattern:    "test",
		MaxResults: 5,
		ShowHidden: false,
	}, results, cancel)

	var found []FuzzyResult
	for r := range results {
		found = append(found, r)
	}
	if len(found) > 5 {
		t.Errorf("expected at most 5 results, got %d", len(found))
	}
}

func TestFuzzySearch_DirectoryResults(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "mydir"), 0755)
	os.WriteFile(filepath.Join(dir, "myfile.txt"), []byte("data"), 0644)

	results := make(chan FuzzyResult, 100)
	cancel := make(chan struct{})
	go FuzzySearch(FuzzySearchOpts{
		RootDir:    dir,
		Pattern:    "my",
		MaxResults: 10,
		ShowHidden: false,
	}, results, cancel)

	var dirs, files int
	for r := range results {
		if r.IsDir {
			dirs++
		} else {
			files++
		}
	}
	if dirs == 0 {
		t.Error("expected at least one directory result")
	}
	if files == 0 {
		t.Error("expected at least one file result")
	}
}

func TestFuzzySearch_DirsOnly(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "mydir"), 0755)
	os.MkdirAll(filepath.Join(dir, "other"), 0755)
	os.WriteFile(filepath.Join(dir, "myfile.txt"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(dir, "mydir", "inner.txt"), []byte("data"), 0644)

	results := make(chan FuzzyResult, 100)
	cancel := make(chan struct{})
	go FuzzySearch(FuzzySearchOpts{
		RootDir:    dir,
		Pattern:    "my",
		MaxResults: 10,
		ShowHidden: false,
		DirsOnly:   true,
	}, results, cancel)

	var found []FuzzyResult
	for r := range results {
		found = append(found, r)
	}

	if len(found) == 0 {
		t.Fatal("expected at least one directory result")
	}
	for _, r := range found {
		if !r.IsDir {
			t.Errorf("expected only directories with DirsOnly=true, got file: %s", r.RelPath)
		}
	}
	// "mydir" should be in the results
	foundMyDir := false
	for _, r := range found {
		if r.RelPath == "mydir" {
			foundMyDir = true
		}
	}
	if !foundMyDir {
		t.Error("expected 'mydir' in results")
	}
}
