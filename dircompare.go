package main

import (
	"fmt"
	"strings"
)

// DiffKind describes how a file differs between two directories.
type DiffKind int

const (
	DiffLeftOnly  DiffKind = iota // exists only in left
	DiffRightOnly                 // exists only in right
	DiffSizeDiff                  // exists in both, different size
	DiffDateDiff                  // exists in both, same size, different date
)

// DirDiffEntry represents one difference between two directories.
type DirDiffEntry struct {
	Name     string
	Kind     DiffKind
	LeftSize int64
	RightSize int64
}

// CompareDirs compares entries from two panels and returns differences.
func CompareDirs(left, right []FileEntry) []DirDiffEntry {
	leftMap := make(map[string]FileEntry)
	rightMap := make(map[string]FileEntry)

	for _, e := range left {
		if e.Name != ".." {
			leftMap[e.Name] = e
		}
	}
	for _, e := range right {
		if e.Name != ".." {
			rightMap[e.Name] = e
		}
	}

	var diffs []DirDiffEntry

	// Check left entries
	for name, le := range leftMap {
		re, inRight := rightMap[name]
		if !inRight {
			diffs = append(diffs, DirDiffEntry{Name: name, Kind: DiffLeftOnly, LeftSize: le.Size})
			continue
		}
		if le.Size != re.Size {
			diffs = append(diffs, DirDiffEntry{Name: name, Kind: DiffSizeDiff, LeftSize: le.Size, RightSize: re.Size})
			continue
		}
		if !le.ModTime.Equal(re.ModTime) {
			diffs = append(diffs, DirDiffEntry{Name: name, Kind: DiffDateDiff, LeftSize: le.Size, RightSize: re.Size})
		}
	}

	// Check right-only entries
	for name, re := range rightMap {
		if _, inLeft := leftMap[name]; !inLeft {
			diffs = append(diffs, DirDiffEntry{Name: name, Kind: DiffRightOnly, RightSize: re.Size})
		}
	}

	return diffs
}

// FormatDirComparison formats directory comparison results as a colored string.
func FormatDirComparison(diffs []DirDiffEntry, leftDir, rightDir string) string {
	if len(diffs) == 0 {
		return "Directories are identical."
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("[yellow]Left:[-]  %s\n", leftDir))
	b.WriteString(fmt.Sprintf("[yellow]Right:[-] %s\n", rightDir))
	b.WriteString(fmt.Sprintf("[yellow]Differences:[-] %d\n", len(diffs)))
	b.WriteString(strings.Repeat("─", 65))
	b.WriteString("\n")

	var leftOnly, rightOnly, sizeDiff, dateDiff []DirDiffEntry
	for _, d := range diffs {
		switch d.Kind {
		case DiffLeftOnly:
			leftOnly = append(leftOnly, d)
		case DiffRightOnly:
			rightOnly = append(rightOnly, d)
		case DiffSizeDiff:
			sizeDiff = append(sizeDiff, d)
		case DiffDateDiff:
			dateDiff = append(dateDiff, d)
		}
	}

	if len(leftOnly) > 0 {
		b.WriteString("\n[red]◀ Left only:[-]\n")
		for _, d := range leftOnly {
			b.WriteString(fmt.Sprintf("  %s (%s)\n", d.Name, FormatSize(d.LeftSize)))
		}
	}

	if len(rightOnly) > 0 {
		b.WriteString("\n[green]▶ Right only:[-]\n")
		for _, d := range rightOnly {
			b.WriteString(fmt.Sprintf("  %s (%s)\n", d.Name, FormatSize(d.RightSize)))
		}
	}

	if len(sizeDiff) > 0 {
		b.WriteString("\n[yellow]≠ Different size:[-]\n")
		for _, d := range sizeDiff {
			b.WriteString(fmt.Sprintf("  %s (left: %s, right: %s)\n", d.Name, FormatSize(d.LeftSize), FormatSize(d.RightSize)))
		}
	}

	if len(dateDiff) > 0 {
		b.WriteString("\n[blue]⏰ Different date (same size):[-]\n")
		for _, d := range dateDiff {
			b.WriteString(fmt.Sprintf("  %s\n", d.Name))
		}
	}

	return b.String()
}
