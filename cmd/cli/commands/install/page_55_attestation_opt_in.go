package install

import (
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AttestationOptInPage is the page for opting into attestation data contribution.
type AttestationOptInPage struct {
	display *InstallDisplay
	page    *tui.Page
	content tview.Primitive
	form    *tview.Form
	enabled bool
}

// NewAttestationOptInPage creates a new AttestationOptInPage.
func NewAttestationOptInPage(display *InstallDisplay) *AttestationOptInPage {
	attestationPage := &AttestationOptInPage{
		display: display,
		enabled: false, // Default to opt-out
	}

	attestationPage.initPage()
	attestationPage.page = tui.NewPage(
		display.outputServerCredentialsPage.GetPage(),
		"install-attestation",
		"Attestation Data Contribution",
		"Choose whether to contribute attestation data",
		attestationPage.content,
	)

	return attestationPage
}

// GetPage returns the page.
func (p *AttestationOptInPage) GetPage() *tui.Page {
	return p.page
}

// initPage initializes the page.
func (p *AttestationOptInPage) initPage() {
	var (
		modalWidth = 70
		height     = 13
	)

	// We need a form to house our checkbox.
	form := tview.NewForm()
	form.SetButtonsAlign(tview.AlignCenter)
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetBackgroundColor(tui.ColorFormBackground)
	form.SetBorderPadding(0, 0, 0, 0)
	form.SetLabelColor(tcell.ColorLightGray)
	p.form = form

	// Get existing value if any
	if cfg := p.display.sidecarCfg.Get(); cfg.AttestationSubnetCheck != nil {
		p.enabled = cfg.AttestationSubnetCheck.Enabled
	}

	// Add checkbox for attestation opt-in
	form.AddCheckbox("Enable attestation data contribution", p.enabled, func(checked bool) {
		p.enabled = checked
	})

	// Add a 'Next' button.
	form.AddButton(tui.ButtonNext, func() {
		p.saveAttestationPreference()
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

	// Create text view with bandwidth information.
	textView := tview.NewTextView()
	textView.SetText(`Would you like to contribute attestation data?

Contributing attestation data helps improve network analysis but will use a little more bandwidth.

`)
	textView.SetTextAlign(tview.AlignLeft)
	textView.SetWordWrap(true)
	textView.SetTextColor(tview.Styles.PrimaryTextColor)
	textView.SetBackgroundColor(tui.ColorFormBackground)
	textView.SetBorderPadding(0, 0, 2, 2)

	// Create help text
	helpText := tview.NewTextView()
	helpText.SetText("[yellow]Tip: Press SPACE to toggle the checkbox[white]")
	helpText.SetDynamicColors(true)
	helpText.SetTextAlign(tview.AlignCenter)
	helpText.SetBackgroundColor(tui.ColorFormBackground)

	// Create a grid to control form alignment
	formGrid := tview.NewGrid().
		SetColumns(-1, -6, -1).
		SetRows(0)
	formGrid.SetBackgroundColor(tui.ColorFormBackground)
	formGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 0, 0, 1, 1, 0, 0, false)
	formGrid.AddItem(form, 0, 1, 1, 1, 0, 0, true)
	formGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 0, 2, 1, 1, 0, 0, false)

	// Create a flex container to stack text and form vertically.
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textView, 4, 0, false).
		AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 1, 0, false). // Spacer between text and form
		AddItem(formGrid, 3, 0, true).                                                    // Form grid with checkbox and button
		AddItem(helpText, 1, 0, false).                                                   // Help text at bottom
		AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 1, 0, false)  // Bottom spacer

	// Create content grid.
	contentGrid := tview.NewGrid()
	contentGrid.SetRows(1, 0, 0)
	contentGrid.SetBackgroundColor(tui.ColorFormBackground)

	// Add items to content grid.
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 0, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(flex, 1, 0, 2, 1, 0, 0, true)
	contentGrid.SetBorder(true)
	contentGrid.SetTitle(" 📊 Attestation Data Contribution ")
	contentGrid.SetBackgroundColor(tui.ColorFormBackground)

	// Create border grid.
	borderGrid := tview.NewGrid()
	borderGrid.SetColumns(0, modalWidth, 0)
	borderGrid.SetRows(0, height, 0, 2)
	borderGrid.SetBackgroundColor(tui.ColorFormBackground)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	p.content = borderGrid
}

func (p *AttestationOptInPage) saveAttestationPreference() {
	// Update config with attestation preference
	if err := p.display.sidecarCfg.Update(func(cfg *config.Config) {
		if p.enabled {
			// Only create the AttestationSubnetCheck if user opts in
			if cfg.AttestationSubnetCheck == nil {
				cfg.AttestationSubnetCheck = &config.AttestationSubnetCheck{}
			}
			cfg.AttestationSubnetCheck.Enabled = true
		} else {
			// Remove the field entirely for opt-out
			cfg.AttestationSubnetCheck = nil
		}
	}); err != nil {
		p.openErrorModal(err)

		return
	}

	// Navigate to the finished page
	p.display.app.SetFocus(p.form.GetButton(0))
	p.display.setPage(p.display.finishedPage.GetPage())
}

func (p *AttestationOptInPage) openErrorModal(err error) {
	p.display.app.SetRoot(tui.CreateErrorModal(
		p.display.app,
		err.Error(),
		func() {
			p.display.app.SetRoot(p.display.frame, true).EnableMouse(true)
		},
	), true).EnableMouse(true)
}
