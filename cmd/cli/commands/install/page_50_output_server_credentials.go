package install

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor-installer/internal/validate"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// OutputServerCredentialsPage is the page for configuring the users output server credentials.
type OutputServerCredentialsPage struct {
	display  *InstallDisplay
	page     *tui.Page
	content  tview.Primitive
	form     *tview.Form
	password string
	username string
}

// NewOutputServerCredentialsPage creates a new OutputServerCredentialsPage.
func NewOutputServerCredentialsPage(display *InstallDisplay) *OutputServerCredentialsPage {
	credentialsPage := &OutputServerCredentialsPage{
		display: display,
	}

	credentialsPage.initPage()
	credentialsPage.page = tui.NewPage(
		display.beaconPage.GetPage(),
		"install-credentials",
		"Output Server Credentials",
		"Configure your output server authentication",
		credentialsPage.content,
	)

	return credentialsPage
}

// GetPage returns the page.
func (p *OutputServerCredentialsPage) GetPage() *tui.Page {
	return p.page
}

// initPage initializes the page.
func (p *OutputServerCredentialsPage) initPage() {
	var (
		modalWidth = 70
		lines      = tview.WordWrap("Please enter your output server credentials", modalWidth-4)
		height     = len(lines) + 4
	)

	// We need a form to house our input fields.
	form := tview.NewForm()
	form.SetButtonsAlign(tview.AlignCenter)
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetBackgroundColor(tui.ColorFormBackground)
	form.SetBorderPadding(0, 0, 0, 0)
	form.SetLabelColor(tcell.ColorLightGray)
	p.form = form

	// Get existing credentials if any
	if cfg := p.display.sidecarCfg.Get(); cfg.OutputServer != nil && cfg.OutputServer.Credentials != "" {
		currentCreds := cfg.OutputServer.Credentials
		if decoded, err := base64.StdEncoding.DecodeString(currentCreds); err == nil {
			parts := strings.Split(string(decoded), ":")
			if len(parts) == 2 {
				p.username = parts[0]
				p.password = parts[1]
			}
		}
	}

	// Add input fields with existing values.
	form.AddInputField("Username", p.username, 40, nil, func(username string) {
		p.username = username
	})

	form.AddPasswordField("Password", p.password, 40, '*', func(password string) {
		p.password = password
	})

	// Add a 'Next' button.
	form.AddButton(tui.ButtonNext, func() {
		validateAndSaveCredentials(p)
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
	contentGrid.SetColumns(-1, -6, -1)
	contentGrid.SetBackgroundColor(tui.ColorFormBackground)

	// Create text view.
	textView := tview.NewTextView()
	textView.SetText("Please enter your output server credentials\nThese would have been provided to you by the ethPandaOps team")
	textView.SetTextAlign(tview.AlignCenter)
	textView.SetWordWrap(true)
	textView.SetTextColor(tview.Styles.PrimaryTextColor)
	textView.SetBackgroundColor(tui.ColorFormBackground)
	textView.SetBorderPadding(0, 0, 0, 0)

	// Add items to content grid.
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 0, 0, 1, 3, 0, 0, false)
	contentGrid.AddItem(textView, 1, 0, 1, 3, 0, 0, false)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 2, 0, 1, 3, 0, 0, false)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 3, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(form, 3, 1, 1, 1, 0, 0, true)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 3, 2, 1, 1, 0, 0, false)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 5, 0, 1, 3, 0, 0, false)
	contentGrid.SetBorder(true)
	contentGrid.SetTitle(" 🔐 Output Server Credentials ")
	contentGrid.SetBackgroundColor(tui.ColorFormBackground)

	// Create border grid.
	borderGrid := tview.NewGrid()
	borderGrid.SetColumns(0, modalWidth, 0)
	borderGrid.SetRows(0, height+9, 0, 2)
	borderGrid.SetBackgroundColor(tui.ColorFormBackground)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	p.content = borderGrid
}

func validateAndSaveCredentials(p *OutputServerCredentialsPage) {
	var username, password string

	if item := p.form.GetFormItem(0); item != nil {
		if inputField, ok := item.(*tview.InputField); ok {
			username = inputField.GetText()
		} else {
			p.openErrorModal(fmt.Errorf("invalid username field type"))

			return
		}
	}

	if item := p.form.GetFormItem(1); item != nil {
		if inputField, ok := item.(*tview.InputField); ok {
			password = inputField.GetText()
		} else {
			p.openErrorModal(fmt.Errorf("invalid password field type"))

			return
		}
	}

	currentAddress := p.display.sidecarCfg.Get().OutputServer.Address
	isEthPandaOps := validate.IsEthPandaOpsServer(currentAddress)

	if err := validate.ValidateOutputServerCredentials(username, password, isEthPandaOps); err != nil {
		p.openErrorModal(err)

		return
	}

	// Update config with credentials
	if err := p.display.sidecarCfg.Update(func(cfg *config.Config) {
		// For custom servers, allow empty credentials
		// For ethPandaOps servers, we know credentials are valid (non-empty) due to validation.
		if username != "" && password != "" {
			cfg.OutputServer.Credentials = validate.EncodeCredentials(username, password)
		} else if !isEthPandaOps {
			// Only clear credentials if it's a custom server.
			cfg.OutputServer.Credentials = ""
		}
	}); err != nil {
		p.openErrorModal(err)

		return
	}

	// Set initial focus on the Next button.
	p.display.app.SetFocus(p.form.GetButton(0))
	p.display.setPage(p.display.finishedPage.GetPage())
}

func (p *OutputServerCredentialsPage) openErrorModal(err error) {
	p.display.app.SetRoot(tui.CreateErrorModal(
		p.display.app,
		err.Error(),
		func() {
			p.display.app.SetRoot(p.display.frame, true).EnableMouse(true)
		},
	), true).EnableMouse(true)
}
