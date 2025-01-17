package config

import (
	"os/exec"
	"sort"
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor-installer/internal/validate"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
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
		networks[i] = network.Value.String()
		networkDescriptions[network.Value.String()] = network.Description
	}

	// Add our form fields.
	// Find the index of the current network (from the sidecar config) in the list.
	currentNetwork := p.display.sidecarCfg.Get().NetworkName
	currentNetworkIndex := 0

	for i, network := range networks {
		if network == currentNetwork.String() {
			currentNetworkIndex = i

			break
		}
	}

	form.AddDropDown("Network", networks, currentNetworkIndex, func(option string, index int) {
		p.description.SetText(networkDescriptions[option])
	})
	form.AddInputField("Beacon Node Address", p.display.sidecarCfg.Get().BeaconNodeAddress, 0, nil, nil)

	// Add Docker network dropdown if using Docker.
	if p.display.sidecarCfg.Get().RunMethod == config.RunMethod_RUN_METHOD_DOCKER {
		// Get list of existing Docker networks.
		networks := []string{"<no network selected>"}
		commonNetworks := []string{"host", "bridge", "default"}
		customNetworks := []string{}
		existingCommonNetworks := []string{}

		cmd := exec.Command("docker", "network", "ls", "--format", "{{.Name}}")

		output, err := cmd.Output()
		if err == nil {
			for _, network := range strings.Split(strings.TrimSpace(string(output)), "\n") {
				if network != "" && !strings.Contains(network, "contributoor") && network != "none" {
					if contains(commonNetworks, network) {
						existingCommonNetworks = append(existingCommonNetworks, network)
					} else {
						customNetworks = append(customNetworks, network)
					}
				}
			}
		}

		// Add custom networks first and then common networks.
		sort.Strings(customNetworks)
		sort.Strings(existingCommonNetworks)

		networks = append(networks, customNetworks...)
		networks = append(networks, existingCommonNetworks...)

		// Add network dropdown.
		networkDropdown := tview.NewDropDown().
			SetLabel("Optional Docker Network: ").
			SetOptions(networks, nil).
			SetFieldBackgroundColor(tcell.ColorBlack).
			SetLabelColor(tcell.ColorLightGray).
			SetFieldTextColor(tcell.ColorLightGray)

		// Set current value if exists.
		currentNetwork := p.display.sidecarCfg.Get().DockerNetwork
		if currentNetwork == "" {
			networkDropdown.SetCurrentOption(0)
		} else {
			for i, network := range networks {
				if network == currentNetwork {
					networkDropdown.SetCurrentOption(i)

					break
				}
			}
		}

		form.AddFormItem(networkDropdown)
	}

	form.AddInputField("Optional Metrics Address", p.display.sidecarCfg.Get().MetricsAddress, 0, nil, nil)

	// Add a save button and ensure we validate the input.
	saveButton := tview.NewButton(tui.ButtonSaveSettings)
	saveButton.SetSelectedFunc(func() {
		validateAndUpdateNetwork(p)
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

func validateAndUpdateNetwork(p *NetworkConfigPage) {
	networkDropdown, _ := p.form.GetFormItem(0).(*tview.DropDown)
	beaconInput, _ := p.form.GetFormItem(1).(*tview.InputField)

	var (
		dockerNetwork  string
		metricsAddress string
	)

	if p.display.sidecarCfg.Get().RunMethod == config.RunMethod_RUN_METHOD_DOCKER {
		dockerDropdown, ok := p.form.GetFormItem(2).(*tview.DropDown)
		if ok {
			_, dockerNetwork = dockerDropdown.GetCurrentOption()
			if dockerNetwork == "<no network selected>" {
				dockerNetwork = ""
			}
		}

		metricsInput, _ := p.form.GetFormItem(3).(*tview.InputField)
		metricsAddress = metricsInput.GetText()
	} else {
		metricsInput, _ := p.form.GetFormItem(2).(*tview.InputField)
		metricsAddress = metricsInput.GetText()
	}

	_, networkName := networkDropdown.GetCurrentOption()
	beaconAddress := beaconInput.GetText()

	if err := validate.ValidateBeaconNodeAddress(beaconAddress); err != nil {
		p.openErrorModal(err)

		return
	}

	if err := validate.ValidateMetricsAddress(metricsAddress); err != nil {
		p.openErrorModal(err)

		return
	}

	if err := p.display.sidecarCfg.Update(func(cfg *config.Config) {
		cfg.NetworkName = config.NetworkName(config.NetworkName_value[networkName])
		cfg.BeaconNodeAddress = beaconAddress
		cfg.MetricsAddress = metricsAddress
		cfg.DockerNetwork = dockerNetwork
	}); err != nil {
		p.openErrorModal(err)

		return
	}

	p.display.markConfigChanged()
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

// contains checks if a string is present in a slice.
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}

	return false
}
