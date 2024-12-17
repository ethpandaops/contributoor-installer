package settings

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type outputServerOption struct {
	Label       string
	Value       string
	Description string
}

var availableOutputServers = []outputServerOption{
	{
		Label:       "ethPandaOps Production",
		Value:       "https://xatu.primary.production.platform.ethpandaops.io",
		Description: "The production server provided by ethPandaOps.",
	},
	{
		Label:       "ethPandaOps Staging",
		Value:       "https://xatu.primary.staging.platform.ethpandaops.io",
		Description: "The staging server provided by ethPandaOps.",
	},
	{
		Label:       "Custom",
		Value:       "custom",
		Description: "Use your own custom output server.",
	},
}

type OutputServerConfigPage struct {
	display     *ConfigDisplay
	page        *page
	content     tview.Primitive
	form        *tview.Form
	description *tview.TextView
}

func NewOutputServerConfigPage(display *ConfigDisplay) *OutputServerConfigPage {
	OutputServerConfigPage := &OutputServerConfigPage{
		display: display,
	}

	OutputServerConfigPage.createContent()
	OutputServerConfigPage.page = newPage(
		display.homePage,
		"config-output-server",
		"Output Server Settings",
		"Configure the output server settings including server selection and credentials",
		OutputServerConfigPage.content,
	)

	return OutputServerConfigPage
}

func (p *OutputServerConfigPage) getPage() *page {
	return p.page
}

func (p *OutputServerConfigPage) handleLayoutChanged() {
	// Implement if needed
}

