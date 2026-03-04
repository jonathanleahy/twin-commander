package main

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// IsArchive returns true if the filename has a recognized archive extension.
func IsArchive(name string) bool {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return true
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return true
	case strings.HasSuffix(lower, ".tar.bz2"), strings.HasSuffix(lower, ".tbz2"):
		return true
	case strings.HasSuffix(lower, ".tar"):
		return true
	}
	return false
}

// ArchiveEntry represents one entry inside an archive.
type ArchiveEntry struct {
	Path string
	Size int64
	Dir  bool
}

// ListArchive returns a listing of entries inside a zip or tar archive.
func ListArchive(path string) ([]ArchiveEntry, error) {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return listZip(path)
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return listTarGz(path)
	case strings.HasSuffix(lower, ".tar.bz2"), strings.HasSuffix(lower, ".tbz2"):
		return listTarBz2(path)
	case strings.HasSuffix(lower, ".tar"):
		return listTar(path)
	default:
		return nil, fmt.Errorf("unsupported archive format")
	}
}

// FormatArchiveListing formats archive entries as a human-readable string.
func FormatArchiveListing(entries []ArchiveEntry, archiveName string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Archive: %s\n", archiveName))
	b.WriteString(fmt.Sprintf("Entries: %d\n", len(entries)))
	b.WriteString(strings.Repeat("─", 70))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%-50s %10s\n", "Name", "Size"))
	b.WriteString(strings.Repeat("─", 70))
	b.WriteString("\n")

	for _, e := range entries {
		name := e.Path
		if e.Dir {
			name += "/"
		}
		if len(name) > 50 {
			name = "..." + name[len(name)-47:]
		}
		b.WriteString(fmt.Sprintf("%-50s %10s\n", name, FormatSize(e.Size)))
	}

	// Total size
	var total int64
	for _, e := range entries {
		total += e.Size
	}
	b.WriteString(strings.Repeat("─", 70))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%-50s %10s\n", "Total (uncompressed)", FormatSize(total)))

	// Archive file size
	if info, err := os.Stat(filepath.Clean(archiveName)); err == nil {
		b.WriteString(fmt.Sprintf("%-50s %10s\n", "Archive size", FormatSize(info.Size())))
	}

	return b.String()
}

func listZip(path string) ([]ArchiveEntry, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var entries []ArchiveEntry
	for _, f := range r.File {
		entries = append(entries, ArchiveEntry{
			Path: f.Name,
			Size: int64(f.UncompressedSize64),
			Dir:  f.FileInfo().IsDir(),
		})
	}
	return entries, nil
}

func listTar(path string) ([]ArchiveEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return readTar(f)
}

func listTarGz(path string) ([]ArchiveEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	return readTar(gz)
}

func listTarBz2(path string) ([]ArchiveEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bz := bzip2.NewReader(f)
	return readTar(bz)
}

func readTar(r io.Reader) ([]ArchiveEntry, error) {
	tr := tar.NewReader(r)
	var entries []ArchiveEntry
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return entries, nil // return what we have so far
		}
		entries = append(entries, ArchiveEntry{
			Path: hdr.Name,
			Size: hdr.Size,
			Dir:  hdr.Typeflag == tar.TypeDir,
		})
	}
	return entries, nil
}
