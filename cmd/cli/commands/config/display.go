package config

import (
	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

type ConfigDisplay struct {
	app                    *tview.Application
	pages                  *tview.Pages
	frame                  *tview.Frame
	log                    *logrus.Logger
	configService          *service.ConfigService
	homePage               *page
	categoryList           *tview.List
	content                tview.Primitive
	settingsPages          []settingsPage
	networkPage            *NetworkConfigPage
	OutputServerConfigPage *OutputServerConfigPage
	descriptionBox         *tview.TextView
	closeButton            *tview.Button
}

func NewConfigDisplay(log *logrus.Logger, app *tview.Application, configService *service.ConfigService) *ConfigDisplay {
	display := &ConfigDisplay{
		app:           app,
		pages:         tview.NewPages(),
		log:           log,
		configService: configService,
	}

	display.homePage = newPage(nil, "config-home", "Categories", "", nil)

	// Create settings subpages
	display.networkPage = NewNetworkConfigPage(display)
	display.OutputServerConfigPage = NewOutputServerConfigPage(display)

	display.settingsPages = []settingsPage{
		display.networkPage,
		display.OutputServerConfigPage,
	}

	// Add subpages to display
	for _, subpage := range display.settingsPages {
		display.pages.AddPage(subpage.GetPage().ID, subpage.GetPage().Content, true, false)
	}

	display.initPage()
	display.homePage.Content = display.content
	display.pages.AddPage(display.homePage.ID, display.content, true, false)

	display.setupGrid()

	// Set initial page to home
	display.setPage(display.homePage)

	return display
}

func (d *ConfigDisplay) setupGrid() {
	// Create the main content area
	content := tview.NewFlex().SetDirection(tview.FlexRow)

	content.AddItem(d.pages, 0, 1, true)

	// Create the frame around the content
	frame := display.CreatePageFrame(display.PageFrameOptions{
		Content:  content,
		Title:    display.TitleSettings,
		HelpType: display.HelpSettings,
		OnEsc: func() {
			// If we're not on the home page, go back to it.
			if d.pages.HasPage("config-home") {
				d.setPage(d.homePage)
			}
		},
	})

	d.frame = frame
	d.app.SetRoot(frame, true)
}

func (d *ConfigDisplay) setPage(page *page) {
	// Update the frame title to show current location
	d.frame.Clear()
	frame := display.CreatePageFrame(display.PageFrameOptions{
		Content:  d.pages,
		Title:    page.Title,
		HelpType: display.HelpSettings,
		OnEsc: func() {
			if d.pages.HasPage("config-home") {
				d.setPage(d.homePage)
			}
		},
	})

	d.frame = frame
	d.app.SetRoot(frame, true)
	d.pages.SwitchToPage(page.ID)
}

func (d *ConfigDisplay) Run() error {
	return d.app.Run()
}

func (d *ConfigDisplay) initPage() {
	// Create category list
	categoryList := tview.NewList().
		SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			// Update description when selection changes
			if index >= 0 && index < len(d.settingsPages) {
				d.updateDescription(d.settingsPages[index].GetPage().Description)
			}
		})
	categoryList.SetBackgroundColor(display.ColorFormBackground)
	categoryList.SetBorderPadding(0, 0, 1, 1)
	d.categoryList = categoryList

	// Create description box
	d.descriptionBox = tview.NewTextView()
	d.descriptionBox.
		SetDynamicColors(true).
		SetWordWrap(true).
		SetTextAlign(tview.AlignLeft).
		SetBackgroundColor(display.ColorFormBackground)
	d.descriptionBox.SetBorder(true)
	d.descriptionBox.SetTitle(display.TitleDescription)
	d.descriptionBox.SetBorderPadding(0, 0, 1, 1)
	d.descriptionBox.SetBorderColor(display.ColorBorder)

	// Set initial description
	if len(d.settingsPages) > 0 {
		d.updateDescription(d.settingsPages[0].GetPage().Description)
	}

	// Set tab handling for the category list
	categoryList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			d.app.SetFocus(d.closeButton)
			return nil
		case tcell.KeyBacktab:
			d.app.SetFocus(d.closeButton)
			return nil
		}
		return event
	})

	// Add categories
	for _, subpage := range d.settingsPages {
		categoryList.AddItem(subpage.GetPage().Title, "", 0, nil)
	}
	categoryList.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		d.setPage(d.settingsPages[i].GetPage())
	})

	// Create a frame around the category list
	categoryFrame := tview.NewFrame(categoryList)
	categoryFrame.SetBorder(true)
	categoryFrame.SetTitle("Select a Category")
	categoryFrame.SetBorderPadding(0, 0, 1, 1)
	categoryFrame.SetBorderColor(display.ColorBorder)
	categoryFrame.SetBackgroundColor(display.ColorFormBackground)

	// Create close button
	closeButton := tview.NewButton(display.ButtonClose)
	closeButton.SetBackgroundColorActivated(display.ColorButtonActivated)
	closeButton.SetLabelColorActivated(display.ColorButtonText)
	d.closeButton = closeButton

	// Set tab handling for the close button
	closeButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			d.app.SetFocus(d.categoryList)
			return nil
		case tcell.KeyBacktab:
			d.app.SetFocus(d.categoryList)
			return nil
		case tcell.KeyUp, tcell.KeyDown:
			d.app.SetFocus(d.categoryList)
			return event
		}
		return event
	})

	closeButton.SetSelectedFunc(func() {
		d.app.Stop()
	})

	// Layout everything in a flex container
	buttonBar := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(closeButton, len(display.ButtonClose)+4, 0, false).
		AddItem(nil, 0, 1, false)

	// Create horizontal flex for category and description
	contentFlex := tview.NewFlex().
		AddItem(categoryFrame, 0, 2, true).
		AddItem(d.descriptionBox, 0, 1, false)
	contentFlex.SetBackgroundColor(display.ColorBackground)

	// Main flex container
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(contentFlex, 0, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(buttonBar, 1, 0, false)
	flex.SetBackgroundColor(display.ColorBackground)

	d.content = flex
}

// Helper function to update description text
func (d *ConfigDisplay) updateDescription(text string) {
	d.descriptionBox.SetText(text)
}

type settingsPage interface {
	GetPage() *display.Page
}
