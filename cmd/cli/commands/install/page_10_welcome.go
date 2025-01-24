package install

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// WelcomePage is the first/introductory page of the install wizard.
type WelcomePage struct {
	display *InstallDisplay
	page    *tui.Page
	content tview.Primitive
}

// NewWelcomePage creates a new WelcomePage.
func NewWelcomePage(display *InstallDisplay) *WelcomePage {
	welcomePage := &WelcomePage{
		display: display,
	}

	welcomePage.initPage()
	welcomePage.page = tui.NewPage(
		nil,
		"install-welcome",
		"Welcome",
		"Welcome to the contributoor configuration wizard",
		welcomePage.content,
	)

	return welcomePage
}

// GetPage returns the page.
func (p *WelcomePage) GetPage() *tui.Page {
	return p.page
}

// initPage initializes the page.
func (p *WelcomePage) initPage() {
	var (
		intro          = "We'll walk you through the basic setup of contributoor.\n\n"
		helperText     = fmt.Sprintf("%s\n\nWelcome to the contributoor configuration wizard!\n\n%s", tui.Logo, intro)
		modalWidth     = 70
		lines          = tview.WordWrap("Select which network you're using", modalWidth-4)
		textViewHeight = len(lines) + 4
	)

	// Create the main text view.
	textView := tview.NewTextView()
	textView.SetText(helperText)
	textView.SetTextAlign(tview.AlignCenter)
	textView.SetWordWrap(true)
	textView.SetTextColor(tview.Styles.PrimaryTextColor)
	textView.SetBackgroundColor(tui.ColorFormBackground)
	textView.SetBorderPadding(0, 0, 0, 0)

	// Create the button form.
	form := tview.NewForm()
	form.AddButton(tui.ButtonNext, func() {
		p.display.setPage(p.display.networkConfigPage.GetPage())
	})
	form.SetButtonsAlign(tview.AlignCenter)
	form.SetBackgroundColor(tui.ColorFormBackground)
	form.SetBorderPadding(0, 0, 0, 0)
	form.SetButtonStyle(tcell.StyleDefault.
		Background(tcell.ColorDefault).
		Foreground(tcell.ColorLightGray))
	form.SetButtonActivatedStyle(tcell.StyleDefault.
		Background(tui.ColorButtonActivated).
		Foreground(tcell.ColorBlack))

	// Set initial focus on the Next button.
	p.display.app.SetFocus(form.GetButton(0))

	// Supporting control between mouse + keyboard.
	// Add input capture to the button for keyboard navigation. This is
	// needed because we "EnableMouse(true)" on the app. If users click
	// away from the button, it will not be focused, so we need to capture
	// the tab key events and ensure it can set focus to the button.
	button := form.GetButton(0)
	button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			p.display.app.SetFocus(button)

			return nil
		case tcell.KeyEsc:
			if p.page.Parent != nil {
				p.display.setPage(p.page.Parent)
			}

			return nil
		}

		return event
	})

	// Add input capture to the text view to allow tabbing to button.
	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			p.display.app.SetFocus(button)

			return nil
		}

		return event
	})

	// Create a flex container to stack text and button vertically.
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 1, 0, false). // Spacer
		AddItem(form, 3, 0, true)                                                         // Make form focusable by default

	// Create content grid with more vertical space.
	contentGrid := tview.NewGrid()
	contentGrid.SetRows(1, 0, 1)
	contentGrid.SetBackgroundColor(tui.ColorFormBackground)
	contentGrid.SetBorder(true)
	contentGrid.SetTitle(" 👋 Welcome ")

	// Add items to content grid.
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 0, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(flex, 1, 0, 1, 1, 0, 0, true)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 2, 0, 1, 1, 0, 0, false)

	borderGrid := tview.NewGrid().SetColumns(0, modalWidth, 0)
	borderGrid.SetRows(0, textViewHeight+20, 0, 2)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	p.content = borderGrid
}
