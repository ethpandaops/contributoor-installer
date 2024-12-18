package install

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type OutputServerCredentialsPage struct {
	display     *InstallDisplay
	page        *page
	content     tview.Primitive
	form        *tview.Form
	description *tview.TextView
	password    string
	username    string
}

func NewOutputServerCredentialsPage(display *InstallDisplay) *OutputServerCredentialsPage {
	credentialsPage := &OutputServerCredentialsPage{
		display: display,
	}

	credentialsPage.initPage()
	credentialsPage.page = newPage(
		display.outputPage.GetPage(), // Set parent to output server page
		"install-credentials",
		"Output Server Credentials",
		"Configure your output server authentication",
		credentialsPage.content,
	)

	return credentialsPage
}

func (p *OutputServerCredentialsPage) GetPage() *display.Page {
	return p.page
}

func (p *OutputServerCredentialsPage) initPage() {
	// Layout components
	var (
		modalWidth = 70 // Match other pages' width
		lines      = tview.WordWrap("Please enter your output server credentials", modalWidth-4)
		height     = len(lines) + 4
	)

	// Initialize form
	form := tview.NewForm()
	form.SetButtonsAlign(tview.AlignCenter)
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetBackgroundColor(display.ColorFormBackground)
	form.SetBorderPadding(0, 0, 0, 0)
	form.SetLabelColor(tcell.ColorLightGray)
	p.form = form

	// Get existing credentials if any
	if currentCreds := p.display.configService.Get().OutputServer.Credentials; currentCreds != "" {
		if decoded, err := base64.StdEncoding.DecodeString(currentCreds); err == nil {
			parts := strings.Split(string(decoded), ":")
			if len(parts) == 2 {
				p.username = parts[0]
				p.password = parts[1]
			}
		}
	}

	// Add input fields with existing values
	form.AddInputField("Username", p.username, 40, nil, func(username string) {
		p.username = username
	})

	form.AddPasswordField("Password", p.password, 40, '*', func(password string) {
		p.password = password
	})

	// Add Next button
	form.AddButton(display.ButtonNext, func() {
		// Only validate credentials for ethpandaops servers
		currentAddress := p.display.configService.Get().OutputServer.Address
		if strings.Contains(currentAddress, "platform.ethpandaops.io") {
			// Validate credentials
			if p.username == "" || p.password == "" {
				errorModal := display.CreateErrorModal(
					p.display.app,
					"Username and password are required for ethpandaops servers",
					func() {
						p.display.app.SetRoot(p.display.frame, true)
						p.display.app.SetFocus(form)
					},
				)
				p.display.app.SetRoot(errorModal, true)
				return
			}
		}

		if p.username != "" && p.password != "" {
			// Set credentials only when validated
			credentials := fmt.Sprintf("%s:%s", p.username, p.password)
			p.display.configService.Update(func(cfg *service.ContributoorConfig) {
				cfg.OutputServer.Credentials = base64.StdEncoding.EncodeToString([]byte(credentials))
			})
		}

		p.display.setPage(p.display.finishedPage.GetPage())
	})

	// Style the button
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

	// Create text view
	textView := tview.NewTextView()
	textView.SetText("Please enter your output server credentials")
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
	contentGrid.SetTitle(" Output Server Credentials ")
	contentGrid.SetBackgroundColor(display.ColorFormBackground)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	p.content = borderGrid
}

func validateAndSaveCredentials(p *OutputServerCredentialsPage) {
	username := p.form.GetFormItem(0).(*tview.InputField).GetText()
	password := p.form.GetFormItem(1).(*tview.InputField).GetText()

	// Only require credentials for non-custom servers
	if p.display.configService.Get().OutputServer.Address != "custom" {
		if username == "" || password == "" {
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
	}

	// Update config with credentials
	p.display.configService.Update(func(cfg *service.ContributoorConfig) {
		if username != "" && password != "" {
			credentials := fmt.Sprintf("%s:%s", username, password)
			cfg.OutputServer.Credentials = base64.StdEncoding.EncodeToString([]byte(credentials))
		} else {
			cfg.OutputServer.Credentials = ""
		}
	})

	// Installation complete
	p.display.app.Stop()
}

func (p *OutputServerCredentialsPage) updateDescription(text string) {
	p.description.SetText(text)
}
