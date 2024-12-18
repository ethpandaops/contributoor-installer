package install

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type BeaconNodePage struct {
	display     *InstallDisplay
	page        *page
	content     tview.Primitive
	form        *tview.Form
	description *tview.TextView
}

func NewBeaconNodePage(display *InstallDisplay) *BeaconNodePage {
	beaconPage := &BeaconNodePage{
		display: display,
	}

	beaconPage.initPage()
	beaconPage.page = newPage(
		display.networkPage.GetPage(), // Set parent to network page
		"install-beacon",
		"Beacon Node",
		"Configure your beacon node connection",
		beaconPage.content,
	)

	return beaconPage
}

func (p *BeaconNodePage) GetPage() *display.Page {
	return p.page
}

func (p *BeaconNodePage) initPage() {
	// Layout components
	var (
		// Calculate dimensions
		modalWidth     = 70
		lines          = tview.WordWrap("Please enter the address of your Beacon Node.\nFor example: http://localhost:5052", modalWidth-4)
		textViewHeight = len(lines) + 4
		formHeight     = 3 // Input field + padding

		// Main grids
		contentGrid = tview.NewGrid()
		borderGrid  = tview.NewGrid().SetColumns(0, modalWidth, 0)

		// Form components
		form = tview.NewForm()
	)

	// Initialize form
	form.SetButtonsAlign(tview.AlignCenter)
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetBackgroundColor(display.ColorFormBackground)
	form.SetBorderPadding(0, 0, 0, 0) // Reset padding
	form.SetLabelColor(tcell.ColorLightGray)

	// Add input field with more visible styling
	inputField := tview.NewInputField().
		SetLabel("Beacon Node Address: ").
		SetText(p.display.configService.Get().BeaconNodeAddress).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetLabelColor(tcell.ColorLightGray)
	form.AddFormItem(inputField)
	p.form = form

	// Create form frame
	formFrame := tview.NewFrame(form)
	formFrame.SetBorderPadding(0, 0, 0, 0) // Reset padding
	formFrame.SetBackgroundColor(display.ColorFormBackground)

	// Add Next button
	form.AddButton(display.ButtonNext, func() {
		validateAndUpdate(p)
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

	// Create the main text view
	textView := tview.NewTextView()
	textView.SetText("Please enter the address of your Beacon Node.\nFor example: http://localhost:5052")
	textView.SetTextAlign(tview.AlignCenter)
	textView.SetWordWrap(true)
	textView.SetTextColor(tview.Styles.PrimaryTextColor)
	textView.SetBackgroundColor(display.ColorFormBackground)
	textView.SetBorderPadding(0, 0, 0, 0)

	// Content grid
	contentGrid.SetRows(2, 2, 1, 4, 1, 2, 2)
	contentGrid.SetBackgroundColor(display.ColorFormBackground)
	contentGrid.SetBorder(true)
	contentGrid.SetTitle(" Beacon Node ")

	// Add items to content grid with proper background boxes
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(display.ColorFormBackground), 0, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(textView, 1, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(display.ColorFormBackground), 2, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(formFrame, 3, 0, 2, 1, 0, 0, true)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(display.ColorFormBackground), 5, 0, 2, 1, 0, 0, false)

	// Border grid
	borderGrid.SetRows(0, textViewHeight+formHeight+4, 0, 2)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	// Set initial focus
	p.display.app.SetFocus(form)
	p.content = borderGrid
}

func validateAndUpdate(p *BeaconNodePage) {
	// Get text from the input field directly
	inputField := p.form.GetFormItem(0).(*tview.InputField)
	address := inputField.GetText()

	// Show loading modal while validating
	loadingModal := display.CreateLoadingModal(
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
				errorModal := display.CreateErrorModal(
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
			p.display.configService.Update(func(cfg *service.ContributoorConfig) {
				cfg.BeaconNodeAddress = address
			})

			// Move to next page
			p.display.setPage(p.display.outputPage.GetPage())
		})
	}()
}

func validateBeaconNode(address string) error {
	// Check if URL is valid
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		return fmt.Errorf("Beacon node address must start with http:// or https://")
	}

	// Try to connect to the beacon node
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(fmt.Sprintf("%s/eth/v1/node/health", address))
	if err != nil {
		return fmt.Errorf("We're unable to connect to your beacon node: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Beacon node returned status %d", resp.StatusCode)
	}

	return nil
}
