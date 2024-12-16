package display

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	DirectionalModalVertical = iota
	DirectionalModalHorizontal
)

type ChoiceModalOptions struct {
	Title        string
	Width        int
	Text         string
	Labels       []string
	Descriptions []string
	OnSelect     func(index int)
	OnBack       func()
}

type ChoiceModal struct {
	App         *tview.Application
	BorderGrid  *tview.Grid
	ContentGrid *tview.Grid
	ButtonGrid  *tview.Grid
	Forms       []*tview.Form
	Selected    int
	DescBox     *tview.TextView
}

func NewChoiceModal(app *tview.Application, opts ChoiceModalOptions) *ChoiceModal {
	modal := &ChoiceModal{
		App:   app,
		Forms: make([]*tview.Form, 0),
	}

	// Create the button grid
	buttonGrid := tview.NewGrid().SetRows(0)
	buttonGrid.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	formsFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	if len(opts.Descriptions) > 0 {
		// Add spacing row to align with description box
		spacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
		formsFlex.AddItem(spacer, 1, 1, false)
	}

	// Create forms for each button
	for i, label := range opts.Labels {
		form := tview.NewForm()
		form.SetButtonsAlign(tview.AlignCenter)
		form.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
		form.SetBorderPadding(0, 0, 0, 0)

		// Add button to form
		index := i // Capture index for closure
		form.AddButton(label, func() {
			if opts.OnSelect != nil {
				opts.OnSelect(index)
			}
		})

		// Style the button
		button := form.GetButton(0)
		button.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		button.SetLabelColor(tcell.ColorLightGray)

		// Set button styles for selected/unselected states
		form.
			SetButtonStyle(tcell.StyleDefault.
				Background(tview.Styles.PrimitiveBackgroundColor).
				Foreground(tcell.ColorLightGray)).
			SetButtonActivatedStyle(tcell.StyleDefault.
				Background(tcell.Color46). // Bright green
				Foreground(tcell.ColorBlack))

		// Set up navigation
		button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyDown, tcell.KeyTab:
				nextIndex := (index + 1) % len(opts.Labels)
				if modal.DescBox != nil {
					modal.DescBox.SetText(opts.Descriptions[nextIndex])
				}
				modal.App.SetFocus(modal.Forms[nextIndex])
				modal.Selected = nextIndex
				return nil
			case tcell.KeyUp:
				nextIndex := index - 1
				if nextIndex < 0 {
					nextIndex = len(opts.Labels) - 1
				}
				if modal.DescBox != nil {
					modal.DescBox.SetText(opts.Descriptions[nextIndex])
				}
				modal.App.SetFocus(modal.Forms[nextIndex])
				modal.Selected = nextIndex
				return nil
			case tcell.KeyEsc:
				if opts.OnBack != nil {
					opts.OnBack()
				}
				return nil
			}
			return event
		})

		modal.Forms = append(modal.Forms, form)
		formsFlex.AddItem(form, 1, 1, true)

		spacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
		formsFlex.AddItem(spacer, 0, 1, false)
	}

	// Create description box if needed
	if len(opts.Descriptions) > 0 {
		modal.DescBox = tview.NewTextView()
		modal.DescBox.
			SetDynamicColors(true).
			SetText(opts.Descriptions[0]).
			SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
			SetBorder(true).
			SetTitle("Description").
			SetBorderPadding(0, 0, 1, 1)
	}

	// Set up the grids
	leftSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	midSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	rightSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	if modal.DescBox != nil {
		buttonGrid.SetColumns(1, -3, 1, -4, 1)
		buttonGrid.AddItem(leftSpacer, 0, 0, 1, 1, 0, 0, false)
		buttonGrid.AddItem(formsFlex, 0, 1, 1, 1, 0, 0, true)
		buttonGrid.AddItem(midSpacer, 0, 2, 1, 1, 0, 0, false)
		buttonGrid.AddItem(modal.DescBox, 0, 3, 1, 1, 0, 0, false)
		buttonGrid.AddItem(rightSpacer, 0, 4, 1, 1, 0, 0, false)
	} else {
		buttonGrid.SetColumns(0, -1, 0)
		buttonGrid.AddItem(leftSpacer, 0, 0, 1, 1, 0, 0, false)
		buttonGrid.AddItem(formsFlex, 0, 1, 1, 1, 0, 0, true)
		buttonGrid.AddItem(rightSpacer, 0, 2, 1, 1, 0, 0, false)
	}

	// Create the main text view
	textView := tview.NewTextView().
		SetText(opts.Text).
		SetTextAlign(tview.AlignCenter).
		SetWordWrap(true).
		SetTextColor(tview.Styles.PrimaryTextColor)
	textView.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	textView.SetBorderPadding(0, 0, 0, 0)

	// Row spacers
	spacer1 := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	spacer2 := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	spacer3 := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	// Content grid
	modal.ContentGrid = tview.NewGrid().
		SetRows(2, 2, 1, 0, 1).
		AddItem(spacer1, 0, 0, 1, 1, 0, 0, false).
		AddItem(textView, 1, 0, 1, 1, 0, 0, false).
		AddItem(spacer2, 2, 0, 1, 1, 0, 0, false).
		AddItem(buttonGrid, 3, 0, 1, 1, 0, 0, true).
		AddItem(spacer3, 4, 0, 1, 1, 0, 0, false)

	modal.ContentGrid.
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorder(true).
		SetTitle(" " + opts.Title + " ")

	// Border grid
	modal.BorderGrid = tview.NewGrid().
		SetColumns(0, opts.Width, 0)

	// Calculate content height
	lines := tview.WordWrap(opts.Text, opts.Width-4)
	textViewHeight := len(lines) + 4
	buttonHeight := len(opts.Labels)*2 + 1

	modal.BorderGrid.SetRows(0, textViewHeight+buttonHeight+3, 0, 2)
	modal.BorderGrid.AddItem(modal.ContentGrid, 1, 1, 1, 1, 0, 0, true)

	// Set initial focus
	modal.App.SetFocus(modal.Forms[0])
	if modal.DescBox != nil {
		modal.DescBox.SetText(opts.Descriptions[0])
	}

	return modal
}
