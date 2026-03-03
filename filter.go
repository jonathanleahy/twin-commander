package main

import (
	"path/filepath"
	"regexp"
	"strings"
)

// FilterEntries returns entries matching the query.
// Supports glob patterns (*/?), regex (/pattern/), negation (!), multi-term (space-separated OR),
// and default case-insensitive substring matching.
// The ".." entry is always included regardless of the query.
// An empty query returns all entries unchanged.
func FilterEntries(entries []FileEntry, query string) []FileEntry {
	if query == "" {
		return entries
	}

	negate := false
	q := query
	if strings.HasPrefix(q, "!") {
		negate = true
		q = q[1:]
	}
	if q == "" {
		return entries
	}

	var matchFn func(name string) bool

	if strings.HasPrefix(q, "/") && strings.HasSuffix(q, "/") && len(q) > 2 {
		// Regex mode
		pattern := q[1 : len(q)-1]
		re, err := regexp.Compile(pattern)
		if err != nil {
			// Invalid regex — fall back to substring
			lq := strings.ToLower(q)
			matchFn = func(name string) bool {
				return strings.Contains(strings.ToLower(name), lq)
			}
		} else {
			matchFn = func(name string) bool {
				return re.MatchString(name)
			}
		}
	} else if strings.ContainsAny(q, "*?") {
		// Glob mode
		lowerPattern := strings.ToLower(q)
		matchFn = func(name string) bool {
			matched, _ := filepath.Match(lowerPattern, strings.ToLower(name))
			return matched
		}
	} else if strings.Contains(q, " ") {
		// Multi-term OR
		terms := strings.Fields(q)
		for i := range terms {
			terms[i] = strings.ToLower(terms[i])
		}
		matchFn = func(name string) bool {
			ln := strings.ToLower(name)
			for _, t := range terms {
				if strings.Contains(ln, t) {
					return true
				}
			}
			return false
		}
	} else {
		// Default: case-insensitive substring
		lq := strings.ToLower(q)
		matchFn = func(name string) bool {
			return strings.Contains(strings.ToLower(name), lq)
		}
	}

	var result []FileEntry
	for _, e := range entries {
		if e.Name == ".." {
			result = append(result, e)
			continue
		}
		matched := matchFn(e.Name)
		if negate {
			matched = !matched
		}
		if matched {
			result = append(result, e)
		}
	}
	return result
}
