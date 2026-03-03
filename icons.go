package main

import (
	"path/filepath"
	"strings"
)

// iconDotDot is the icon prefix for the parent directory entry.
const iconDotDot = "↩ "

// FileIcon returns a Nerd Font icon for the given filename.
// Parameters indicate whether it's a directory, symlink, or executable.
func FileIcon(name string, isDir, isSymlink, isExec bool) string {
	if isSymlink {
		return " "
	}
	if isDir {
		return dirIcon(name) + " "
	}
	if isExec {
		return " "
	}
	return fileIcon(name) + " "
}

// dirIcon returns a Nerd Font icon for a directory name.
func dirIcon(name string) string {
	lower := strings.ToLower(name)
	switch lower {
	case ".git":
		return ""
	case "node_modules":
		return ""
	case "src", "source":
		return ""
	case "test", "tests", "__tests__":
		return ""
	case "docs", "doc", "documentation":
		return ""
	case "build", "dist", "out", "target":
		return ""
	case "vendor", "third_party":
		return ""
	case ".github":
		return ""
	case ".vscode":
		return ""
	case "config", "conf":
		return ""
	default:
		return ""
	}
}

// fileIcon returns a Nerd Font icon based on file extension or name.
func fileIcon(name string) string {
	lower := strings.ToLower(name)

	// Special filenames
	switch lower {
	case "makefile", "cmakelists.txt":
		return ""
	case "dockerfile":
		return ""
	case "docker-compose.yml", "docker-compose.yaml":
		return ""
	case ".gitignore", ".gitattributes":
		return ""
	case "license", "licence":
		return ""
	case "readme", "readme.md", "readme.txt":
		return ""
	}

	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	// Programming languages
	case ".go":
		return ""
	case ".rs":
		return ""
	case ".py":
		return ""
	case ".js":
		return ""
	case ".ts":
		return ""
	case ".jsx", ".tsx":
		return ""
	case ".java":
		return ""
	case ".c", ".h":
		return ""
	case ".cpp", ".hpp", ".cc", ".cxx":
		return ""
	case ".cs":
		return ""
	case ".rb":
		return ""
	case ".php":
		return ""
	case ".swift":
		return ""
	case ".kt", ".kts":
		return ""
	case ".lua":
		return ""
	case ".sh", ".bash", ".zsh", ".fish":
		return ""
	case ".ps1":
		return ""

	// Web
	case ".html", ".htm":
		return ""
	case ".css":
		return ""
	case ".scss", ".sass", ".less":
		return ""
	case ".svg":
		return ""

	// Data / Config
	case ".json":
		return ""
	case ".yaml", ".yml":
		return ""
	case ".toml":
		return ""
	case ".xml":
		return ""
	case ".ini", ".cfg":
		return ""
	case ".env":
		return ""
	case ".sql":
		return ""
	case ".graphql", ".gql":
		return ""

	// Documents
	case ".md", ".markdown":
		return ""
	case ".txt":
		return ""
	case ".pdf":
		return ""
	case ".doc", ".docx":
		return ""

	// Images
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".webp":
		return ""

	// Audio/Video
	case ".mp3", ".wav", ".flac", ".ogg", ".aac":
		return ""
	case ".mp4", ".avi", ".mkv", ".mov", ".webm":
		return ""

	// Archives
	case ".zip", ".tar", ".gz", ".bz2", ".xz", ".rar", ".7z":
		return ""

	// Lock files
	case ".lock":
		return ""

	// Binary/Compiled
	case ".o", ".so", ".dylib", ".dll", ".exe":
		return ""
	case ".wasm":
		return ""

	default:
		return ""
	}
}
