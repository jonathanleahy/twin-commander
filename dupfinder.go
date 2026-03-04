package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DuplicateGroup represents a set of files with identical content.
type DuplicateGroup struct {
	Hash  string
	Size  int64
	Paths []string
}

// FindDuplicates scans a directory for duplicate files by size then MD5 hash.
// Returns groups of duplicates (2+ files with same content).
func FindDuplicates(root string, showHidden bool) []DuplicateGroup {
	// Phase 1: Group files by size (cheap filter)
	sizeMap := make(map[int64][]string)
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if !showHidden && strings.HasPrefix(d.Name(), ".") && path != root {
				return filepath.SkipDir
			}
			return nil
		}
		if !showHidden && strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Size() == 0 {
			return nil
		}
		sizeMap[info.Size()] = append(sizeMap[info.Size()], path)
		return nil
	})

	// Phase 2: For size groups with 2+ files, compute MD5
	hashMap := make(map[string][]string)
	for size, paths := range sizeMap {
		if len(paths) < 2 {
			continue
		}
		for _, p := range paths {
			hash, err := hashFile(p)
			if err != nil {
				continue
			}
			key := fmt.Sprintf("%d:%s", size, hash)
			hashMap[key] = append(hashMap[key], p)
		}
	}

	// Phase 3: Collect groups with 2+ files
	var groups []DuplicateGroup
	for key, paths := range hashMap {
		if len(paths) < 2 {
			continue
		}
		// Parse size from key
		var size int64
		fmt.Sscanf(key, "%d:", &size)
		groups = append(groups, DuplicateGroup{
			Hash:  key,
			Size:  size,
			Paths: paths,
		})
	}

	return groups
}

// FormatDuplicates formats duplicate groups for display.
func FormatDuplicates(groups []DuplicateGroup, root string) string {
	if len(groups) == 0 {
		return "No duplicates found."
	}

	var b strings.Builder
	totalWasted := int64(0)
	totalDups := 0

	for i, g := range groups {
		b.WriteString(fmt.Sprintf("[yellow]Group %d[-] (%s each, %d copies):\n", i+1, FormatSize(g.Size), len(g.Paths)))
		for _, p := range g.Paths {
			rel, _ := filepath.Rel(root, p)
			b.WriteString(fmt.Sprintf("  %s\n", rel))
		}
		totalWasted += g.Size * int64(len(g.Paths)-1)
		totalDups += len(g.Paths) - 1
		b.WriteString("\n")
	}

	b.WriteString(fmt.Sprintf("[yellow]Summary:[-] %d duplicate groups, %d extra copies, %s wasted\n",
		len(groups), totalDups, FormatSize(totalWasted)))

	return b.String()
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