func (p *OutputServerConfigPage) createContent() {
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
		"Output Server":  "Select the output server to send your data to.",
		"Username":       "Your output server username for authentication.",
		"Password":       "Your output server password for authentication.",
		"Server Address": "The address of your custom output server.",
	}

	// Server options
	serverLabels := make([]string, len(availableOutputServers))
	for i, server := range availableOutputServers {
		serverLabels[i] = server.Label
	}

	// Find current server index based on the URL
	defaultIndex := 0
	currentAddress := p.display.configService.Get().OutputServer.Address

	// Check if it's a custom address
	if currentAddress != "" && !strings.Contains(currentAddress, "platform.ethpandaops.io") {
		// Set to Custom option
		for i, server := range availableOutputServers {
			if server.Label == "Custom" {
				defaultIndex = i
				break
			}
		}
	} else {
		// Find matching ethPandaOps server
		for i, server := range availableOutputServers {
			if server.Value == currentAddress {
				defaultIndex = i
				break
			}
		}
	}

	// Add form fields without immediate config updates
	dropdown := tview.NewDropDown().
		SetLabel("Output Server ").
		SetOptions(serverLabels, func(option string, index int) {
			// Remove all fields except the dropdown
			for i := form.GetFormItemCount() - 1; i > 0; i-- {
				form.RemoveFormItem(i)
			}

			// Add appropriate fields based on selection
			if option == "Custom" {
				defaultAddress := p.display.configService.Get().OutputServer.Address
				if strings.Contains(defaultAddress, "platform.ethpandaops.io") {
					defaultAddress = ""
				}

				username, password := getCredentialsFromConfig(p.display.configService.Get())
				form.AddInputField("Server Address", defaultAddress, 0, nil, nil)
				form.AddInputField("Username", username, 0, nil, nil)
				form.AddPasswordField("Password", password, 0, '*', nil)
			} else {
				username, password := getCredentialsFromConfig(p.display.configService.Get())
				form.AddInputField("Username", username, 0, nil, nil)
				form.AddPasswordField("Password", password, 0, '*', nil)
			}
			p.display.app.SetFocus(form)
		})

	// Add dropdown to form and set initial selection
	form.AddFormItem(dropdown)
	dropdown.SetCurrentOption(defaultIndex)

	// Create save button with validation
	saveButton := tview.NewButton(display.ButtonSaveSettings)
	saveButton.SetSelectedFunc(func() {
		validateAndUpdateOutputServer(p)
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

	// Set up form input capture with access to saveButton
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
	formFrame.SetTitle("Output Server Settings")
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
func (p *OutputServerConfigPage) updateDescription(text string) {
	p.description.SetText(text)
}

func validateAndUpdateOutputServer(p *OutputServerConfigPage) {
	dropdown := p.form.GetFormItem(0).(*tview.DropDown)
	_, serverLabel := dropdown.GetCurrentOption()

	// Find the corresponding URL
	var serverAddress string
	for _, server := range availableOutputServers {
		if server.Label == serverLabel {
			serverAddress = server.Value
			break
		}
	}

	if serverAddress == "custom" {
		// Get and validate custom address
		customAddress := p.form.GetFormItem(1).(*tview.InputField).GetText()
		if customAddress == "" {
			errorModal := display.CreateErrorModal(
				p.display.app,
				"Server address is required for custom server",
				func() {
					p.display.app.SetRoot(p.display.frame, true)
					p.display.app.SetFocus(p.form)
				},
			)
			p.display.app.SetRoot(errorModal, true)
			return
		}

		// Validate URL format
		if !strings.HasPrefix(customAddress, "http://") && !strings.HasPrefix(customAddress, "https://") {
			errorModal := display.CreateErrorModal(
				p.display.app,
				"Server address must start with http:// or https://",
				func() {
					p.display.app.SetRoot(p.display.frame, true)
					p.display.app.SetFocus(p.form)
				},
			)
			p.display.app.SetRoot(errorModal, true)
			return
		}

		serverAddress = customAddress

		// Get optional credentials
		username := p.form.GetFormItem(2).(*tview.InputField)
		password := p.form.GetFormItem(3).(*tview.InputField)
		usernameText := username.GetText()
		passwordText := password.GetText()

		// Only set credentials if both username and password are provided
		if usernameText != "" && passwordText != "" {
			credentials := fmt.Sprintf("%s:%s", usernameText, passwordText)
			p.display.configService.Update(func(cfg *service.ContributoorConfig) {
				cfg.OutputServer.Address = serverAddress
				cfg.OutputServer.Credentials = base64.StdEncoding.EncodeToString([]byte(credentials))
			})
		} else if usernameText == "" && passwordText == "" {
			// Both empty - clear credentials
			p.display.configService.Update(func(cfg *service.ContributoorConfig) {
				cfg.OutputServer.Address = serverAddress
				cfg.OutputServer.Credentials = ""
			})
		} else {
			// One is empty but not both
			errorModal := display.CreateErrorModal(
				p.display.app,
				"Both username and password must be provided if using credentials",
				func() {
					p.display.app.SetRoot(p.display.frame, true)
					p.display.app.SetFocus(p.form)
				},
			)
			p.display.app.SetRoot(errorModal, true)
			return
		}
	} else {
		// Get and validate credentials
		username := p.form.GetFormItem(1).(*tview.InputField)
		password := p.form.GetFormItem(2).(*tview.InputField)
		usernameText := username.GetText()
		passwordText := password.GetText()

		if usernameText == "" || passwordText == "" {
			errorModal := display.CreateErrorModal(
				p.display.app,
				"Username and password are required for ethPandaOps servers",
				func() {
					p.display.app.SetRoot(p.display.frame, true)
					p.display.app.SetFocus(p.form)
				},
			)
			p.display.app.SetRoot(errorModal, true)
			return
		}

		// Update credentials
		credentials := fmt.Sprintf("%s:%s", usernameText, passwordText)
		p.display.configService.Update(func(cfg *service.ContributoorConfig) {
			cfg.OutputServer.Address = serverAddress
			cfg.OutputServer.Credentials = base64.StdEncoding.EncodeToString([]byte(credentials))
		})
	}

	// Update config
	p.display.configService.Update(func(cfg *service.ContributoorConfig) {
		cfg.OutputServer.Address = serverAddress
		if serverAddress == "custom" {
			cfg.OutputServer.Credentials = "" // Clear credentials for custom server
		}
	})

	// Return to settings home
	p.display.setPage(p.display.homePage)
}

// Helper to decode credentials
func getCredentialsFromConfig(cfg *service.ContributoorConfig) (username, password string) {
	if cfg.OutputServer.Credentials == "" {
		return "", ""
	}

	decoded, err := base64.StdEncoding.DecodeString(cfg.OutputServer.Credentials)
	if err != nil {
		return "", ""
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) != 2 {
		return "", ""
	}

	return parts[0], parts[1]
}
