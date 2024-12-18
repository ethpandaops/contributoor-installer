package install

import (
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type OutputServerPage struct {
	display *InstallDisplay
	page    *display.Page
	content tview.Primitive
	form    *tview.Form
}

func NewOutputServerPage(id *InstallDisplay) *OutputServerPage {
	outputPage := &OutputServerPage{
		display: id,
	}

	outputPage.initPage()

	outputPage.page = display.NewPage(
		id.beaconPage.GetPage(), // Set parent to beacon page
		"install-output",
		"Output Server",
		"Select the output server you'd like to use",
		outputPage.content,
	)

	return outputPage
}

func (p *OutputServerPage) GetPage() *display.Page {
	return p.page
}

func (p *OutputServerPage) initPage() {
	// Layout components
	var (
		modalWidth   = 70 // Match other pages' width
		lines        = tview.WordWrap("Select which output server you'd like to use", modalWidth-4)
		height       = len(lines) + 4
		serverLabels = make([]string, len(display.AvailableOutputServers))
	)

	// Create server options
	for i, server := range display.AvailableOutputServers {
		serverLabels[i] = server.Label
	}

	// Initialize form with proper background
	form := tview.NewForm()
	form.SetBackgroundColor(display.ColorFormBackground)
	form.SetBorderPadding(0, 0, 0, 0)
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetLabelColor(tcell.ColorLightGray)
	form.SetButtonBackgroundColor(display.ColorFormBackground)
	form.SetButtonsAlign(tview.AlignCenter)
	p.form = form

	// Get current selection from config
	currentAddress := p.display.configService.Get().OutputServer.Address
	defaultIndex := 0 // Default to first option

	// Check if it's a custom address
	if currentAddress != "" {
		if !strings.Contains(currentAddress, "platform.ethpandaops.io") {
			// Set to Custom option
			for i, server := range display.AvailableOutputServers {
				if server.Label == "Custom" {
					defaultIndex = i
					break
				}
			}
		} else {
			// Find matching ethPandaOps server
			for i, server := range display.AvailableOutputServers {
				if server.Value == currentAddress {
					defaultIndex = i
					break
				}
			}
		}
	}

	// Add dropdown with proper background and current selection
	form.AddDropDown("Output Server", serverLabels, defaultIndex, func(text string, index int) {
		selectedValue := display.AvailableOutputServers[index].Value

		// Remove any existing Server Address field first
		if item := form.GetFormItemByLabel("Server Address"); item != nil {
			form.RemoveFormItem(form.GetFormItemIndex("Server Address"))
		}

		// Handle custom server field
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
			input.SetBackgroundColor(display.ColorFormBackground)
		} else {
			// Only update config when explicitly selecting a standard server
			p.display.configService.Update(func(cfg *service.ContributoorConfig) {
				cfg.OutputServer.Address = selectedValue
			})
		}
	})

	// Set dropdown width and trigger initial selection
	if dropdown, ok := form.GetFormItemByLabel("Output Server").(*tview.DropDown); ok {
		dropdown.SetFieldWidth(50)              // Make dropdown wider
		dropdown.SetCurrentOption(defaultIndex) // This will trigger the handler for initial setup
	}

	// Add Next button with padding
	form.AddButton(display.ButtonNext, func() {
		// Validate custom server address
		selectedValue, _ := form.GetFormItemByLabel("Output Server").(*tview.DropDown).GetCurrentOption()
		if selectedValue == 2 { // Custom option
			if input := form.GetFormItemByLabel("Server Address"); input != nil {
				address := input.(*tview.InputField).GetText()
				if address == "" {
					// Show error modal
					errorModal := display.CreateErrorModal(
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
					errorModal := display.CreateErrorModal(
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

		// Proceed to next page
		p.display.setPage(p.display.outputServerCredentialsPage.GetPage())
	})

	// Style the button with proper background
	if button := form.GetButton(0); button != nil {
		button.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		button.SetLabelColor(tcell.ColorLightGray)
		form.SetButtonStyle(tcell.StyleDefault.
			Background(tview.Styles.PrimitiveBackgroundColor).
			Foreground(tcell.ColorLightGray))
		form.SetButtonActivatedStyle(tcell.StyleDefault.
			Background(display.ColorButtonActivated).
			Foreground(tcell.ColorBlack))
	}

	// Create content grid
	contentGrid := tview.NewGrid()
	contentGrid.SetRows(2, 3, 1, 6, 1, 2)
	contentGrid.SetColumns(1, -4, 1)
	contentGrid.SetBackgroundColor(display.ColorFormBackground)

	// Create the main text view
	textView := tview.NewTextView()
	textView.SetText("Select which output server you'd like to use")
	textView.SetTextAlign(tview.AlignCenter)
	textView.SetWordWrap(true)
	textView.SetTextColor(tview.Styles.PrimaryTextColor)
	textView.SetBackgroundColor(display.ColorFormBackground)
	textView.SetBorderPadding(0, 0, 0, 0)

	// Add items to content grid
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(display.ColorFormBackground), 0, 0, 1, 3, 0, 0, false)
	contentGrid.AddItem(textView, 1, 0, 1, 3, 0, 0, false)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(display.ColorFormBackground), 2, 0, 1, 3, 0, 0, false)
	contentGrid.AddItem(form, 3, 0, 1, 3, 0, 0, true)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(display.ColorFormBackground), 5, 0, 1, 3, 0, 0, false)

	// Create border grid
	borderGrid := tview.NewGrid()
	borderGrid.SetColumns(0, modalWidth, 0)
	borderGrid.SetRows(0, height+9, 0, 2)
	borderGrid.SetBackgroundColor(display.ColorFormBackground)

	contentGrid.SetBorder(true)
	contentGrid.SetTitle(" Output Server ")
	contentGrid.SetBackgroundColor(display.ColorFormBackground)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	p.content = borderGrid
}
