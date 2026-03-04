package main

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/rivo/tview"
)

// IsDarwin returns true if the current OS is macOS.
func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

// ModifierLabel returns "Opt" on macOS and "Alt" on Linux/Windows.
func ModifierLabel() string {
	if IsDarwin() {
		return "Opt"
	}
	return "Alt"
}

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
// Priority: editorOverride (if non-empty) → $EDITOR → "vi".
func OpenInEditor(app *tview.Application, path string, editorOverride string) error {
	editor := editorOverride
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
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

// OpenWithDefault opens a file with the system's default application.
// Uses xdg-open on Linux, open on macOS.
func OpenWithDefault(path string) error {
	var cmd *exec.Cmd
	if IsDarwin() {
		cmd = exec.Command("open", path)
	} else {
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start() // Don't wait — let the app open in background
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

// ReadFromClipboard reads text from the system clipboard.
func ReadFromClipboard() (string, error) {
	// Try xclip first
	if path, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command(path, "-selection", "clipboard", "-o")
		out, err := cmd.Output()
		return string(out), err
	}

	// Try xsel
	if path, err := exec.LookPath("xsel"); err == nil {
		cmd := exec.Command(path, "--clipboard", "--output")
		out, err := cmd.Output()
		return string(out), err
	}

	// Try wl-paste (Wayland)
	if path, err := exec.LookPath("wl-paste"); err == nil {
		cmd := exec.Command(path, "--no-newline")
		out, err := cmd.Output()
		return string(out), err
	}

	return "", exec.ErrNotFound
}

// HasNerdFont checks if a Nerd Font is installed.
// Uses fc-list on Linux, system_profiler on macOS.
func HasNerdFont() bool {
	if IsDarwin() {
		return hasNerdFontDarwin()
	}
	return hasNerdFontLinux()
}

// hasNerdFontLinux checks for Nerd Fonts using fc-list.
func hasNerdFontLinux() bool {
	cmd := exec.Command("fc-list", ":family")
	out, err := cmd.Output()
	if err != nil {
		return true // Assume yes if we can't check
	}
	return matchesNerdFont(string(out))
}

// hasNerdFontDarwin checks for Nerd Fonts on macOS using system_profiler.
func hasNerdFontDarwin() bool {
	cmd := exec.Command("system_profiler", "SPFontsDataType")
	out, err := cmd.Output()
	if err != nil {
		return true // Assume yes if we can't check
	}
	return matchesNerdFont(string(out))
}

// matchesNerdFont returns true if the given text contains known Nerd Font family names.
func matchesNerdFont(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "nerd") ||
		strings.Contains(lower, "firacode") ||
		strings.Contains(lower, "jetbrains") ||
		strings.Contains(lower, "hack") ||
		strings.Contains(lower, "cascadia")
}

// NerdFontInstallHint returns platform-specific Nerd Font installation instructions.
func NerdFontInstallHint() string {
	if IsDarwin() {
		return "Install via Homebrew:\n" +
			"  brew install --cask font-fira-code-nerd-font\n\n" +
			"Then set it as your terminal font\n" +
			"(Terminal > Settings > Profiles > Font)."
	}
	return "Example (Linux):\n" +
		"  mkdir -p ~/.local/share/fonts\n" +
		"  cd ~/.local/share/fonts\n" +
		"  curl -fLO https://github.com/ryanoasis/\n" +
		"    nerd-fonts/releases/latest/download/FiraCode.zip\n" +
		"  unzip FiraCode.zip -d FiraCode && rm FiraCode.zip\n" +
		"  fc-cache -fv"
}
