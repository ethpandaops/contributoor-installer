package install

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
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
		lines          = tview.WordWrap("Please enter the address of your Beacon Node.\nFor example: http://localhost:5052", modalWidth-4)
		textViewHeight = len(lines) + 4
		formHeight     = 3 // Input field + a bit of padding.

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
		SetText(p.display.configService.Get().BeaconNodeAddress).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetLabelColor(tcell.ColorLightGray)
	form.AddFormItem(inputField)

	// Add our form to the page for easy access during validation.
	p.form = form

	// Wrap our form in a frame to add a border.
	formFrame := tview.NewFrame(form)
	formFrame.SetBorderPadding(0, 0, 0, 0) // Reset padding
	formFrame.SetBackgroundColor(tui.ColorFormBackground)

	// Add 'Next' button to our form.
	form.AddButton(tui.ButtonNext, func() {
		validateAndUpdate(p)
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

	// Create the main text view.
	textView := tview.NewTextView()
	textView.SetText("Please enter the address of your Beacon Node.\nFor example: http://localhost:5052")
	textView.SetTextAlign(tview.AlignCenter)
	textView.SetWordWrap(true)
	textView.SetTextColor(tview.Styles.PrimaryTextColor)
	textView.SetBackgroundColor(tui.ColorFormBackground)
	textView.SetBorderPadding(0, 0, 0, 0)

	// Set up the content grid.
	contentGrid.SetRows(2, 2, 1, 4, 1, 2, 2)
	contentGrid.SetBackgroundColor(tui.ColorFormBackground)
	contentGrid.SetBorder(true)
	contentGrid.SetTitle(" Beacon Node ")

	// Add items to content grid using spacers.
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 0, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(textView, 1, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 2, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(formFrame, 3, 0, 2, 1, 0, 0, true)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 5, 0, 2, 1, 0, 0, false)

	// Border grid.
	borderGrid.SetRows(0, textViewHeight+formHeight+4, 0, 2)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	// Set initial focus.
	p.display.app.SetFocus(form)
	p.content = borderGrid
}

func validateAndUpdate(p *BeaconNodePage) {
	// Get text from the input field directly
	inputField, _ := p.form.GetFormItem(0).(*tview.InputField)
	address := inputField.GetText()

	// Show loading modal while validating
	loadingModal := tui.CreateLoadingModal(
		p.display.app,
		"\n[yellow]Validating beacon node connection...\nPlease wait...[white]",
	)
	p.display.app.SetRoot(loadingModal, true)

	// Validate in goroutine to not block UI
	go func() {
		err := validateBeaconNode(address)

		p.display.app.QueueUpdateDraw(func() {
			if err != nil {
				// Show error modal
				errorModal := tui.CreateErrorModal(
					p.display.app,
					err.Error(),
					func() {
						p.display.app.SetRoot(p.display.frame, true)
						p.display.app.SetFocus(p.form)
					},
				)

				p.display.app.SetRoot(errorModal, true)

				return
			}

			// Update config if validation passes
			if err := p.display.configService.Update(func(cfg *service.ContributoorConfig) {
				cfg.BeaconNodeAddress = address
			}); err != nil {
				p.openErrorModal(err)

				return
			}

			// Move to next page
			p.display.setPage(p.display.outputPage.GetPage())
		})
	}()
}

func validateBeaconNode(address string) error {
	// Check if URL is valid
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		return fmt.Errorf("beacon node address must start with http:// or https://")
	}

	// Try to connect to the beacon node
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(fmt.Sprintf("%s/eth/v1/node/health", address))
	if err != nil {
		return fmt.Errorf("we're unable to connect to your beacon node: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("beacon node returned status %d", resp.StatusCode)
	}

	return nil
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
