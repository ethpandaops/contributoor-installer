package display

import (
	"fmt"
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
	Page        *page
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

	return modal
}

func (m *TextBoxModalLayout) setupForm(labels []string, maxLengths []int, regexes []string) int {
	m.ControlGrid = tview.NewGrid().
		SetRows(0).
		SetColumns(-1, -5, -1)
	m.ControlGrid.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	form := NewForm()
	form.SetButtonsAlign(tview.AlignCenter)
	form.SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	form.SetButtonTextColor(tview.Styles.PrimaryTextColor)
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	form.SetBorderPadding(0, 0, 0, 0)
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
		SetButtonTextColor(tcell.ColorLightGray).
		SetButtonBackgroundActivatedColor(tcell.Color46).
		SetButtonTextActivatedColor(tcell.ColorBlack)

	leftSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	rightSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	m.ControlGrid.
		AddItem(leftSpacer, 0, 0, 1, 1, 0, 0, false).
		AddItem(m.Form, 0, 1, 1, 1, 0, 0, true).
		AddItem(rightSpacer, 0, 2, 1, 1, 0, 0, false)

	return len(labels)*2 + 1
}

func (m *TextBoxModalLayout) setupNavigation() {
	navString1 := "Arrow keys: Navigate     Space/Enter: Select"
	navTextView1 := tview.NewTextView().
		SetDynamicColors(false).
		SetRegions(false).
		SetWrap(false)
	fmt.Fprint(navTextView1, navString1)

	navString2 := "Esc: Go Back     Ctrl+C: Quit without Saving"
	navTextView2 := tview.NewTextView().
		SetDynamicColors(false).
		SetRegions(false).
		SetWrap(false)
	fmt.Fprint(navTextView2, navString2)

	navBar := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(navTextView1, len(navString1), 1, false).
			AddItem(tview.NewBox(), 0, 1, false),
			1, 1, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(navTextView2, len(navString2), 1, false).
			AddItem(tview.NewBox(), 0, 1, false),
			1, 1, false)

	m.BorderGrid.AddItem(navBar, 3, 1, 1, 1, 0, 0, true)

	m.ControlGrid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape && m.Back != nil {
			m.Back()
			return nil
		}
		return event
	})
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

func (m *TextBoxModalLayout) Focus() {
	m.App.SetFocus(m.FirstBox)
	m.Form.SetFocus(0)
}

func (m *TextBoxModalLayout) GetTitle() string {
	return m.Title
}
