package config

import (
	"fmt"
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor-installer/internal/validate"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// OutputServerConfigPage is the page for configuring the output server.
type OutputServerConfigPage struct {
	display     *ConfigDisplay
	page        *tui.Page
	content     tview.Primitive
	form        *tview.Form
	description *tview.TextView
}

// NewOutputServerConfigPage creates a new OutputServerConfigPage.
func NewOutputServerConfigPage(cd *ConfigDisplay) *OutputServerConfigPage {
	OutputServerConfigPage := &OutputServerConfigPage{
		display: cd,
	}

	OutputServerConfigPage.initPage()
	OutputServerConfigPage.page = tui.NewPage(
		cd.homePage,
		"config-output-server",
		"Output Server Settings",
		"Configure the output server settings including server selection and credentials",
		OutputServerConfigPage.content,
	)

	return OutputServerConfigPage
}

// GetPage returns the page.
func (p *OutputServerConfigPage) GetPage() *tui.Page {
	return p.page
}

// initPage initializes the page.
func (p *OutputServerConfigPage) initPage() {
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

	// Define our field descriptions.
	descriptions := map[string]string{
		"Output Server":  "Select the output server to send your data to.",
		"Username":       "Your output server username for authentication.",
		"Password":       "Your output server password for authentication.",
		"Server Address": "The address of your custom output server.",
	}

	// Pull together a list of possible output servers and their descriptions.
	serverLabels := make([]string, len(tui.AvailableOutputServers))
	serverDescriptions := make(map[string]string)

	for i, server := range tui.AvailableOutputServers {
		serverLabels[i] = server.Label
		serverDescriptions[server.Label] = server.Description
		descriptions["Output Server"] = server.Description
	}

	// We might already have a server configured, so we need to find the index
	// of the current server so we can prepopulate the form with the current
	// values.
	defaultIndex := 0
	currentAddress := p.display.sidecarCfg.Get().OutputServer.Address

	// Check if it's a custom output server address.
	if !strings.Contains(currentAddress, "platform.ethpandaops.io") {
		// Set to Custom option
		for i, server := range tui.AvailableOutputServers {
			if server.Label == "Custom" {
				defaultIndex = i

				break
			}
		}
	} else {
		// Otherwise, it'll be an ethPandaOps server.
		for i, server := range tui.AvailableOutputServers {
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
				defaultAddress := p.display.sidecarCfg.Get().OutputServer.Address
				if strings.Contains(defaultAddress, "platform.ethpandaops.io") {
					defaultAddress = ""
				}

				// Add the server address field.
				serverInput := tview.NewInputField().
					SetLabel("Server Address").
					SetText(defaultAddress)
				serverInput.SetFocusFunc(func() {
					p.description.SetText("The address of your custom output server (e.g., myserver.com:443)")
				})
				form.AddFormItem(serverInput)

				// Add the username and password fields.
				username, password := getCredentialsFromConfig(p.display.sidecarCfg.Get())
				usernameInput := tview.NewInputField().
					SetLabel("Username").
					SetText(username)
				usernameInput.SetFocusFunc(func() {
					p.description.SetText("Your output server username for authentication")
				})
				form.AddFormItem(usernameInput)

				form.AddPasswordField("Password", password, 0, '*', nil).
					SetFocusFunc(func() {
						p.description.SetText("Your output server password for authentication")
					})
			} else {
				// Otherwise, it's an ethPandaOps server.
				username, password := getCredentialsFromConfig(p.display.sidecarCfg.Get())
				usernameInput := tview.NewInputField().
					SetLabel("Username").
					SetText(username)
				usernameInput.SetFocusFunc(func() {
					p.description.SetText("Your ethPandaOps platform username for authentication")
				})
				form.AddFormItem(usernameInput)

				form.AddPasswordField("Password", password, 0, '*', nil).
					SetFocusFunc(func() {
						p.description.SetText("Your ethPandaOps platform password for authentication")
					})
			}

			p.display.app.SetFocus(form)
		})

	// Add dropdown to our form and set initial selection.
	form.AddFormItem(dropdown)
	dropdown.SetCurrentOption(defaultIndex)

	// Build our save button out.
	saveButton := tview.NewButton(tui.ButtonSaveSettings)
	saveButton.SetSelectedFunc(func() {
		validateAndUpdateOutputServer(p)
	})
	saveButton.SetBackgroundColorActivated(tui.ColorButtonActivated)
	saveButton.SetLabelColorActivated(tui.ColorButtonText)

	// Define key bindings for the save button.
	saveButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			// When tabbing from save button, go back to first form item.
			form.SetFocus(0)
			p.display.app.SetFocus(form)

			return nil
		case tcell.KeyBacktab:
			// When back-tabbing from save button, go to last form item.
			form.SetFocus(form.GetFormItemCount() - 1)
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
	formFrame.SetTitle("Output Server Settings")
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

	// Create a main layout to hold the form and description and save button.
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(formDescriptionFlex, 0, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(buttonFlex, 1, 0, false).
		AddItem(nil, 1, 0, false)
	mainFlex.SetBackgroundColor(tui.ColorBackground)

	p.content = mainFlex
}

func validateAndUpdateOutputServer(p *OutputServerConfigPage) {
	// Get the currently selected server
	dropdown, _ := p.form.GetFormItem(0).(*tview.DropDown)
	_, serverLabel := dropdown.GetCurrentOption()

	// Find the corresponding URL
	var serverAddress string

	for _, server := range tui.AvailableOutputServers {
		if server.Label == serverLabel {
			serverAddress = server.Value

			break
		}
	}

	// Get form values based on server type
	var (
		isCustom  = serverAddress == "custom"
		username  string
		password  string
		formStart = 1
	)

	if isCustom {
		// For custom servers, get address from input field.
		var address string

		if item := p.form.GetFormItem(formStart); item != nil {
			if inputField, ok := item.(*tview.InputField); ok {
				address = inputField.GetText()
			} else {
				p.openErrorModal(fmt.Errorf("invalid address field type"))

				return
			}
		}

		if err := validate.ValidateOutputServerAddress(address); err != nil {
			p.openErrorModal(err)

			return
		}

		serverAddress = address
		formStart++
	}

	// Get credentials from form.
	if formItem := p.form.GetFormItem(formStart); formItem != nil {
		if inputField, ok := formItem.(*tview.InputField); ok {
			username = inputField.GetText()
		} else {
			p.openErrorModal(fmt.Errorf("invalid username field type"))

			return
		}
	}

	if formItem := p.form.GetFormItem(formStart + 1); formItem != nil {
		if inputField, ok := formItem.(*tview.InputField); ok {
			password = inputField.GetText()
		} else {
			p.openErrorModal(fmt.Errorf("invalid password field type"))

			return
		}
	}

	// Validate credentials. These are optional for custom servers.
	if err := validate.ValidateOutputServerCredentials(
		username,
		password,
		validate.IsEthPandaOpsServer(serverAddress),
	); err != nil {
		p.openErrorModal(err)

		return
	}

	// Update config with validated values.
	if err := p.display.sidecarCfg.Update(func(cfg *config.Config) {
		cfg.OutputServer.Address = serverAddress

		if username != "" && password != "" {
			cfg.OutputServer.Credentials = validate.EncodeCredentials(username, password)
		} else {
			cfg.OutputServer.Credentials = ""
		}

		if validate.IsEthPandaOpsServer(serverAddress) {
			cfg.OutputServer.Tls = true
		}
	}); err != nil {
		p.openErrorModal(err)

		return
	}

	p.display.markConfigChanged()
	p.display.setPage(p.display.homePage)
}

func (p *OutputServerConfigPage) openErrorModal(err error) {
	p.display.app.SetRoot(tui.CreateErrorModal(
		p.display.app,
		err.Error(),
		func() {
			p.display.app.SetRoot(p.display.frame, true)
		},
	), true)
}

// Update getCredentialsFromConfig to use the validation package.
func getCredentialsFromConfig(cfg *config.Config) (username, password string) {
	username, password, err := validate.DecodeCredentials(cfg.OutputServer.Credentials)
	if err != nil {
		return "", ""
	}

	return username, password
}
