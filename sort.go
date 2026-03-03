package main

import (
	"path/filepath"
	"sort"
	"strings"
)

// SortMode determines the primary sort key for file entries.
type SortMode int

const (
	SortByName      SortMode = iota
	SortBySize
	SortByDate
	SortByExtension
)

// SortOrder determines ascending or descending sort direction.
type SortOrder int

const (
	SortAsc  SortOrder = iota
	SortDesc
)

// NextSortMode cycles to the next sort mode.
func NextSortMode(m SortMode) SortMode {
	return (m + 1) % 4
}

// ToggleSortOrder flips between ascending and descending.
func ToggleSortOrder(o SortOrder) SortOrder {
	if o == SortAsc {
		return SortDesc
	}
	return SortAsc
}

// SortModeLabel returns a short label for the sort mode.
func SortModeLabel(m SortMode) string {
	switch m {
	case SortBySize:
		return "size"
	case SortByDate:
		return "date"
	case SortByExtension:
		return "ext"
	default:
		return "name"
	}
}

// SortOrderArrow returns an arrow indicator for the sort direction.
func SortOrderArrow(o SortOrder) string {
	if o == SortDesc {
		return "↓"
	}
	return "↑"
}

// SortEntriesBy sorts entries with directories first, then by the given mode and order.
func SortEntriesBy(entries []FileEntry, mode SortMode, order SortOrder) []FileEntry {
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]

		// Directories always come first
		if a.IsDir != b.IsDir {
			return a.IsDir
		}

		var less bool
		switch mode {
		case SortBySize:
			less = a.Size < b.Size
		case SortByDate:
			less = a.ModTime.Before(b.ModTime)
		case SortByExtension:
			extA := strings.ToLower(filepath.Ext(a.Name))
			extB := strings.ToLower(filepath.Ext(b.Name))
			if extA == extB {
				less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
			} else {
				less = extA < extB
			}
		default: // SortByName
			less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
		}

		if order == SortDesc {
			return !less
		}
		return less
	})
	return entries
}
