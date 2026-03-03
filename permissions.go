package main

import (
	"fmt"
	"os"
	"strconv"
)

// FormatPermissions converts a file mode to a "rwxr-xr-x" style string.
// Only the lower 9 permission bits (user/group/other rwx) are considered.
func FormatPermissions(mode os.FileMode) string {
	var buf [9]byte
	const rwx = "rwx"
	perm := mode.Perm()
	for i := 0; i < 9; i++ {
		if perm&(1<<uint(8-i)) != 0 {
			buf[i] = rwx[i%3]
		} else {
			buf[i] = '-'
		}
	}
	return string(buf[:])
}

// ParseOctalMode parses an octal permission string like "755" into an os.FileMode.
// Accepts 3 or 4 digit strings where each digit is 0-7.
func ParseOctalMode(s string) (os.FileMode, error) {
	if len(s) == 0 {
		return 0, fmt.Errorf("empty permission string")
	}
	if len(s) < 3 || len(s) > 4 {
		return 0, fmt.Errorf("invalid permission string %q: must be 3 or 4 digits", s)
	}
	for _, c := range s {
		if c < '0' || c > '7' {
			if c >= '0' && c <= '9' {
				return 0, fmt.Errorf("invalid octal digit %c in %q", c, s)
			}
			return 0, fmt.Errorf("invalid character %c in permission string %q", c, s)
		}
	}
	val, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid permission string %q: %w", s, err)
	}
	return os.FileMode(val), nil
}

// ChmodPath changes the permissions of the file at path.
func ChmodPath(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}
