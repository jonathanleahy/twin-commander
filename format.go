package main

import "fmt"

// FormatSize converts a byte count to a human-readable string.
// Negative values (inaccessible entries) return "---".
func FormatSize(bytes int64) string {
	if bytes < 0 {
		return "---"
	}
	if bytes < 1024 {
		return fmt.Sprintf("%d", bytes)
	}
	if bytes < 1048576 {
		return fmt.Sprintf("%.1fK", float64(bytes)/1024.0)
	}
	if bytes < 1073741824 {
		return fmt.Sprintf("%.1fM", float64(bytes)/1048576.0)
	}
	return fmt.Sprintf("%.1fG", float64(bytes)/1073741824.0)
}
