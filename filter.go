package main

import "strings"

// FilterEntries returns entries matching the query (case-insensitive substring).
// The ".." entry is always included regardless of the query.
// An empty query returns all entries unchanged.
func FilterEntries(entries []FileEntry, query string) []FileEntry {
	if query == "" {
		return entries
	}
	lowerQuery := strings.ToLower(query)
	var result []FileEntry
	for _, e := range entries {
		if e.Name == ".." || strings.Contains(strings.ToLower(e.Name), lowerQuery) {
			result = append(result, e)
		}
	}
	return result
}
