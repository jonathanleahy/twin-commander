package main

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// GitRepo represents a detected git repository with cached status information.
type GitRepo struct {
	RootDir string
	Branch  string
	status  map[string]string // relative path -> status code (e.g. "M", "A", "?")
}

// DetectGitRepo tries to find a git repository containing the given directory.
// Returns nil if not in a git repo or git is not available.
func DetectGitRepo(dir string) *GitRepo {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	root := strings.TrimSpace(string(out))
	if root == "" {
		return nil
	}

	repo := &GitRepo{RootDir: root}
	repo.Refresh()
	return repo
}

// Refresh reloads git status and branch information.
func (g *GitRepo) Refresh() {
	g.status = make(map[string]string)

	// Get branch
	cmd := exec.Command("git", "-C", g.RootDir, "branch", "--show-current")
	out, err := cmd.Output()
	if err == nil {
		g.Branch = strings.TrimSpace(string(out))
	}

	// Get status
	cmd = exec.Command("git", "-C", g.RootDir, "status", "--porcelain=v1")
	out, err = cmd.Output()
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(out), "\n") {
		if len(line) < 4 {
			continue
		}
		xy := line[:2]
		path := strings.TrimSpace(line[3:])
		// Handle renamed files
		if idx := strings.Index(path, " -> "); idx >= 0 {
			path = path[idx+4:]
		}
		g.status[path] = xy
	}
}

// RelPath returns the path relative to the git root.
func (g *GitRepo) RelPath(absPath string) string {
	rel, err := filepath.Rel(g.RootDir, absPath)
	if err != nil {
		return absPath
	}
	return rel
}

// GetFileStatus returns the git status code for a file.
func (g *GitRepo) GetFileStatus(relPath string) string {
	return g.status[relPath]
}

// GetDirStatus returns an aggregate status for a directory.
// Returns the "most important" status found in any file under the directory.
func (g *GitRepo) GetDirStatus(relDir string) string {
	prefix := relDir + "/"
	best := ""
	for path, status := range g.status {
		if strings.HasPrefix(path, prefix) || path == relDir {
			if statusPriority(status) > statusPriority(best) {
				best = status
			}
		}
	}
	return best
}

// GetDiff returns the git diff output for a file.
func (g *GitRepo) GetDiff(relPath string) (string, error) {
	cmd := exec.Command("git", "-C", g.RootDir, "diff", "--", relPath)
	out, err := cmd.Output()
	if err != nil {
		// Try staged diff
		cmd = exec.Command("git", "-C", g.RootDir, "diff", "--cached", "--", relPath)
		out, err = cmd.Output()
		if err != nil {
			return "", err
		}
	}
	return string(out), nil
}

// ToggleStaged stages or unstages a file.
func (g *GitRepo) ToggleStaged(relPath string) error {
	status := g.status[relPath]
	if len(status) >= 2 && status[0] != ' ' && status[0] != '?' {
		// File is staged — unstage it
		return exec.Command("git", "-C", g.RootDir, "reset", "HEAD", "--", relPath).Run()
	}
	// File is not staged — stage it
	return exec.Command("git", "-C", g.RootDir, "add", "--", relPath).Run()
}

// GitStatusColor returns the color for a git status code.
// Returns (color, true) if a color should be applied, (0, false) otherwise.
func GitStatusColor(status string) (tcell.Color, bool) {
	if len(status) < 2 {
		return 0, false
	}
	// Index (staged) takes priority
	switch status[0] {
	case 'M':
		return tcell.ColorOrange, true
	case 'A':
		return tcell.ColorGreen, true
	case 'D':
		return tcell.ColorRed, true
	case 'R':
		return tcell.ColorYellow, true
	case 'C':
		return tcell.ColorYellow, true
	}
	// Worktree (unstaged)
	switch status[1] {
	case 'M':
		return tcell.ColorOrange, true
	case 'D':
		return tcell.ColorRed, true
	}
	// Untracked
	if status == "??" {
		return tcell.ColorGray, true
	}
	return 0, false
}

// GitStatusLabel returns a short human-readable label for a git status code.
func GitStatusLabel(status string) string {
	if len(status) < 2 {
		return ""
	}
	switch {
	case status == "??":
		return "?"
	case status[0] == 'A':
		return "A"
	case status[0] == 'M' || status[1] == 'M':
		return "M"
	case status[0] == 'D' || status[1] == 'D':
		return "D"
	case status[0] == 'R':
		return "R"
	case status[0] == 'C':
		return "C"
	default:
		return string(status[0])
	}
}

// statusPriority returns a priority score for sorting statuses.
func statusPriority(status string) int {
	if len(status) < 2 {
		return 0
	}
	switch {
	case status == "??":
		return 1
	case status[1] == 'M':
		return 2
	case status[0] == 'M':
		return 3
	case status[0] == 'A':
		return 4
	case status[0] == 'D' || status[1] == 'D':
		return 5
	default:
		return 1
	}
}
