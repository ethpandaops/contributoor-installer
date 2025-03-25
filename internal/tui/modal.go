package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CreateErrorModal creates a standardised error modal used throughout the installer and
// configuration screens.
func CreateErrorModal(app *tview.Application, msg string, onDone func()) *tview.Modal {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("⛔ %s", msg)).
		AddButtons([]string{ButtonTryAgain}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if onDone != nil {
				onDone()
			}
		}).
		SetBackgroundColor(tcell.ColorLightSlateGray).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tcell.ColorLightGray).
		SetTextColor(tview.Styles.PrimaryTextColor)

	// Border and button colors must be set using the primitive methods.
	modal.SetBorderColor(tcell.ColorWhite)
	modal.SetBackgroundColor(tcell.ColorLightSlateGray)

	modal.SetButtonStyle(tcell.StyleDefault.
		Background(tcell.ColorDefault).
		Foreground(tcell.ColorLightGray)).
		SetButtonActivatedStyle(tcell.StyleDefault.
			Background(ColorButtonActivated).
			Foreground(tcell.ColorBlack))

	return modal
}

// CreateLoadingModal creates a standardised loading modal used throughout the installer and
// configuration screens.
func CreateLoadingModal(app *tview.Application, msg string) *tview.Modal {
	modal := tview.NewModal().
		SetText(msg).
		SetBackgroundColor(tcell.ColorLightSlateGray).
		SetTextColor(tview.Styles.PrimaryTextColor)

	// Border and button colors must be set using the primitive methods.
	modal.SetBorderColor(tcell.ColorWhite)
	modal.SetBackgroundColor(tcell.ColorLightSlateGray)

	return modal
}
