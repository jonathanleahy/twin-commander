package main

import (
	"fmt"
	"os"

	"github.com/rivo/tview"
)

// Viewer provides a full-screen file viewing overlay with scrollbar.
type Viewer struct {
	TextView *tview.TextView
	Wrapper  *tview.Flex // Flex containing TextView + scrollbar
}

// NewViewer creates a new file viewer.
func NewViewer() *Viewer {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)

	wrapper, _ := WrapWithScrollBar(tv)
	wrapper.SetBorder(true)
	wrapper.SetTitle(" Viewer ")
	wrapper.SetBorderPadding(0, 0, 1, 1)

	return &Viewer{
		TextView: tv,
		Wrapper:  wrapper,
	}
}

// Open loads a file into the viewer.
func (v *Viewer) Open(path string) error {
	const maxViewerBytes = 256 * 1024 // 256KB limit
	data, err := readFileHead(path, maxViewerBytes)
	if err != nil {
		return err
	}

	if isBinary(data) {
		return fmt.Errorf("binary file")
	}

	// Apply syntax highlighting
	mode := DetectHighlight(path)
	highlighted := HighlightContent(string(data), mode)

	info, _ := os.Stat(path)
	title := path
	if info != nil {
		title = fmt.Sprintf(" %s (%s) ", path, FormatSize(info.Size()))
	}

	v.Wrapper.SetTitle(title)
	v.TextView.SetText(highlighted)
	v.TextView.ScrollToBeginning()
	return nil
}
