package main

import "github.com/rivo/tview"

// WrapWithScrollBar wraps a tview.TextView in a Flex container that provides
// a visual scrollbar indicator. Returns (wrapper, scrollbar).
// The scrollbar is a narrow text view on the right side that shows position.
func WrapWithScrollBar(tv *tview.TextView) (*tview.Flex, *tview.TextView) {
	scrollbar := tview.NewTextView()
	scrollbar.SetDynamicColors(true)

	wrapper := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tv, 0, 1, true).
		AddItem(scrollbar, 1, 0, false)

	return wrapper, scrollbar
}
