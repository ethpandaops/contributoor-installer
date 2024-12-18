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

type NetworkConfigPage struct {
	display     *ConfigDisplay
	page        *page
	content     tview.Primitive
	form        *tview.Form
	description *tview.TextView
}

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

func (p *NetworkConfigPage) GetPage() *display.Page {
	return p.page
}

func (p *NetworkConfigPage) initPage() {
	// Create form
	form := tview.NewForm()
	p.form = form
	form.SetBackgroundColor(display.ColorFormBackground)

	// Create description box
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

	// Field descriptions
	descriptions := map[string]string{
		"Network":             "The Ethereum network your beacon node is running against.",
		"Beacon Node Address": "The URL of your beacon node's HTTP endpoint.",
	}

	// Network options from display constants
	networks := make([]string, len(display.AvailableNetworks))
	networkDescriptions := make(map[string]string)
	for i, network := range display.AvailableNetworks {
		networks[i] = network.Value
		networkDescriptions[network.Value] = network.Description
		descriptions["Network"] = network.Description // Update description when network is selected
	}

	// Find current network index
	defaultIndex := 0
	for i, network := range networks {
		if network == p.display.configService.Get().NetworkName {
			defaultIndex = i
			break
		}
	}

	// Add form fields without immediate config updates
	form.AddDropDown("Network", networks, defaultIndex, func(option string, index int) {
		// Update description when network changes
		p.updateDescription(networkDescriptions[option])
	})
	form.AddInputField("Beacon Node Address", p.display.configService.Get().BeaconNodeAddress, 0, nil, nil)

	// Create save button with validation
	saveButton := tview.NewButton(display.ButtonSaveSettings)
	saveButton.SetSelectedFunc(func() {
		validateAndUpdate(p, form.GetFormItem(1).(*tview.InputField))
	})
	saveButton.SetBackgroundColorActivated(display.ColorButtonActivated)
	saveButton.SetLabelColorActivated(display.ColorButtonText)

	// Handle save button input
	saveButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			p.display.app.SetFocus(form)
			return nil
		case tcell.KeyBacktab:
			p.display.app.SetFocus(form)
			return nil
		}
		return event
	})

	// Now set up form input capture with access to saveButton
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Get the currently focused item
		formIndex, _ := form.GetFocusedItemIndex()

		switch event.Key() {
		case tcell.KeyTab:
			// If we're on the last form item, move to save button
			if formIndex == form.GetFormItemCount()-1 {
				p.display.app.SetFocus(saveButton)
				return nil
			}
			// Otherwise, let the form handle tab navigation
			return event
		case tcell.KeyBacktab:
			// If we're on the first form item, move to save button
			if formIndex == 0 {
				p.display.app.SetFocus(saveButton)
				return nil
			}
			// Otherwise, let the form handle tab navigation
			return event
		default:
			// Update description for current field
			if item := form.GetFormItem(formIndex); item != nil {
				p.updateDescription(descriptions[item.GetLabel()])
			}
			return event
		}
	})

	// Create frame for the form
	formFrame := tview.NewFrame(form)
	formFrame.SetBorder(true)
	formFrame.SetTitle("Network Settings")
	formFrame.SetBorderPadding(0, 0, 1, 1)
	formFrame.SetBorderColor(display.ColorBorder)
	formFrame.SetBackgroundColor(display.ColorFormBackground)

	// Create button container
	buttonFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(saveButton, len(display.ButtonSaveSettings)+4, 0, true).
		AddItem(nil, 0, 1, false)

	// Create horizontal flex for form and description
	formDescriptionFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(formFrame, 0, 2, true).
		AddItem(p.description, 0, 1, false)

	// Create main layout with form+description and save button
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(formDescriptionFlex, 0, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(buttonFlex, 1, 0, false).
		AddItem(nil, 1, 0, false)
	mainFlex.SetBackgroundColor(display.ColorBackground)

	p.content = mainFlex
}

// Helper function to update description text
func (p *NetworkConfigPage) updateDescription(text string) {
	p.description.SetText(text)
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
	text := input.GetText()

	// Get current network selection before validation
	dropdown := p.form.GetFormItem(0).(*tview.DropDown)
	index, networkOption := dropdown.GetCurrentOption()

	// Show loading modal while validating
	loadingModal := display.CreateLoadingModal(
		p.display.app,
		"\n[yellow]Validating configuration\nPlease wait...[white]",
	)
	p.display.app.SetRoot(loadingModal, true)

	// Validate in goroutine to not block UI
	go func() {
		err := validateBeaconNode(text)

		p.display.app.QueueUpdateDraw(func() {
			if err != nil {
				// Show error modal
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

			// Update config only if validation passes
			p.display.configService.Update(func(cfg *service.ContributoorConfig) {
				if index != -1 {
					cfg.NetworkName = networkOption
				}
				cfg.BeaconNodeAddress = text
			})

			// Return to settings home
			p.display.setPage(p.display.homePage)
		})
	}()
}
