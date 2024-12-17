package settings

import (
	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

type SettingsDisplay struct {
	app              *tview.Application
	pages            *tview.Pages
	frame            *tview.Frame
	log              *logrus.Logger
	configService    *service.ConfigService
	homePage         *page
	categoryList     *tview.List
	content          tview.Primitive
	settingsPages    []settingsPage
	networkPage      *NetworkSettingsPage
	outputServerPage *OutputServerPage
	descriptionBox   *tview.TextView
	closeButton      *tview.Button
}

func NewSettingsDisplay(log *logrus.Logger, app *tview.Application, configService *service.ConfigService) *SettingsDisplay {
	display := &SettingsDisplay{
		app:           app,
		pages:         tview.NewPages(),
		log:           log,
		configService: configService,
	}

	display.homePage = newPage(nil, "settings-home", "Categories", "", nil)

	// Create settings subpages
	display.networkPage = NewNetworkSettingsPage(display)
	display.outputServerPage = NewOutputServerPage(display)

	display.settingsPages = []settingsPage{
		display.networkPage,
		display.outputServerPage,
	}

	// Add subpages to display
	for _, subpage := range display.settingsPages {
		display.pages.AddPage(subpage.getPage().id, subpage.getPage().content, true, false)
	}

	display.createContent()
	display.homePage.content = display.content
	display.pages.AddPage(display.homePage.id, display.content, true, false)

	display.setupGrid()

	// Set initial page to home
	display.setPage(display.homePage)

	return display
}

func (d *SettingsDisplay) setupGrid() {
	// Create the main content area
	content := tview.NewFlex().SetDirection(tview.FlexRow)

	content.AddItem(d.pages, 0, 1, true)

	// Create the frame around the content
	frame := display.CreateWizardFrame(display.WizardFrameOptions{
		Content:  content,
		Title:    "Settings",
		HelpType: display.HelpSettings,
		OnEsc: func() {
			// If we're not on the home page, go back to it.
			if d.pages.HasPage("settings-home") {
				d.setPage(d.homePage)
			}
		},
	})

	d.frame = frame
	d.app.SetRoot(frame, true)
}

func (d *SettingsDisplay) setPage(page *page) {
	// Update the frame title to show current location
	d.frame.Clear()
	frame := display.CreateWizardFrame(display.WizardFrameOptions{
		Content:  d.pages,
		Title:    page.title,
		HelpType: display.HelpSettings,
		OnEsc: func() {
			if d.pages.HasPage("settings-home") {
				d.setPage(d.homePage)
			}
		},
	})

	d.frame = frame
	d.app.SetRoot(frame, true)
	d.pages.SwitchToPage(page.id)
}

func (d *SettingsDisplay) Run() error {
	return d.app.Run()
}

func (d *SettingsDisplay) createContent() {
	// Create category list
	categoryList := tview.NewList().
		SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			// Update description when selection changes
			if index >= 0 && index < len(d.settingsPages) {
				d.updateDescription(d.settingsPages[index].getPage().description)
			}
		})
	categoryList.SetBackgroundColor(tcell.ColorLightSlateGray)
	categoryList.SetBorderPadding(0, 0, 1, 1)
	d.categoryList = categoryList

	// Create description box
	d.descriptionBox = tview.NewTextView()
	d.descriptionBox.
		SetDynamicColors(true).
		SetWordWrap(true).
		SetTextAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorLightSlateGray)
	d.descriptionBox.SetBorder(true)
	d.descriptionBox.SetTitle("Description")
	d.descriptionBox.SetBorderPadding(0, 0, 1, 1)
	d.descriptionBox.SetBorderColor(tcell.ColorWhite)

	// Set initial description
	if len(d.settingsPages) > 0 {
		d.updateDescription(d.settingsPages[0].getPage().description)
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
		categoryList.AddItem(subpage.getPage().title, "", 0, nil)
	}
	categoryList.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		d.setPage(d.settingsPages[i].getPage())
	})

	// Create a frame around the category list
	categoryFrame := tview.NewFrame(categoryList)
	categoryFrame.SetBorder(true)
	categoryFrame.SetTitle("Select a Category")
	categoryFrame.SetBorderPadding(0, 0, 1, 1)
	categoryFrame.SetBorderColor(tcell.ColorWhite)
	categoryFrame.SetBackgroundColor(tcell.ColorLightSlateGray)

	// Create close button
	closeButton := tview.NewButton("Close Config")
	closeButton.SetBackgroundColorActivated(tcell.ColorGreen)
	closeButton.SetLabelColorActivated(tcell.ColorBlack)
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
		AddItem(closeButton, len("Close Config")+4, 0, false).
		AddItem(nil, 0, 1, false)

	// Create horizontal flex for category and description
	contentFlex := tview.NewFlex().
		AddItem(categoryFrame, 0, 2, true).
		AddItem(d.descriptionBox, 0, 1, false)
	contentFlex.SetBackgroundColor(tcell.ColorDarkSlateGray)

	// Main flex container
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(contentFlex, 0, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(buttonBar, 1, 0, false)
	flex.SetBackgroundColor(tcell.ColorDarkSlateGray)

	d.content = flex
}

// Helper function to update description text
func (d *SettingsDisplay) updateDescription(text string) {
	d.descriptionBox.SetText(text)
}
