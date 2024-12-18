package install

import (
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
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
	currentAddress := p.display.configService.Get().OutputServer.Address
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

		// Handle custom server field.
		if selectedValue == "custom" {
			// If we're switching to custom, preserve existing custom address.
			existingAddress := p.display.configService.Get().OutputServer.Address
			if strings.Contains(existingAddress, "platform.ethpandaops.io") {
				existingAddress = ""
				p.display.configService.Update(func(cfg *service.ContributoorConfig) {
					cfg.OutputServer.Address = existingAddress
				})
			}

			input := form.AddInputField("Server Address", existingAddress, 40, nil, func(address string) {
				p.display.configService.Update(func(cfg *service.ContributoorConfig) {
					cfg.OutputServer.Address = address
				})
			})
			input.SetBackgroundColor(tui.ColorFormBackground)
		} else {
			// Only update config when explicitly selecting a standard server.
			p.display.configService.Update(func(cfg *service.ContributoorConfig) {
				cfg.OutputServer.Address = selectedValue
			})
		}
	})

	// Set dropdown width and trigger initial selection.
	if dropdown, ok := form.GetFormItemByLabel("Output Server").(*tview.DropDown); ok {
		dropdown.SetFieldWidth(50)
		dropdown.SetCurrentOption(defaultIndex)
	}

	// Add Next button with padding.
	form.AddButton(tui.ButtonNext, func() {
		// Validate custom server address.
		selectedValue, _ := form.GetFormItemByLabel("Output Server").(*tview.DropDown).GetCurrentOption()
		if selectedValue == 2 { // Custom option.
			if input := form.GetFormItemByLabel("Server Address"); input != nil {
				address := input.(*tview.InputField).GetText()
				if address == "" {
					// Show error modal
					errorModal := tui.CreateErrorModal(
						p.display.app,
						"Server address is required for custom server",
						func() {
							p.display.app.SetRoot(p.display.frame, true)
							p.display.app.SetFocus(form)
						},
					)
					p.display.app.SetRoot(errorModal, true)
					return
				}

				// Validate URL format
				if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
					errorModal := tui.CreateErrorModal(
						p.display.app,
						"Server address must start with http:// or https://",
						func() {
							p.display.app.SetRoot(p.display.frame, true)
							p.display.app.SetFocus(form)
						},
					)
					p.display.app.SetRoot(errorModal, true)
					return
				}
			}
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
	contentGrid.SetTitle(" Output Server ")
	contentGrid.SetBackgroundColor(tui.ColorFormBackground)

	// Create border grid.
	borderGrid := tview.NewGrid()
	borderGrid.SetColumns(0, modalWidth, 0)
	borderGrid.SetRows(0, height+9, 0, 2)
	borderGrid.SetBackgroundColor(tui.ColorFormBackground)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	p.content = borderGrid
}
