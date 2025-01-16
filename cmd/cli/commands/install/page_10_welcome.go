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
		"Welcome to the contributoor configuration wizard!",
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
	intro := "We'll walk you through the basic setup of contributoor.\n\n"
	helperText := fmt.Sprintf("%s\n\nWelcome to the contributoor configuration wizard!\n\n%s\n\n", tui.Logo, intro)

	modal := tview.NewModal().
		SetText(helperText).
		AddButtons([]string{tui.ButtonNext}).
		SetBackgroundColor(tui.ColorFormBackground).
		SetButtonStyle(tcell.StyleDefault.
			Background(tcell.ColorDefault).
			Foreground(tcell.ColorLightGray)).
		SetButtonActivatedStyle(tcell.StyleDefault.
			Background(tui.ColorButtonActivated).
			Foreground(tcell.ColorBlack)).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 0 {
				p.display.setPage(p.display.networkConfigPage.GetPage())
			} else {
				p.display.app.Stop()
			}
		})

	modal.Box.SetBackgroundColor(tui.ColorFormBackground)
	modal.Box.SetBorderColor(tui.ColorBorder)

	p.content = modal
}
