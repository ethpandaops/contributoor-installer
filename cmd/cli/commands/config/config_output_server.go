package config

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// OutputServerConfigPage is the page for configuring the output server.
type OutputServerConfigPage struct {
	display     *ConfigDisplay
	page        *page
	content     tview.Primitive
	form        *tview.Form
	description *tview.TextView
}

// NewOutputServerConfigPage creates a new OutputServerConfigPage.
func NewOutputServerConfigPage(display *ConfigDisplay) *OutputServerConfigPage {
	OutputServerConfigPage := &OutputServerConfigPage{
		display: display,
	}

	OutputServerConfigPage.initPage()
	OutputServerConfigPage.page = newPage(
		display.homePage,
		"config-output-server",
		"Output Server Settings",
		"Configure the output server settings including server selection and credentials",
		OutputServerConfigPage.content,
	)

	return OutputServerConfigPage
}

// GetPage returns the page.
func (p *OutputServerConfigPage) GetPage() *display.Page {
	return p.page
}

// initPage initializes the page.
func (p *OutputServerConfigPage) initPage() {
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

	// Define our field descriptions.
	descriptions := map[string]string{
		"Output Server":  "Select the output server to send your data to.",
		"Username":       "Your output server username for authentication.",
		"Password":       "Your output server password for authentication.",
		"Server Address": "The address of your custom output server.",
	}

	// Pull together a list of possible output servers and their descriptions.
	serverLabels := make([]string, len(display.AvailableOutputServers))
	serverDescriptions := make(map[string]string)
	for i, server := range display.AvailableOutputServers {
		serverLabels[i] = server.Label
		serverDescriptions[server.Label] = server.Description
		descriptions["Output Server"] = server.Description
	}

	// We might already have a server configured, so we need to find the index
	// of the current server so we can prepopulate the form with the current
	// values.
	defaultIndex := 0
	currentAddress := p.display.configService.Get().OutputServer.Address

	// Check if it's a custom output server address.
	if !strings.Contains(currentAddress, "platform.ethpandaops.io") {
		// Set to Custom option
		for i, server := range display.AvailableOutputServers {
			if server.Label == "Custom" {
				defaultIndex = i
				break
			}
		}
	} else {
		// Otherwise, it'll be an ethPandaOps server.
		for i, server := range display.AvailableOutputServers {
			if server.Value == currentAddress {
				defaultIndex = i
				break
			}
		}
	}

	// Setup our dropdown to select the output server.
	dropdown := tview.NewDropDown().
		SetLabel("Output Server ").
		SetOptions(serverLabels, func(option string, index int) {
			// Update description when server changes.
			p.description.SetText(serverDescriptions[option])

			// We've got to do some trickery here. Remove all fields except the dropdown
			// and then add the appropriate fields based on the selection.
			for i := form.GetFormItemCount() - 1; i > 0; i-- {
				form.RemoveFormItem(i)
			}

			// Add appropriate fields based on selection.
			if option == "Custom" {
				// If it's a custom server, we need to add the server address field.
				defaultAddress := p.display.configService.Get().OutputServer.Address
				if strings.Contains(defaultAddress, "platform.ethpandaops.io") {
					defaultAddress = ""
				}

				// Add the server address field.
				form.AddInputField("Server Address", defaultAddress, 0, nil, nil)

				// Add the username and password fields.
				username, password := getCredentialsFromConfig(p.display.configService.Get())
				form.AddInputField("Username", username, 0, nil, nil)
				form.AddPasswordField("Password", password, 0, '*', nil)
			} else {
				// Otherwise, it's an ethPandaOps server.
				username, password := getCredentialsFromConfig(p.display.configService.Get())
				form.AddInputField("Username", username, 0, nil, nil)
				form.AddPasswordField("Password", password, 0, '*', nil)
			}

			p.display.app.SetFocus(form)
		})

	// Add dropdown to our form and set initial selection.
	form.AddFormItem(dropdown)
	dropdown.SetCurrentOption(defaultIndex)

	// Build our save button out.
	saveButton := tview.NewButton(display.ButtonSaveSettings)
	saveButton.SetSelectedFunc(func() {
		validateAndUpdateOutputServer(p)
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
		// Get the currently focused item
		formIndex, _ := form.GetFocusedItemIndex()

		switch event.Key() {
		case tcell.KeyTab:
			// If we're on the last form item, move to save button
			if formIndex == form.GetFormItemCount()-1 {
				p.display.app.SetFocus(saveButton)
				return nil
			}

			return event
		case tcell.KeyBacktab:
			// If we're on the first form item, move to save button
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
	formFrame.SetTitle("Output Server Settings")
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

	// Create a main layout to hold the form and description and save button.
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(formDescriptionFlex, 0, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(buttonFlex, 1, 0, false).
		AddItem(nil, 1, 0, false)
	mainFlex.SetBackgroundColor(display.ColorBackground)

	p.content = mainFlex
}

func validateAndUpdateOutputServer(p *OutputServerConfigPage) {
	// Get the currently selected server.
	dropdown := p.form.GetFormItem(0).(*tview.DropDown)
	_, serverLabel := dropdown.GetCurrentOption()

	// Find the corresponding URL
	var serverAddress string
	for _, server := range display.AvailableOutputServers {
		if server.Label == serverLabel {
			serverAddress = server.Value
			break
		}
	}

	// If it's a custom server, we need to get the server address and validate it.
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

		// Validate URL format.
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

		// Set the server address.
		serverAddress = customAddress

		// Get credentials, these are optional for custom servers.
		var (
			username     = p.form.GetFormItem(2).(*tview.InputField)
			password     = p.form.GetFormItem(3).(*tview.InputField)
			usernameText = username.GetText()
			passwordText = password.GetText()
		)

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
		// Get and validate credentials, these are required for ethPandaOps servers.
		var (
			username     = p.form.GetFormItem(1).(*tview.InputField)
			password     = p.form.GetFormItem(2).(*tview.InputField)
			usernameText = username.GetText()
			passwordText = password.GetText()
		)

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

		// Update credentials.
		credentials := fmt.Sprintf("%s:%s", usernameText, passwordText)
		p.display.configService.Update(func(cfg *service.ContributoorConfig) {
			cfg.OutputServer.Address = serverAddress
			cfg.OutputServer.Credentials = base64.StdEncoding.EncodeToString([]byte(credentials))
		})
	}

	p.display.setPage(p.display.homePage)
}

// getCredentialsFromConfig is a helper function to get the user/pass from the config.
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
