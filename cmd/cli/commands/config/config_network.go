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

	beaconInput := tview.NewInputField().
		SetLabel("Beacon Node Address: ").
		SetText(p.display.sidecarCfg.Get().BeaconNodeAddress)
	beaconInput.SetFocusFunc(func() {
		p.description.SetText("The address of your beacon node (e.g., http://127.0.0.1:5052)")
	})
	form.AddFormItem(beaconInput)

	// Add Docker network dropdown if using Docker.
	if p.display.sidecarCfg.Get().RunMethod == config.RunMethod_RUN_METHOD_DOCKER {
		var (
			networks               = []string{"<no network selected>"}
			commonNetworks         = []string{"host", "bridge", "default"}
			customNetworks         = make([]string, 0)
			existingCommonNetworks = make([]string, 0)
		)

		// Get list of existing Docker networks.
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

		networkDropdown.SetFocusFunc(func() {
			p.description.SetText("You can optionally attach the Contributoor container to one of your existing Docker networks.")
		})

		form.AddFormItem(networkDropdown)
	}

	metricsInput := tview.NewInputField().
		SetLabel("Optional Metrics Address: ").
		SetText(p.display.sidecarCfg.Get().MetricsAddress)
	metricsInput.SetFocusFunc(func() {
		p.description.SetText("The optional address to expose contributoor metrics on (e.g., :9090). This is NOT the address of your Beacon Node prometheus metrics. If you don't know what this is - leave it empty.")
	})
	form.AddFormItem(metricsInput)

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
		case tcell.KeyTab:
			// When tabbing from save button, go back to first form item.
			p.form.SetFocus(0)
			p.display.app.SetFocus(p.form)

			return nil
		case tcell.KeyBacktab:
			// When back-tabbing from save button, go to last form item.
			p.form.SetFocus(p.form.GetFormItemCount() - 1)
			p.display.app.SetFocus(p.form)

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
		case tcell.KeyBacktab:
			// If we're on the first form item, move to save button.
			if formIndex == 0 {
				p.display.app.SetFocus(saveButton)

				return nil
			}
		}

		return event
	})

	// Set initial focus to first form item
	form.SetFocus(0)
	p.display.app.SetFocus(form)

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
	beaconInput, _ := p.form.GetFormItem(0).(*tview.InputField)

	var (
		dockerNetwork  string
		metricsAddress string
	)

	if p.display.sidecarCfg.Get().RunMethod == config.RunMethod_RUN_METHOD_DOCKER {
		dockerDropdown, ok := p.form.GetFormItem(1).(*tview.DropDown)
		if ok {
			_, dockerNetwork = dockerDropdown.GetCurrentOption()
			if dockerNetwork == "<no network selected>" {
				dockerNetwork = ""
			}
		}

		metricsInput, _ := p.form.GetFormItem(2).(*tview.InputField)
		metricsAddress = metricsInput.GetText()
	} else {
		metricsInput, _ := p.form.GetFormItem(1).(*tview.InputField)
		metricsAddress = metricsInput.GetText()
	}

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
			p.display.app.SetRoot(p.display.frame, true).EnableMouse(true)
			p.display.app.SetFocus(p.form)
		},
	), true).EnableMouse(true)
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
