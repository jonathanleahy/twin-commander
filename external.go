package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/rivo/tview"
)

// BCompAvailable returns true if Beyond Compare (bcomp) is in the PATH.
func BCompAvailable() bool {
	_, err := exec.LookPath("bcomp")
	return err == nil
}

// LaunchBComp launches Beyond Compare with the given left and right paths.
// Suspends the TUI while bcomp runs.
func LaunchBComp(app *tview.Application, left, right string) error {
	app.Suspend(func() {
		cmd := exec.Command("bcomp", left, right)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
	})
	return nil
}

// OpenInEditor opens a file in the user's preferred editor.
// Uses $EDITOR, falling back to "vi".
func OpenInEditor(app *tview.Application, path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	var err error
	app.Suspend(func() {
		cmd := exec.Command(editor, path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
	})
	return err
}

// CopyToClipboard copies text to the system clipboard using xclip or xsel.
func CopyToClipboard(text string) error {
	// Try xclip first
	if path, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command(path, "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	}

	// Try xsel
	if path, err := exec.LookPath("xsel"); err == nil {
		cmd := exec.Command(path, "--clipboard", "--input")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	}

	// Try wl-copy (Wayland)
	if path, err := exec.LookPath("wl-copy"); err == nil {
		cmd := exec.Command(path)
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	}

	return exec.ErrNotFound
}

// HasNerdFont checks if a Nerd Font is installed by looking at fc-list output.
func HasNerdFont() bool {
	cmd := exec.Command("fc-list", ":family")
	out, err := cmd.Output()
	if err != nil {
		return true // Assume yes if we can't check
	}
	lower := strings.ToLower(string(out))
	return strings.Contains(lower, "nerd") ||
		strings.Contains(lower, "firacode") ||
		strings.Contains(lower, "jetbrains") ||
		strings.Contains(lower, "hack") ||
		strings.Contains(lower, "cascadia")
}
