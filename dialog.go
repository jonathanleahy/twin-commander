package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ShowConfirmDialog displays a yes/no confirmation dialog.
func ShowConfirmDialog(pages *tview.Pages, title, message string, callback func(confirmed bool)) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("confirm-dialog")
			callback(buttonLabel == "Yes")
		})
	modal.SetTitle(title)

	pages.AddPage("confirm-dialog", modal, true, true)
}

// ShowErrorDialog displays an error message dialog.
func ShowErrorDialog(pages *tview.Pages, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("error-dialog")
		})
	modal.SetTitle("Error")
	modal.SetBorderColor(tcell.ColorRed)

	pages.AddPage("error-dialog", modal, true, true)
}

// ShowChoiceDialog displays a dialog with multiple choices.
func ShowChoiceDialog(pages *tview.Pages, title, message string, choices []string, callback func(label string)) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons(choices).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("choice-dialog")
			callback(buttonLabel)
		})
	modal.SetTitle(title)

	pages.AddPage("choice-dialog", modal, true, true)
}

// ShowInputDialog displays a dialog with a text input field.
func ShowInputDialog(pages *tview.Pages, app *tview.Application, title, label, defaultValue string, callback func(value string, cancelled bool)) {
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle(" " + title + " ")
	form.SetBorderPadding(1, 1, 2, 2)
	form.SetButtonsAlign(tview.AlignRight)

	form.AddInputField(label, defaultValue, 40, nil, nil)
	form.AddButton("OK", func() {
		value := form.GetFormItemByLabel(label).(*tview.InputField).GetText()
		pages.RemovePage("input-dialog")
		callback(value, false)
	})
	form.AddButton("Cancel", func() {
		pages.RemovePage("input-dialog")
		callback("", true)
	})
	form.SetCancelFunc(func() {
		pages.RemovePage("input-dialog")
		callback("", true)
	})

	// Center the dialog
	width := 50
	height := 9
	overlay := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, height, 0, true).
			AddItem(nil, 0, 1, false), width, 0, true).
		AddItem(nil, 0, 1, false)

	pages.AddPage("input-dialog", overlay, true, true)
	app.SetFocus(form)
}
