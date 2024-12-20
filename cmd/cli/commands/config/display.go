package config

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

// ConfigDisplay is the main display for the config UI.
type ConfigDisplay struct {
	app                    *tview.Application
	pages                  *tview.Pages
	frame                  *tview.Frame
	log                    *logrus.Logger
	sidecarCfg             sidecar.ConfigManager
	homePage               *tui.Page
	categoryList           *tview.List
	content                tview.Primitive
	descriptionBox         *tview.TextView
	settingsPages          []tui.PageInterface
	closeButton            *tview.Button
	networkConfigPage      *NetworkConfigPage
	outputServerConfigPage *OutputServerConfigPage
	settingsPage           *ContributoorSettingsPage
	hasChanges             bool
}

// NewConfigDisplay creates a new Configtui.
func NewConfigDisplay(log *logrus.Logger, app *tview.Application, sidecarCfg sidecar.ConfigManager) *ConfigDisplay {
	display := &ConfigDisplay{
		app:        app,
		pages:      tview.NewPages(),
		log:        log,
		sidecarCfg: sidecarCfg,
	}

	display.homePage = tui.NewPage(nil, "config-home", "Categories", "", nil)

	// Create all the config sub-pages.
	display.networkConfigPage = NewNetworkConfigPage(display)
	display.outputServerConfigPage = NewOutputServerConfigPage(display)
	display.settingsPage = NewContributoorSettingsPage(display)
	display.settingsPages = []tui.PageInterface{
		display.networkConfigPage,
		display.outputServerConfigPage,
		display.settingsPage,
	}

	// Add all the sub-pages to the tui.
	for _, subpage := range display.settingsPages {
		display.pages.AddPage(subpage.GetPage().ID, subpage.GetPage().Content, true, false)
	}

	// Initialize the page layout.
	display.initPage()
	display.homePage.Content = display.content
	display.pages.AddPage(display.homePage.ID, display.content, true, false)
	display.setupGrid()

	// ... and finally, set the home page as the current page.
	display.setPage(display.homePage)

	return display
}

// Run starts the application.
func (d *ConfigDisplay) Run() error {
	return d.app.Run()
}

// setupGrid creates the main content area and adds the pages to it.
func (d *ConfigDisplay) setupGrid() {
	// Create the main content area.
	content := tview.NewFlex().SetDirection(tview.FlexRow)

	// Add the pages to the content area.
	content.AddItem(d.pages, 0, 1, true)

	// Create the frame around the content. This holds breadcrumbs, page counts, etc.
	frame := tui.CreatePageFrame(tui.PageFrameOptions{
		Content:  content,
		Title:    tui.TitleSettings,
		HelpType: tui.HelpSettings,
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

// setPage sets the current page and updates the frame.
func (d *ConfigDisplay) setPage(page *tui.Page) {
	d.app.SetRoot(d.frame, true)
	d.pages.SwitchToPage(page.ID)
}

func (d *ConfigDisplay) initPage() {
	// Create the list of config categories.
	categoryList := tview.NewList().
		SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			// Update description when selection changes
			if index >= 0 && index < len(d.settingsPages) {
				d.descriptionBox.SetText(d.settingsPages[index].GetPage().Description)
			}
		})
	categoryList.SetBackgroundColor(tui.ColorFormBackground)
	categoryList.SetBorderPadding(0, 0, 1, 1)
	d.categoryList = categoryList

	// Create the description box. This will show help text for the current page.
	d.descriptionBox = tview.NewTextView()
	d.descriptionBox.
		SetDynamicColors(true).
		SetWordWrap(true).
		SetTextAlign(tview.AlignLeft).
		SetBackgroundColor(tui.ColorFormBackground)
	d.descriptionBox.SetBorder(true)
	d.descriptionBox.SetTitle(tui.TitleDescription)
	d.descriptionBox.SetBorderPadding(0, 0, 1, 1)
	d.descriptionBox.SetBorderColor(tui.ColorBorder)

	// Set the initial description to the first page.
	if len(d.settingsPages) > 0 {
		d.descriptionBox.SetText(d.settingsPages[0].GetPage().Description)
	}

	// Define key bindings for the category list.
	categoryList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab, tcell.KeyBacktab:
			d.app.SetFocus(d.closeButton)

			return nil
		}

		return event
	})

	// Add categories to the list.
	for _, subpage := range d.settingsPages {
		categoryList.AddItem(subpage.GetPage().Title, "", 0, nil)
	}

	// Bind the category list to the page selection. This ensures that when a category
	// is selected, the corresponding page is shown.
	categoryList.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		d.setPage(d.settingsPages[i].GetPage())
	})

	// Create a frame around the category list.
	categoryFrame := tview.NewFrame(categoryList)
	categoryFrame.SetBorder(true)
	categoryFrame.SetTitle("Select a Category")
	categoryFrame.SetBorderPadding(0, 0, 1, 1)
	categoryFrame.SetBorderColor(tui.ColorBorder)
	categoryFrame.SetBackgroundColor(tui.ColorFormBackground)

	// Create the close button, providing users a way to bail.
	closeButton := tview.NewButton(tui.ButtonClose)
	closeButton.SetBackgroundColorActivated(tui.ColorButtonActivated)
	closeButton.SetLabelColorActivated(tui.ColorButtonText)
	d.closeButton = closeButton

	// Define key bindings for the close button.
	closeButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab, tcell.KeyBacktab:
			d.app.SetFocus(d.categoryList)

			return nil
		case tcell.KeyUp, tcell.KeyDown:
			d.app.SetFocus(d.categoryList)

			return event
		}

		return event
	})

	// Define the action for the close button.
	closeButton.SetSelectedFunc(func() {
		d.app.Stop()

		if d.hasChanges {
			fmt.Printf("%sConfiguration updated successfully%s\n", tui.TerminalColorGreen, tui.TerminalColorReset)
			fmt.Printf("For these changes to take effect, you must restart the service:\n")
			fmt.Printf("    contributoor restart\n")
		}
	})

	// Layout everything in a flex container.
	buttonBar := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(closeButton, len(tui.ButtonClose)+4, 0, false).
		AddItem(nil, 0, 1, false)

	// Create horizontal flex for category and description
	contentFlex := tview.NewFlex().
		AddItem(categoryFrame, 0, 2, true).
		AddItem(d.descriptionBox, 0, 1, false)
	contentFlex.SetBackgroundColor(tui.ColorBackground)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(contentFlex, 0, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(buttonBar, 1, 0, false)
	flex.SetBackgroundColor(tui.ColorBackground)

	d.content = flex
}

// markConfigChanged tracks that config has been modified.
func (d *ConfigDisplay) markConfigChanged() {
	d.hasChanges = true
}
