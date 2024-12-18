package config

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// NetworkConfigPage is a page that allows the user to configure the network settings.
type NetworkConfigPage struct {
	display     *ConfigDisplay
	page        *page
	content     tview.Primitive
	form        *tview.Form
	description *tview.TextView
}

// NewNetworkConfigPage creates a new NetworkConfigPage.
func NewNetworkConfigPage(display *ConfigDisplay) *NetworkConfigPage {
	networkPage := &NetworkConfigPage{
		display: display,
	}

	networkPage.initPage()
	networkPage.page = newPage(
		display.homePage,
		"config-network",
		"Network Settings",
		"Configure network settings including client endpoints and network selection",
		networkPage.content,
	)

	return networkPage
}

// GetPage returns the page.
func (p *NetworkConfigPage) GetPage() *display.Page {
	return p.page
}

// initPage initializes the page.
func (p *NetworkConfigPage) initPage() {
	// Create a form to collect user input.
	form := tview.NewForm()
	p.form = form
	form.SetBackgroundColor(display.ColorFormBackground)

	// Create a description box to display help text.
	p.description = tview.NewTextView()
	p.description.
		SetDynamicColors(true).
		SetWordWrap(true).
		SetTextAlign(tview.AlignLeft).
		SetBackgroundColor(display.ColorFormBackground)
	p.description.SetBorder(true)
	p.description.SetTitle(display.TitleDescription)
	p.description.SetBorderPadding(0, 0, 1, 1)
	p.description.SetBorderColor(display.ColorBorder)

	// Grab the available networks and their descriptions.
	networks := make([]string, len(display.AvailableNetworks))
	networkDescriptions := make(map[string]string)
	for i, network := range display.AvailableNetworks {
		networks[i] = network.Value
		networkDescriptions[network.Value] = network.Description
	}

	// Add our form fields.
	form.AddDropDown("Network", networks, 0, func(option string, index int) {
		p.description.SetText(networkDescriptions[option])
	})
	form.AddInputField("Beacon Node Address", p.display.configService.Get().BeaconNodeAddress, 0, nil, nil)

	// Add a save button and ensure we validate the input.
	saveButton := tview.NewButton(display.ButtonSaveSettings)
	saveButton.SetSelectedFunc(func() {
		validateAndUpdate(p, form.GetFormItem(1).(*tview.InputField))
	})
	saveButton.SetBackgroundColorActivated(display.ColorButtonActivated)
	saveButton.SetLabelColorActivated(display.ColorButtonText)

	// Define key bindings for the save button.
	saveButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab, tcell.KeyBacktab:
			p.display.app.SetFocus(form)
			return nil
		}
		return event
	})

	// Define key bindings for the form.
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Get the currently focused item.
		formIndex, _ := form.GetFocusedItemIndex()

		switch event.Key() {
		case tcell.KeyTab:
			// If we're on the last form item, move to save button.
			if formIndex == form.GetFormItemCount()-1 {
				p.display.app.SetFocus(saveButton)
				return nil
			}

			return event
		case tcell.KeyBacktab:
			// If we're on the first form item, move to save button.
			if formIndex == 0 {
				p.display.app.SetFocus(saveButton)
				return nil
			}

			return event
		default:
			return event
		}
	})

	// We wrap the form in a frame to add a border and title.
	formFrame := tview.NewFrame(form)
	formFrame.SetBorder(true)
	formFrame.SetTitle("Network Settings")
	formFrame.SetBorderPadding(0, 0, 1, 1)
	formFrame.SetBorderColor(display.ColorBorder)
	formFrame.SetBackgroundColor(display.ColorFormBackground)

	// Create a button container to hold the save button.
	buttonFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(saveButton, len(display.ButtonSaveSettings)+4, 0, true).
		AddItem(nil, 0, 1, false)

	// Create a horizontal flex to hold the form and description.
	formDescriptionFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(formFrame, 0, 2, true).
		AddItem(p.description, 0, 1, false)

	// Create a main layout both the flexes.
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(formDescriptionFlex, 0, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(buttonFlex, 1, 0, false).
		AddItem(nil, 1, 0, false)
	mainFlex.SetBackgroundColor(display.ColorBackground)

	p.content = mainFlex
}

func validateBeaconNode(address string) error {
	// Check if URL is valid
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		return fmt.Errorf("Beacon node address must start with http:// or https://")
	}

	// Try to connect to the beacon node
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(fmt.Sprintf("%s/eth/v1/node/health", address))
	if err != nil {
		return fmt.Errorf("We're unable to connect to your beacon node: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Beacon node returned status %d", resp.StatusCode)
	}

	return nil
}

func validateAndUpdate(p *NetworkConfigPage, input *tview.InputField) {
	var (
		text                 = input.GetText()
		dropdown             = p.form.GetFormItem(0).(*tview.DropDown)
		index, networkOption = dropdown.GetCurrentOption()
	)

	// Show loading modal while validating, we reach out to the beacon node
	// to validate the address, which can lock-up the UI while it does it.
	// Better to show a loading modal than the user seeing a blank screen.
	loadingModal := display.CreateLoadingModal(
		p.display.app,
		"\n[yellow]Validating configuration\nPlease wait...[white]",
	)
	p.display.app.SetRoot(loadingModal, true)

	go func() {
		err := validateBeaconNode(text)

		p.display.app.QueueUpdateDraw(func() {
			if err != nil {
				errorModal := display.CreateErrorModal(
					p.display.app,
					err.Error(),
					func() {
						p.display.app.SetRoot(p.display.frame, true)
						p.display.app.SetFocus(p.form)
					},
				)
				p.display.app.SetRoot(errorModal, true)
				return
			}

			p.display.configService.Update(func(cfg *service.ContributoorConfig) {
				if index != -1 {
					cfg.NetworkName = networkOption
				}
				cfg.BeaconNodeAddress = text
			})

			p.display.setPage(p.display.homePage)
		})
	}()
}
