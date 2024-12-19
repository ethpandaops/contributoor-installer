package config

import (
	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor-installer/internal/validate"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// NetworkConfigPage is a page that allows the user to configure the network settings.
type NetworkConfigPage struct {
	display     *ConfigDisplay
	page        *tui.Page
	content     tview.Primitive
	form        *tview.Form
	description *tview.TextView
}

// NewNetworkConfigPage creates a new NetworkConfigPage.
func NewNetworkConfigPage(cd *ConfigDisplay) *NetworkConfigPage {
	networkConfigPage := &NetworkConfigPage{
		display: cd,
	}

	networkConfigPage.initPage()
	networkConfigPage.page = tui.NewPage(
		cd.homePage,
		"config-network",
		"Network Settings",
		"Configure network settings including client endpoints and network selection",
		networkConfigPage.content,
	)

	return networkConfigPage
}

// GetPage returns the page.
func (p *NetworkConfigPage) GetPage() *tui.Page {
	return p.page
}

// initPage initializes the page.
func (p *NetworkConfigPage) initPage() {
	// Create a form to collect user input.
	form := tview.NewForm()
	p.form = form
	form.SetBackgroundColor(tui.ColorFormBackground)

	// Create a description box to display help text.
	p.description = tview.NewTextView()
	p.description.
		SetDynamicColors(true).
		SetWordWrap(true).
		SetTextAlign(tview.AlignLeft).
		SetBackgroundColor(tui.ColorFormBackground)
	p.description.SetBorder(true)
	p.description.SetTitle(tui.TitleDescription)
	p.description.SetBorderPadding(0, 0, 1, 1)
	p.description.SetBorderColor(tui.ColorBorder)

	// Grab the available networks and their descriptions.
	networks := make([]string, len(tui.AvailableNetworks))
	networkDescriptions := make(map[string]string)

	for i, network := range tui.AvailableNetworks {
		networks[i] = network.Value
		networkDescriptions[network.Value] = network.Description
	}

	// Add our form fields.
	// Find the index of the current network (from the sidecar config) in the list.
	currentNetwork := p.display.sidecarConfig.Get().NetworkName
	currentNetworkIndex := 0

	for i, network := range networks {
		if network == currentNetwork {
			currentNetworkIndex = i

			break
		}
	}

	form.AddDropDown("Network", networks, currentNetworkIndex, func(option string, index int) {
		p.description.SetText(networkDescriptions[option])
	})
	form.AddInputField("Beacon Node Address", p.display.sidecarConfig.Get().BeaconNodeAddress, 0, nil, nil)

	// Add a save button and ensure we validate the input.
	saveButton := tview.NewButton(tui.ButtonSaveSettings)
	saveButton.SetSelectedFunc(func() {
		beaconNodeAddress, _ := form.GetFormItem(1).(*tview.InputField)
		validateAndUpdate(p, beaconNodeAddress)
	})
	saveButton.SetBackgroundColorActivated(tui.ColorButtonActivated)
	saveButton.SetLabelColorActivated(tui.ColorButtonText)

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
	formFrame.SetBorderColor(tui.ColorBorder)
	formFrame.SetBackgroundColor(tui.ColorFormBackground)

	// Create a button container to hold the save button.
	buttonFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(saveButton, len(tui.ButtonSaveSettings)+4, 0, true).
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
	mainFlex.SetBackgroundColor(tui.ColorBackground)

	p.content = mainFlex
}

func validateAndUpdate(p *NetworkConfigPage, input *tview.InputField) {
	if err := validate.ValidateBeaconNodeAddress(input.GetText()); err != nil {
		p.openErrorModal(err)

		return
	}

	if err := p.display.sidecarConfig.Update(func(cfg *sidecar.Config) {
		cfg.BeaconNodeAddress = input.GetText()
	}); err != nil {
		p.openErrorModal(err)

		return
	}

	p.display.setPage(p.display.homePage)
}

func (p *NetworkConfigPage) openErrorModal(err error) {
	p.display.app.SetRoot(tui.CreateErrorModal(
		p.display.app,
		err.Error(),
		func() {
			p.display.app.SetRoot(p.display.frame, true)
			p.display.app.SetFocus(p.form)
		},
	), true)
}
