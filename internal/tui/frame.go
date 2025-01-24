package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// HelpType is the type of help to display. This allows us to tweak the help text
// for the installer wizard and config pages.
type HelpType int

const (
	HelpWizard HelpType = iota
	HelpSettings
)

// PageFrameOptions is the options for the PageFrame.
type PageFrameOptions struct {
	Content  tview.Primitive
	Step     int
	Total    int
	Title    string
	OnEsc    func()
	HelpType HelpType
}

// CreatePageFrame creates a standardised frame for the installer wizard or contributoor config pages.
func CreatePageFrame(opts PageFrameOptions) *tview.Frame {
	frame := tview.NewFrame(opts.Content)
	frame.SetBorders(2, 2, 2, 2, 4, 4)

	// Set navigation text based on context
	switch opts.HelpType {
	case HelpSettings:
		frame.AddText("Contributoor Configuration", true, tview.AlignCenter, ColorHeading)
		frame.AddText("Tab: Go to the Buttons", false, tview.AlignLeft, tcell.ColorWhite)
		frame.AddText("Ctrl+C: Quit without Saving", false, tview.AlignLeft, tcell.ColorWhite)
		frame.AddText("Arrow keys: Navigate", false, tview.AlignRight, tcell.ColorWhite)
		frame.AddText("Space/Enter: Select", false, tview.AlignRight, tcell.ColorWhite)
	default: // HelpWizard
		frame.AddText("Contributoor Installation", true, tview.AlignRight, ColorHeading)
		frame.AddText(fmt.Sprintf("Navigation: Install Wizard > [%d/%d] %s", opts.Step, opts.Total, opts.Title), true, tview.AlignLeft, tcell.ColorWhite)

		frame.AddText("Esc: Go Back", false, tview.AlignLeft, tcell.ColorWhite)
		frame.AddText("Ctrl+C: Quit without Saving", false, tview.AlignLeft, tcell.ColorWhite)
		frame.AddText("Arrow keys: Navigate", false, tview.AlignRight, tcell.ColorWhite)
		frame.AddText("Space/Enter: Select", false, tview.AlignRight, tcell.ColorWhite)
	}

	frame.SetBorderColor(ColorHeading)
	frame.SetBorder(true)

	if opts.OnEsc != nil {
		frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				opts.OnEsc()

				return nil
			}

			return event
		})
	}

	return frame
}
