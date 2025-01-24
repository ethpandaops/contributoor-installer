package install

import (
	"fmt"
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor-installer/internal/validate"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// OutputServerPage is the page for configuring the users output server.
type OutputServerPage struct {
	display *InstallDisplay
	page    *tui.Page
	content tview.Primitive
	form    *tview.Form
}

// NewOutputServerPage creates a new OutputServerPage.
func NewOutputServerPage(display *InstallDisplay) *OutputServerPage {
	outputPage := &OutputServerPage{
		display: display,
	}

	outputPage.initPage()

	outputPage.page = tui.NewPage(
		display.beaconPage.GetPage(), // Set parent to beacon page
		"install-output",
		"Output Server",
		"Select the output server you'd like to use",
		outputPage.content,
	)

	return outputPage
}

// GetPage returns the page.
func (p *OutputServerPage) GetPage() *tui.Page {
	return p.page
}

// initPage initializes the page.
func (p *OutputServerPage) initPage() {
	var (
		modalWidth   = 70
		lines        = tview.WordWrap("Select which output server you'd like to use", modalWidth-4)
		height       = len(lines) + 4
		serverLabels = make([]string, len(tui.AvailableOutputServers))
	)

	// Create server options available to the user.
	for i, server := range tui.AvailableOutputServers {
		serverLabels[i] = server.Label
	}

	// We need a form to house our dropdown and input field.
	form := tview.NewForm()
	form.SetBackgroundColor(tui.ColorFormBackground)
	form.SetBorderPadding(0, 0, 0, 0)
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetLabelColor(tcell.ColorLightGray)
	form.SetButtonBackgroundColor(tui.ColorFormBackground)
	form.SetButtonsAlign(tview.AlignCenter)
	p.form = form

	// Get current selection from config
	if p.display.sidecarCfg.Get().OutputServer == nil {
		if err := p.display.sidecarCfg.Update(func(cfg *config.Config) {
			cfg.OutputServer = &config.OutputServer{}
		}); err != nil {
			p.openErrorModal(err)

			return
		}
	}

	currentAddress := p.display.sidecarCfg.Get().OutputServer.Address
	defaultIndex := 0 // Default to first option

	// Check if it's a custom server address.
	if currentAddress != "" {
		if !strings.Contains(currentAddress, "platform.ethpandaops.io") {
			// Set to Custom option.
			for i, server := range tui.AvailableOutputServers {
				if server.Label == "Custom" {
					defaultIndex = i

					break
				}
			}
		} else {
			// Find matching ethPandaOps server.
			for i, server := range tui.AvailableOutputServers {
				if server.Value == currentAddress {
					defaultIndex = i

					break
				}
			}
		}
	}

	// Add dropdown with proper background and current selection.
	form.AddDropDown("Output Server", serverLabels, defaultIndex, func(text string, index int) {
		selectedValue := tui.AvailableOutputServers[index].Value

		// Remove any existing Server Address field first.
		if item := form.GetFormItemByLabel("Server Address"); item != nil {
			form.RemoveFormItem(form.GetFormItemIndex("Server Address"))
		}

		// Clear credentials when switching server types
		currentAddress := p.display.sidecarCfg.Get().OutputServer.Address
		wasEthPandaOps := validate.IsEthPandaOpsServer(currentAddress)
		isEthPandaOps := validate.IsEthPandaOpsServer(selectedValue)

		if wasEthPandaOps != isEthPandaOps {
			// Server type changed, clear credentials
			if err := p.display.sidecarCfg.Update(func(cfg *config.Config) {
				cfg.OutputServer.Credentials = ""
			}); err != nil {
				p.openErrorModal(err)

				return
			}
		}

		// Handle custom server field.
		if selectedValue == "custom" {
			// If we're switching to custom, preserve existing custom address.
			existingAddress := p.display.sidecarCfg.Get().OutputServer.Address
			if strings.Contains(existingAddress, "platform.ethpandaops.io") {
				existingAddress = ""
			}

			if err := p.display.sidecarCfg.Update(func(cfg *config.Config) {
				cfg.OutputServer.Address = existingAddress
			}); err != nil {
				p.openErrorModal(err)

				return
			}

			input := form.AddInputField("Server Address", existingAddress, 40, nil, func(address string) {
				if err := p.display.sidecarCfg.Update(func(cfg *config.Config) {
					cfg.OutputServer.Address = address
				}); err != nil {
					p.openErrorModal(err)

					return
				}
			})
			input.SetBackgroundColor(tui.ColorFormBackground)
		} else {
			// Only update config when explicitly selecting a standard server.
			if err := p.display.sidecarCfg.Update(func(cfg *config.Config) {
				cfg.OutputServer.Address = selectedValue
			}); err != nil {
				p.openErrorModal(err)

				return
			}
		}
	})

	// Set dropdown width and trigger initial selection.
	if dropdown, ok := form.GetFormItemByLabel("Output Server").(*tview.DropDown); ok {
		dropdown.SetFieldWidth(50)
		dropdown.SetCurrentOption(defaultIndex)
	}

	// Add Next button with padding.
	form.AddButton(tui.ButtonNext, func() {
		dropdown, _ := form.GetFormItemByLabel("Output Server").(*tview.DropDown)
		selectedIndex, _ := dropdown.GetCurrentOption()
		selectedValue := tui.AvailableOutputServers[selectedIndex].Value
		isCustom := selectedValue == "custom"

		var address string

		if isCustom {
			if input := form.GetFormItemByLabel("Server Address"); input != nil {
				if inputField, ok := input.(*tview.InputField); ok {
					address = inputField.GetText()
				} else {
					p.openErrorModal(fmt.Errorf("invalid input field type"))

					return
				}
			}

			if err := validate.ValidateOutputServerAddress(address); err != nil {
				p.openErrorModal(err)

				return
			}
		} else {
			address = selectedValue
		}

		// Update config with validated address
		if err := p.display.sidecarCfg.Update(func(cfg *config.Config) {
			cfg.OutputServer.Address = address
		}); err != nil {
			p.openErrorModal(err)

			return
		}

		p.display.setPage(p.display.outputServerCredentialsPage.GetPage())
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
	}

	// Create content grid.
	contentGrid := tview.NewGrid()
	contentGrid.SetRows(2, 3, 1, 6, 1, 2)
	contentGrid.SetColumns(1, -4, 1)
	contentGrid.SetBackgroundColor(tui.ColorFormBackground)

	// Create the main text view.
	textView := tview.NewTextView()
	textView.SetText("Select which output server you'd like to use")
	textView.SetTextAlign(tview.AlignCenter)
	textView.SetWordWrap(true)
	textView.SetTextColor(tview.Styles.PrimaryTextColor)
	textView.SetBackgroundColor(tui.ColorFormBackground)
	textView.SetBorderPadding(0, 0, 0, 0)

	// Add items to content grid.
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 0, 0, 1, 3, 0, 0, false)
	contentGrid.AddItem(textView, 1, 0, 1, 3, 0, 0, false)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 2, 0, 1, 3, 0, 0, false)
	contentGrid.AddItem(form, 3, 0, 1, 3, 0, 0, true)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 5, 0, 1, 3, 0, 0, false)
	contentGrid.SetBorder(true)
	contentGrid.SetTitle(" 🌎 Output Server ")
	contentGrid.SetBackgroundColor(tui.ColorFormBackground)

	// Create border grid.
	borderGrid := tview.NewGrid()
	borderGrid.SetColumns(0, modalWidth, 0)
	borderGrid.SetRows(0, height+9, 0, 2)
	borderGrid.SetBackgroundColor(tui.ColorFormBackground)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	p.content = borderGrid
}

func (p *OutputServerPage) openErrorModal(err error) {
	p.display.app.SetRoot(tui.CreateErrorModal(
		p.display.app,
		err.Error(),
		func() {
			p.display.app.SetRoot(p.display.frame, true).EnableMouse(true)
		},
	), true).EnableMouse(true)
}
