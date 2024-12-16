// Package display provides the display components for the CLI.
//
// Adapted from https://github.com/rocket-pool/smartnode.
package display

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TextBoxModalLayout struct {
	App         *tview.Application
	Title       string
	Width       int
	BorderGrid  *tview.Grid
	ContentGrid *tview.Grid
	ControlGrid *tview.Grid
	Done        func(text map[string]string)
	Back        func()
	Form        *Form
	FirstBox    *tview.InputField
	TextBoxes   map[string]*tview.InputField
}

type TextBoxModalOptions struct {
	Title      string
	Width      int
	Text       string
	Labels     []string
	MaxLengths []int
	Regexes    []string
	OnDone     func(text map[string]string)
	OnBack     func()
	OnEsc      func()
}

func NewTextBoxModal(app *tview.Application, opts TextBoxModalOptions) *TextBoxModalLayout {
	modal := &TextBoxModalLayout{
		App:       app,
		Title:     opts.Title,
		Width:     opts.Width,
		TextBoxes: make(map[string]*tview.InputField),
		Done:      opts.OnDone,
		Back:      opts.OnBack,
	}

	// Create the button grid
	height := modal.setupForm(opts.Labels, opts.MaxLengths, opts.Regexes)

	// Create the main text view
	textView := tview.NewTextView().
		SetText(opts.Text).
		SetTextAlign(tview.AlignCenter).
		SetWordWrap(true).
		SetTextColor(tview.Styles.PrimaryTextColor).
		SetDynamicColors(true)
	textView.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	textView.SetBorderPadding(0, 0, 1, 1)

	// Row spacers
	spacer1 := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	spacer2 := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	spacer3 := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	// Content grid
	modal.ContentGrid = tview.NewGrid().
		SetRows(1, 0, 1, height, 1).
		AddItem(spacer1, 0, 0, 1, 1, 0, 0, false).
		AddItem(textView, 1, 0, 1, 1, 0, 0, false).
		AddItem(spacer2, 2, 0, 1, 1, 0, 0, false).
		AddItem(modal.ControlGrid, 3, 0, 1, 1, 0, 0, true).
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
	modal.BorderGrid.SetRows(0, textViewHeight+height+3, 0, 2)
	modal.BorderGrid.AddItem(modal.ContentGrid, 1, 1, 1, 1, 0, 0, true)

	// Navigation footer
	// modal.setupNavigation()

	// Add key handler for ESC
	modal.Form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc && opts.OnEsc != nil {
			opts.OnEsc()
			return nil
		}
		return event
	})

	return modal
}

func (m *TextBoxModalLayout) setupForm(labels []string, maxLengths []int, regexes []string) int {
	m.ControlGrid = tview.NewGrid().
		SetRows(0).
		SetColumns(-1, -5, -1)
	m.ControlGrid.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	form := NewForm()
	form.SetButtonsAlign(tview.AlignCenter)
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	form.SetBorderPadding(0, 0, 0, 0)
	form.SetLabelColor(tcell.ColorLightGray)
	form.SetButtonStyle(tcell.StyleDefault.
		Background(tcell.ColorDefault).
		Foreground(tcell.ColorLightGray)).
		SetButtonActivatedStyle(tcell.StyleDefault.
			Background(tcell.Color46).
			Foreground(tcell.ColorBlack))

	m.Form = form

	for i, label := range labels {
		textbox := tview.NewInputField().SetLabel(label)
		maxLength := maxLengths[i]

		textbox.SetAcceptanceFunc(func(textToCheck string, lastChar rune) bool {
			return maxLength <= 0 || len(textToCheck) <= maxLength
		})

		m.Form.AddFormItem(textbox)
		m.TextBoxes[label] = textbox

		if m.FirstBox == nil {
			m.FirstBox = textbox
		}
	}

	m.Form.AddButton("Next", m.handleNext).
		SetButtonStyle(tcell.StyleDefault.
			Background(tcell.ColorDefault).
			Foreground(tcell.ColorLightGray)).
		SetButtonActivatedStyle(tcell.StyleDefault.
			Background(tcell.Color46).
			Foreground(tcell.ColorBlack)).
		SetButtonBackgroundActivatedColor(tcell.Color46).
		SetButtonTextColor(tcell.ColorLightGray).
		SetButtonTextActivatedColor(tcell.ColorBlack)

	leftSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	rightSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	m.ControlGrid.
		AddItem(leftSpacer, 0, 0, 1, 1, 0, 0, false).
		AddItem(m.Form, 0, 1, 1, 1, 0, 0, true).
		AddItem(rightSpacer, 0, 2, 1, 1, 0, 0, false)

	return len(labels)*2 + 1
}

func (m *TextBoxModalLayout) handleNext() {
	if m.Done == nil {
		return
	}

	text := make(map[string]string)
	for label, textbox := range m.TextBoxes {
		text[label] = strings.TrimSpace(textbox.GetText())
	}

	m.Done(text)
}

func (m *TextBoxModalLayout) handleBack() {
	if m.Back != nil {
		m.Back()
	}
}

func (m *TextBoxModalLayout) Focus() {
	m.App.SetFocus(m.FirstBox)
	m.Form.SetFocus(0)
}
