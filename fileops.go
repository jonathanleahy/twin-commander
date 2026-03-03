package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// PathExists returns true if the path exists.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CopyFile copies a single file from src to dst.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// CopyDir recursively copies a directory from src to dst.
func CopyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.Type()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(srcPath)
			if err != nil {
				return err
			}
			if err := os.Symlink(linkTarget, dstPath); err != nil {
				return err
			}
		} else if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// MoveFile moves a file or directory from src to dst using os.Rename.
// Falls back to copy+remove when src and dst are on different filesystems.
func MoveFile(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	linkErr, ok := err.(*os.LinkError)
	if !ok || linkErr.Err != syscall.EXDEV {
		return err
	}

	// Cross-filesystem move: copy then remove.
	info, statErr := os.Lstat(src)
	if statErr != nil {
		return statErr
	}

	if info.IsDir() {
		if copyErr := CopyDir(src, dst); copyErr != nil {
			return copyErr
		}
		return os.RemoveAll(src)
	}

	if copyErr := CopyFile(src, dst); copyErr != nil {
		return copyErr
	}
	return os.Remove(src)
}

// DeletePath permanently removes a file or directory.
func DeletePath(path string) error {
	return os.RemoveAll(path)
}

// MoveToTrash moves a file to the XDG trash directory.
func MoveToTrash(path string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot find home directory: %w", err)
	}

	trashDir := filepath.Join(home, ".local", "share", "Trash", "files")
	if err := os.MkdirAll(trashDir, 0o755); err != nil {
		return fmt.Errorf("cannot create trash directory: %w", err)
	}

	base := filepath.Base(path)
	dst := filepath.Join(trashDir, base)

	// Handle name collisions
	finalName := base
	i := 1
	for PathExists(dst) {
		ext := filepath.Ext(base)
		name := base[:len(base)-len(ext)]
		finalName = fmt.Sprintf("%s.%d%s", name, i, ext)
		dst = filepath.Join(trashDir, finalName)
		i++
	}

	// Create the .trashinfo metadata file
	infoDir := filepath.Join(home, ".local", "share", "Trash", "info")
	if err := os.MkdirAll(infoDir, 0o755); err != nil {
		return fmt.Errorf("cannot create trash info directory: %w", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("cannot resolve absolute path: %w", err)
	}

	trashInfoPath := filepath.Join(infoDir, finalName+".trashinfo")
	trashInfoContent := fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		url.PathEscape(absPath),
		time.Now().Format("2006-01-02T15:04:05"),
	)
	if err := os.WriteFile(trashInfoPath, []byte(trashInfoContent), 0o644); err != nil {
		return fmt.Errorf("cannot write trashinfo file: %w", err)
	}

	return os.Rename(path, dst)
}

// MakeDirSafe creates a directory with standard permissions.
func MakeDirSafe(path string) error {
	return os.MkdirAll(path, 0o755)
}

// RenamePath renames a file or directory (newName is the base name only).
func RenamePath(oldPath, newName string) error {
	dir := filepath.Dir(oldPath)
	newPath := filepath.Join(dir, newName)
	return os.Rename(oldPath, newPath)
}

// CalcDirSize recursively calculates the total size of a directory.
func CalcDirSize(path string) (int64, error) {
	var total int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}
