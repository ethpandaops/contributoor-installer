package display

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
	frame.AddText("Contributoor Configuration", true, tview.AlignCenter, ColorHeading)

	// Set navigation text based on context
	switch opts.HelpType {
	case HelpSettings:
		frame.AddText("Navigation: Settings > "+opts.Title, true, tview.AlignLeft, tcell.ColorWhite)
		frame.AddText("Tab: Go to the Buttons   Ctrl+C: Quit without Saving", false, tview.AlignCenter, tcell.ColorWhite)
		frame.AddText("Arrow keys: Navigate             Space/Enter: Select", false, tview.AlignCenter, tcell.ColorWhite)
	default: // HelpWizard
		frame.AddText(fmt.Sprintf("Navigation: Install Wizard > [%d/%d] %s", opts.Step, opts.Total, opts.Title), true, tview.AlignLeft, tcell.ColorWhite)
		frame.AddText("Esc: Go Back    Ctrl+C: Quit without Saving", false, tview.AlignCenter, tcell.ColorWhite)
		frame.AddText("Arrow keys: Navigate    Space/Enter: Select", false, tview.AlignCenter, tcell.ColorWhite)
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
