package install

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type WelcomePage struct {
	display     *InstallDisplay
	page        *page
	content     tview.Primitive
	description *tview.TextView
}

func NewWelcomePage(display *InstallDisplay) *WelcomePage {
	welcomePage := &WelcomePage{
		display: display,
	}

	welcomePage.initPage()
	welcomePage.page = newPage(
		nil,
		"install-welcome",
		"Welcome",
		"Welcome to the contributoor configuration wizard!",
		welcomePage.content,
	)

	return welcomePage
}

func (p *WelcomePage) GetPage() *display.Page {
	return p.page
}

func (p *WelcomePage) initPage() {
	intro := "We'll walk you through the basic setup of contributoor.\n\n"
	helperText := fmt.Sprintf("%s\n\nWelcome to the contributoor configuration wizard!\n\n%s\n\n", display.Logo, intro)

	modal := tview.NewModal().
		SetText(helperText).
		AddButtons([]string{display.ButtonNext, display.ButtonClose}).
		SetBackgroundColor(display.ColorFormBackground).
		SetButtonStyle(tcell.StyleDefault.
			Background(tcell.ColorDefault).
			Foreground(tcell.ColorLightGray)).
		SetButtonActivatedStyle(tcell.StyleDefault.
			Background(display.ColorButtonActivated).
			Foreground(tcell.ColorBlack)).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 0 {
				p.display.setPage(p.display.networkPage.GetPage())
			} else {
				p.display.app.Stop()
			}
		})

	modal.Box.SetBackgroundColor(display.ColorFormBackground)
	modal.Box.SetBorderColor(display.ColorBorder)

	p.content = modal
}
