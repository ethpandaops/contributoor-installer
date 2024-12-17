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
	Done        func(values map[string]string, setError func(string))
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
	IsPassword []bool
	OnDone     func(values map[string]string, setError func(string))
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
	height := modal.setupForm(opts.Labels, opts.MaxLengths, opts.Regexes, opts.IsPassword)

	// Create the main text view
	textView := tview.NewTextView().
		SetText(opts.Text).
		SetTextAlign(tview.AlignCenter).
		SetWordWrap(true).
		SetTextColor(tview.Styles.PrimaryTextColor).
		SetDynamicColors(true)
	textView.SetBackgroundColor(ColorFormBackground)
	textView.SetBorderPadding(0, 0, 1, 1)

	// Row spacers
	spacer1 := tview.NewBox().SetBackgroundColor(ColorFormBackground)
	spacer2 := tview.NewBox().SetBackgroundColor(ColorFormBackground)
	spacer3 := tview.NewBox().SetBackgroundColor(ColorFormBackground)

	// Content grid
	modal.ContentGrid = tview.NewGrid().
		SetRows(1, 0, 1, height, 1).
		AddItem(spacer1, 0, 0, 1, 1, 0, 0, false).
		AddItem(textView, 1, 0, 1, 1, 0, 0, false).
		AddItem(spacer2, 2, 0, 1, 1, 0, 0, false).
		AddItem(modal.ControlGrid, 3, 0, 1, 1, 0, 0, true).
		AddItem(spacer3, 4, 0, 1, 1, 0, 0, false)

	modal.ContentGrid.
		SetBackgroundColor(ColorFormBackground).
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

func (m *TextBoxModalLayout) setupForm(labels []string, maxLengths []int, regexes []string, isPassword []bool) int {
	m.ControlGrid = tview.NewGrid().
		SetRows(0).
		SetColumns(-1, -5, -1)
	m.ControlGrid.SetBackgroundColor(ColorFormBackground)

	form := NewForm()
	form.SetButtonsAlign(tview.AlignCenter)
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetBackgroundColor(ColorFormBackground)
	form.SetBorderPadding(0, 0, 0, 0)
	form.SetLabelColor(tcell.ColorLightGray)
	form.SetButtonStyle(tcell.StyleDefault.
		Background(tcell.ColorDefault).
		Foreground(tcell.ColorLightGray)).
		SetButtonActivatedStyle(tcell.StyleDefault.
			Background(ColorButtonActivated).
			Foreground(tcell.ColorBlack))

	m.Form = form

	for i, label := range labels {
		var textbox *tview.InputField
		if i < len(isPassword) && isPassword[i] {
			textbox = tview.NewInputField().
				SetLabel(label).
				SetMaskCharacter('*')
		} else {
			textbox = tview.NewInputField().SetLabel(label)
		}
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

	m.Form.AddButton(ButtonNext, m.handleNext).
		SetButtonStyle(tcell.StyleDefault.
			Background(tcell.ColorDefault).
			Foreground(tcell.ColorLightGray)).
		SetButtonActivatedStyle(tcell.StyleDefault.
			Background(ColorButtonActivated).
			Foreground(tcell.ColorBlack)).
		SetButtonBackgroundActivatedColor(ColorButtonActivated).
		SetButtonTextColor(tcell.ColorLightGray).
		SetButtonTextActivatedColor(tcell.ColorBlack)

	leftSpacer := tview.NewBox().SetBackgroundColor(ColorFormBackground)
	rightSpacer := tview.NewBox().SetBackgroundColor(ColorFormBackground)

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

	m.Done(text, m.ShowError)
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

func (m *TextBoxModalLayout) ShowError(msg string) {
	errorModal := CreateErrorModal(m.App, msg, func() {
		// Clear error state and restore focus
		for _, box := range m.TextBoxes {
			box.SetBorderColor(tcell.ColorWhite)
		}
		m.App.SetRoot(m.BorderGrid, true)
		m.Focus()
	})

	// Show the error modal
	m.App.SetRoot(errorModal, true)
}
