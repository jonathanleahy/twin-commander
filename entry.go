package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileEntry represents a single directory entry with metadata.
type FileEntry struct {
	Name         string
	Size         int64
	ModTime      time.Time
	IsDir        bool
	IsSymlink    bool
	IsExecutable bool
	Accessible   bool
	Mode         os.FileMode
}

// ReadEntries reads directory contents and returns a slice of FileEntry values.
// If showHidden is false, entries starting with '.' are skipped.
// The returned slice is unsorted.
func ReadEntries(path string, showHidden bool) ([]FileEntry, error) {
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var entries []FileEntry
	for _, de := range dirEntries {
		name := de.Name()

		// Skip hidden files if not showing them
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}

		entry := FileEntry{
			Name:       name,
			Accessible: true,
		}

		isSymlink := de.Type()&fs.ModeSymlink != 0
		entry.IsSymlink = isSymlink

		if isSymlink {
			// Follow symlink for metadata
			targetInfo, statErr := os.Stat(filepath.Join(path, name))
			if statErr != nil {
				// Broken symlink
				entry.Accessible = false
				entry.Size = -1
				entry.IsDir = false
				entries = append(entries, entry)
				continue
			}
			entry.Size = targetInfo.Size()
			entry.ModTime = targetInfo.ModTime()
			entry.IsDir = targetInfo.IsDir()
			entry.Mode = targetInfo.Mode()
		} else {
			info, infoErr := de.Info()
			if infoErr != nil {
				entry.Accessible = false
				entry.Size = -1
				entries = append(entries, entry)
				continue
			}
			entry.Size = info.Size()
			entry.ModTime = info.ModTime()
			entry.IsDir = info.IsDir()
			entry.Mode = info.Mode()

			// Executable detection: regular files only
			if !info.IsDir() && info.Mode()&0111 != 0 {
				entry.IsExecutable = true
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// ReadEntriesRecursive walks a directory tree and returns all files with
// relative paths as names. Directories are not included in the output.
// Limited to 10,000 entries to avoid overwhelming the UI.
func ReadEntriesRecursive(root string, showHidden bool) ([]FileEntry, error) {
	const maxEntries = 10000
	var entries []FileEntry

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible dirs
		}

		name := d.Name()
		if !showHidden && strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip the root itself
		if path == root {
			return nil
		}

		// Skip directories — we only show files in flatten mode
		if d.IsDir() {
			return nil
		}

		rel, _ := filepath.Rel(root, path)

		entry := FileEntry{
			Name:       rel,
			Accessible: true,
			IsSymlink:  d.Type()&fs.ModeSymlink != 0,
		}

		info, infoErr := d.Info()
		if infoErr != nil {
			entry.Accessible = false
			entry.Size = -1
		} else {
			entry.Size = info.Size()
			entry.ModTime = info.ModTime()
			entry.Mode = info.Mode()
			if info.Mode()&0111 != 0 {
				entry.IsExecutable = true
			}
		}

		entries = append(entries, entry)
		if len(entries) >= maxEntries {
			return filepath.SkipAll
		}
		return nil
	})

	return entries, err
}

// SortEntries sorts entries with directories first, then files.
// Within each group, entries are sorted alphabetically (case-insensitive).
// Does NOT include ".." in the sort; ".." is prepended separately by Panel.
func SortEntries(entries []FileEntry) []FileEntry {
	sort.SliceStable(entries, func(i, j int) bool {
		// Directories first
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		// Then alphabetical (case-insensitive)
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	return entries
}
