package display

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type WizardFrameOptions struct {
	Content tview.Primitive
	Step    int
	Total   int
	Title   string
	OnEsc   func()
}

// CreateWizardFrame creates a standardized frame for wizard steps
func CreateWizardFrame(opts WizardFrameOptions) *tview.Frame {
	frame := tview.NewFrame(opts.Content)
	frame.SetBorders(2, 2, 2, 2, 4, 4)
	frame.AddText("Contributoor Configuration", true, tview.AlignCenter, tcell.ColorYellow)
	frame.AddText(fmt.Sprintf("Navigation: Config Wizard > [%d/%d] %s", opts.Step, opts.Total, opts.Title), true, tview.AlignLeft, tcell.ColorWhite)
	frame.AddText("Arrow keys: Navigate    Space/Enter: Select", false, tview.AlignCenter, tcell.ColorWhite)
	frame.AddText("Esc: Go Back    Ctrl+C: Quit without Saving", false, tview.AlignCenter, tcell.ColorWhite)
	frame.SetBorderColor(tcell.ColorYellow)
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
