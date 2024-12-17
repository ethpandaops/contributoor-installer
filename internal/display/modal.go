package display

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CreateErrorModal creates a consistently styled error modal
func CreateErrorModal(app *tview.Application, msg string, onDone func()) *tview.Modal {
	errorModal := tview.NewModal().
		SetText("⛔ " + msg).
		AddButtons([]string{"Try Again"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if onDone != nil {
				onDone()
			}
		}).
		SetBackgroundColor(tcell.ColorLightSlateGray).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tcell.ColorLightGray).
		SetTextColor(tview.Styles.PrimaryTextColor)

	// Set border and button colors using the primitive methods
	errorModal.Box.SetBorderColor(tcell.ColorWhite)
	errorModal.Box.SetBackgroundColor(tcell.ColorLightSlateGray)

	// Style the button
	errorModal.SetButtonStyle(tcell.StyleDefault.
		Background(tcell.ColorDefault).
		Foreground(tcell.ColorLightGray)).
		SetButtonActivatedStyle(tcell.StyleDefault.
			Background(tcell.Color46).
			Foreground(tcell.ColorBlack))

	return errorModal
}

// CreateLoadingModal creates a consistently styled loading modal
func CreateLoadingModal(app *tview.Application, msg string) *tview.Modal {
	loadingModal := tview.NewModal().
		SetText(msg).
		SetBackgroundColor(tcell.ColorLightSlateGray).
		SetTextColor(tview.Styles.PrimaryTextColor)

	// Set border and button colors using the primitive methods
	loadingModal.Box.SetBorderColor(tcell.ColorWhite)
	loadingModal.Box.SetBackgroundColor(tcell.ColorLightSlateGray)

	return loadingModal
}
