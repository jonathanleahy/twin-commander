package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// FuzzyResult represents a single fuzzy search match.
type FuzzyResult struct {
	Path    string // Absolute path
	RelPath string // Path relative to search root
	IsDir   bool
	Score   int
}

// FuzzySearchOpts configures a fuzzy file search.
type FuzzySearchOpts struct {
	RootDir    string
	Pattern    string
	MaxResults int
	ShowHidden bool
}

// FuzzyScore scores how well pattern matches text using fuzzy matching.
// All pattern characters must appear in text in order (case-insensitive).
// Returns (score, matched). Score is 0 for no bonus, higher is better.
func FuzzyScore(pattern, text string) (int, bool) {
	if pattern == "" {
		return 0, true
	}

	lowerPattern := strings.ToLower(pattern)
	lowerText := strings.ToLower(text)

	// Check if all pattern chars appear in order
	pi := 0
	matchPositions := make([]int, 0, len(lowerPattern))
	for ti := 0; ti < len(lowerText) && pi < len(lowerPattern); ti++ {
		if lowerText[ti] == lowerPattern[pi] {
			matchPositions = append(matchPositions, ti)
			pi++
		}
	}
	if pi < len(lowerPattern) {
		return 0, false
	}

	score := 0

	// Contiguous match bonus (escalating)
	contiguous := 0
	for i := 1; i < len(matchPositions); i++ {
		if matchPositions[i] == matchPositions[i-1]+1 {
			contiguous++
			score += contiguous * 3 // escalating: 3, 6, 9, ...
		} else {
			contiguous = 0
		}
	}

	// Word-boundary bonus: matched char right after / . _ -
	for _, pos := range matchPositions {
		if pos == 0 {
			score += 5
		} else {
			prev := rune(text[pos-1])
			if prev == '/' || prev == '.' || prev == '_' || prev == '-' {
				score += 5
			}
		}
	}

	// Exact case bonus
	for i, pos := range matchPositions {
		if text[pos] == pattern[i] {
			score += 1
		}
	}

	// Filename prefix bonus: if match starts at beginning of filename
	lastSlash := strings.LastIndex(text, "/")
	filenameStart := lastSlash + 1
	if len(matchPositions) > 0 && matchPositions[0] == filenameStart {
		score += 10
	}

	// Path length penalty: shorter paths rank higher
	score -= len(text) / 10

	return score, true
}

// FuzzySearch walks the filesystem from opts.RootDir, scores each file against
// the pattern, and sends the top results through the results channel.
func FuzzySearch(opts FuzzySearchOpts, results chan<- FuzzyResult, cancel <-chan struct{}) {
	defer close(results)

	if opts.Pattern == "" {
		return
	}

	var all []FuzzyResult

	_ = filepath.Walk(opts.RootDir, func(path string, info os.FileInfo, err error) error {
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

		relPath, _ := filepath.Rel(opts.RootDir, path)
		score, matched := FuzzyScore(opts.Pattern, relPath)
		if !matched {
			return nil
		}

		all = append(all, FuzzyResult{
			Path:    path,
			RelPath: relPath,
			IsDir:   info.IsDir(),
			Score:   score,
		})

		return nil
	})

	// Sort by score descending
	sort.Slice(all, func(i, j int) bool {
		if all[i].Score != all[j].Score {
			return all[i].Score > all[j].Score
		}
		return all[i].RelPath < all[j].RelPath
	})

	max := opts.MaxResults
	if max <= 0 || max > len(all) {
		max = len(all)
	}

	for i := 0; i < max; i++ {
		select {
		case results <- all[i]:
		case <-cancel:
			return
		}
	}
}

// isWordBoundary checks if a rune is a word separator.
func isWordBoundary(r rune) bool {
	return r == '/' || r == '.' || r == '_' || r == '-' || unicode.IsSpace(r)
}
