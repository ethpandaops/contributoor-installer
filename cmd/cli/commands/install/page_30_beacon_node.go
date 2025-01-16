package install

import (
	"os/exec"
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor-installer/internal/validate"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// BeaconNodePage is the page for configuring the users beacon node.
type BeaconNodePage struct {
	display *InstallDisplay
	page    *tui.Page
	content tview.Primitive
	form    *tview.Form
}

// NewBeaconNodePage creates a new BeaconNodePage.
func NewBeaconNodePage(display *InstallDisplay) *BeaconNodePage {
	beaconPage := &BeaconNodePage{
		display: display,
	}

	beaconPage.initPage()
	beaconPage.page = tui.NewPage(
		display.networkConfigPage.GetPage(),
		"install-beacon",
		"Beacon Node",
		"Configure your beacon node connection",
		beaconPage.content,
	)

	return beaconPage
}

// GetPage returns the page.
func (p *BeaconNodePage) GetPage() *tui.Page {
	return p.page
}

// initPage initializes the page.
func (p *BeaconNodePage) initPage() {
	var (
		// Some basic dimensions for the page modal.
		modalWidth     = 70
		lines          = tview.WordWrap("Please enter the address of your Beacon Node.\nFor example: http://127.0.0.1:5052", modalWidth-4)
		textViewHeight = len(lines) + 8
		formHeight     = 6 // Input field + network dropdown + button + padding

		// Main grids.
		contentGrid = tview.NewGrid()
		borderGrid  = tview.NewGrid().SetColumns(0, modalWidth, 0)

		// Form components.
		form = tview.NewForm()
	)

	// We need a form to house our input field.
	form.SetButtonsAlign(tview.AlignCenter)
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetBackgroundColor(tui.ColorFormBackground)
	form.SetBorderPadding(0, 0, 0, 0) // Reset padding
	form.SetLabelColor(tcell.ColorLightGray)

	// Add input field to our form to capture the users beacon node address.
	inputField := tview.NewInputField().
		SetLabel("Beacon Node Address: ").
		SetText(p.display.sidecarCfg.Get().BeaconNodeAddress).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetLabelColor(tcell.ColorLightGray)
	form.AddFormItem(inputField)

	// Only show docker network options if runMethod is Docker.
	if p.display.sidecarCfg.Get().RunMethod == config.RunMethod_RUN_METHOD_DOCKER {
		// Get list of existing Docker networks the user has.
		networks := []string{"<Please Select>"}
		cmd := exec.Command("docker", "network", "ls", "--format", "{{.Name}}")

		output, err := cmd.Output()
		if err == nil {
			for _, network := range strings.Split(strings.TrimSpace(string(output)), "\n") {
				if network != "" && !strings.Contains(network, "contributoor") {
					networks = append(networks, network)
				}
			}
		}

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
			networkDropdown.SetCurrentOption(0) // Select placeholder by default
		} else {
			for i, network := range networks {
				if network == currentNetwork {
					networkDropdown.SetCurrentOption(i)

					break
				}
			}
		}

		form.AddFormItem(networkDropdown)

		// Now that we have the button, set up the dropdown callback
		if button := form.GetButton(0); button != nil {
			if item := p.form.GetFormItem(1); item != nil {
				if dropdown, ok := item.(*tview.DropDown); ok && dropdown != nil {
					dropdown.SetSelectedFunc(func(text string, index int) {
						p.display.app.SetFocus(button)
					})
				}
			}
		}
	}

	// Add our form to the page for easy access during validation.
	p.form = form

	// Wrap our form in a frame to add a border.
	formFrame := tview.NewFrame(form)
	formFrame.SetBorderPadding(0, 0, 0, 0) // Reset padding
	formFrame.SetBackgroundColor(tui.ColorFormBackground)

	// Add 'Next' button to our form.
	form.AddButton(tui.ButtonNext, func() {
		validateAndUpdate(p, inputField)
	})

	if button := form.GetButton(0); button != nil {
		button.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		button.SetLabelColor(tcell.ColorLightGray)
		form.SetButtonStyle(tcell.StyleDefault.
			Background(tview.Styles.PrimitiveBackgroundColor).
			Foreground(tcell.ColorLightGray))
		form.SetButtonActivatedStyle(tcell.StyleDefault.
			Background(tui.ColorButtonActivated).
			Foreground(tcell.ColorBlack))

		// Set up dropdown callback now that we have the button
		if p.display.sidecarCfg.Get().RunMethod == config.RunMethod_RUN_METHOD_DOCKER {
			if item := p.form.GetFormItem(1); item != nil {
				if dropdown, ok := item.(*tview.DropDown); ok && dropdown != nil {
					dropdown.SetSelectedFunc(func(text string, index int) {
						p.display.app.SetFocus(button)
					})
				}
			}
		}
	}

	// Create the header text view
	headerView := tview.NewTextView()
	headerView.SetText("Please enter the address of your Beacon Node.")
	headerView.SetTextAlign(tview.AlignCenter)
	headerView.SetTextColor(tview.Styles.PrimaryTextColor)
	headerView.SetBackgroundColor(tui.ColorFormBackground)
	headerView.SetBorderPadding(0, 0, 0, 0)

	// Create the examples text view.
	optionsView := tview.NewTextView()

	var optionsText string
	if p.display.sidecarCfg.Get().RunMethod == config.RunMethod_RUN_METHOD_DOCKER {
		optionsText = "\nExamples:\n1. Local beacon node (e.g., http://127.0.0.1:5052)\n2. Docker container (e.g., http://beacon:5052)\n   - Optionally specify an existing Docker network to join"
	} else {
		optionsText = "\nExample: http://127.0.0.1:5052"
	}

	optionsView.SetText(optionsText)
	optionsView.SetTextAlign(tview.AlignLeft)
	optionsView.SetWordWrap(true)
	optionsView.SetTextColor(tview.Styles.PrimaryTextColor)
	optionsView.SetBackgroundColor(tui.ColorFormBackground)
	optionsView.SetBorderPadding(0, 0, 0, 0)

	// Set up the content grid.
	contentGrid.SetRows(1, 1, 5, 1, 6, 1)
	contentGrid.SetBackgroundColor(tui.ColorFormBackground)
	contentGrid.SetBorder(true)
	contentGrid.SetTitle(" Beacon Node ")

	// Add items to content grid using spacers.
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 0, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(headerView, 1, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(optionsView, 2, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 3, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(formFrame, 4, 0, 2, 1, 0, 0, true)

	// Border grid.
	borderGrid.SetRows(0, textViewHeight+formHeight+1, 0, 2)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	// Set initial focus.
	p.display.app.SetFocus(form)
	p.content = borderGrid
}

func validateAndUpdate(p *BeaconNodePage, input *tview.InputField) {
	var networkName string

	if p.display.sidecarCfg.Get().RunMethod == config.RunMethod_RUN_METHOD_DOCKER {
		if item := p.form.GetFormItem(1); item != nil {
			if dropdown, ok := item.(*tview.DropDown); ok {
				_, networkName = dropdown.GetCurrentOption()
				if networkName == "<Please Select>" {
					networkName = ""
				}
			}
		}
	}

	if err := validate.ValidateBeaconNodeAddress(input.GetText()); err != nil {
		p.openErrorModal(err)

		return
	}

	if err := p.display.sidecarCfg.Update(func(cfg *config.Config) {
		cfg.BeaconNodeAddress = input.GetText()
		cfg.DockerNetwork = networkName
	}); err != nil {
		p.openErrorModal(err)

		return
	}

	p.display.setPage(p.display.outputServerCredentialsPage.GetPage())
}

func (p *BeaconNodePage) openErrorModal(err error) {
	p.display.app.SetRoot(tui.CreateErrorModal(
		p.display.app,
		err.Error(),
		func() {
			p.display.app.SetRoot(p.display.frame, true)
		},
	), true)
}
