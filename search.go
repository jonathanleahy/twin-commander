package main

import (
	"os"
	"path/filepath"
	"strings"
)

// SearchOpts configures a recursive file search.
type SearchOpts struct {
	RootDir    string
	Query      string
	MaxResults int
	ShowHidden bool
}

// SearchResult represents a single search match.
type SearchResult struct {
	Path    string // Absolute path
	RelPath string // Path relative to search root
	IsDir   bool
}

// Search performs a recursive filename search, sending results to the channel.
// The search respects the cancel channel and stops when it's closed.
func Search(opts SearchOpts, results chan<- SearchResult, cancel <-chan struct{}) {
	defer close(results)

	lowerQuery := strings.ToLower(opts.Query)
	count := 0

	_ = filepath.Walk(opts.RootDir, func(path string, info os.FileInfo, err error) error {
		// Check for cancellation
		select {
		case <-cancel:
			return filepath.SkipAll
		default:
		}

		if err != nil {
			return nil
		}

		name := info.Name()

		// Skip hidden files/dirs
		if !opts.ShowHidden && strings.HasPrefix(name, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip the root itself
		if path == opts.RootDir {
			return nil
		}

		// Match against query
		if !strings.Contains(strings.ToLower(name), lowerQuery) {
			return nil
		}

		relPath, _ := filepath.Rel(opts.RootDir, path)

		select {
		case results <- SearchResult{
			Path:    path,
			RelPath: relPath,
			IsDir:   info.IsDir(),
		}:
		case <-cancel:
			return filepath.SkipAll
		}

		count++
		if opts.MaxResults > 0 && count >= opts.MaxResults {
			return filepath.SkipAll
		}

		return nil
	})
}
