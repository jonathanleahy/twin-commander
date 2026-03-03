package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// GrepResult represents a single content search match within a file.
type GrepResult struct {
	Path    string // Absolute path
	RelPath string // Relative to search root
	Line    int    // Line number (1-based)
	Content string // Matching line content (trimmed)
}

// GrepOpts configures a recursive content search.
type GrepOpts struct {
	RootDir    string
	Pattern    string
	MaxResults int
	ShowHidden bool
	IgnoreCase bool
}

// maxFileSize is the maximum file size (1MB) that ContentSearch will read.
const maxFileSize = 1 << 20

// ContentSearch walks the directory tree searching file contents for the pattern.
// Skips binary files (files containing null bytes in first 512 bytes).
// Skips files larger than 1MB.
// Sends results through the channel, respects cancel.
func ContentSearch(opts GrepOpts, results chan<- GrepResult, cancel <-chan struct{}) {
	defer close(results)

	pattern := opts.Pattern
	if opts.IgnoreCase {
		pattern = strings.ToLower(pattern)
	}
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

		// Skip the root itself and directories
		if path == opts.RootDir {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		// Skip files larger than 1MB
		if info.Size() > maxFileSize {
			return nil
		}

		// Skip non-regular files
		if !info.Mode().IsRegular() {
			return nil
		}

		// Open and check for binary content
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		// Read first 512 bytes to detect binary files
		header := make([]byte, 512)
		n, err := f.Read(header)
		if err != nil && n == 0 {
			return nil
		}
		for i := 0; i < n; i++ {
			if header[i] == 0 {
				return nil // Binary file, skip
			}
		}

		// Seek back to the beginning
		if _, err := f.Seek(0, 0); err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(opts.RootDir, path)

		// Scan line by line
		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			// Check for cancellation
			select {
			case <-cancel:
				return filepath.SkipAll
			default:
			}

			lineNum++
			line := scanner.Text()

			matchLine := line
			if opts.IgnoreCase {
				matchLine = strings.ToLower(line)
			}

			if !strings.Contains(matchLine, pattern) {
				continue
			}

			select {
			case results <- GrepResult{
				Path:    path,
				RelPath: relPath,
				Line:    lineNum,
				Content: strings.TrimSpace(line),
			}:
			case <-cancel:
				return filepath.SkipAll
			}

			count++
			if opts.MaxResults > 0 && count >= opts.MaxResults {
				return filepath.SkipAll
			}
		}

		return nil
	})
}
