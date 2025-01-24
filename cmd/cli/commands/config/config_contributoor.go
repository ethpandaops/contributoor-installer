package config

import (
	"runtime"
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

// Available run modes.
var runModes = []config.RunMethod{
	config.RunMethod_RUN_METHOD_DOCKER,
	config.RunMethod_RUN_METHOD_SYSTEMD,
	config.RunMethod_RUN_METHOD_BINARY,
}

// ContributoorSettingsPage is a page that allows the user to configure core contributoor settings.
type ContributoorSettingsPage struct {
	display     *ConfigDisplay
	page        *tui.Page
	content     tview.Primitive
	form        *tview.Form
	description *tview.TextView
}

// NewContributoorSettingsPage creates a new ContributoorSettingsPage.
func NewContributoorSettingsPage(cd *ConfigDisplay) *ContributoorSettingsPage {
	settingsPage := &ContributoorSettingsPage{
		display: cd,
	}

	settingsPage.initPage()
	settingsPage.page = tui.NewPage(
		cd.homePage,
		"config-contributoor",
		"Contributoor Settings",
		"Configure core contributoor settings like logging and run mode",
		settingsPage.content,
	)

	return settingsPage
}

// GetPage returns the page.
func (p *ContributoorSettingsPage) GetPage() *tui.Page {
	return p.page
}

// initPage initializes the page.
func (p *ContributoorSettingsPage) initPage() {
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

	// Available log levels
	logLevels := []string{
		logrus.TraceLevel.String(),
		logrus.DebugLevel.String(),
		logrus.InfoLevel.String(),
		logrus.WarnLevel.String(),
		logrus.ErrorLevel.String(),
	}

	// Find current log level index
	currentLogLevel := p.display.sidecarCfg.Get().LogLevel
	currentLogLevelIndex := 2 // Default to info

	for i, level := range logLevels {
		if level == currentLogLevel {
			currentLogLevelIndex = i

			break
		}
	}

	// Find current run mode index
	currentRunMode := p.display.sidecarCfg.Get().RunMethod
	currentRunModeIndex := 0 // Default to docker

	for i, mode := range runModes {
		if mode == currentRunMode {
			currentRunModeIndex = i

			break
		}
	}

	// Create display labels that show launchd on macOS
	runModeLabels := make([]string, len(runModes))

	for i, mode := range runModes {
		if mode == config.RunMethod_RUN_METHOD_SYSTEMD {
			runModeLabels[i] = getServiceManagerLabel()
		} else {
			runModeLabels[i] = strings.ToLower(mode.DisplayName())
		}
	}

	// Add our form fields.
	form.AddDropDown("Log Level", logLevels, currentLogLevelIndex, func(text string, index int) {
		p.description.SetText("Set the logging verbosity level. Debug and Trace provide more detailed output.")
	})

	form.AddDropDown("Run Mode", runModeLabels, currentRunModeIndex, func(text string, index int) {
		if text == strings.ToLower(config.RunMethod_RUN_METHOD_DOCKER.DisplayName()) {
			p.description.SetText("Run using Docker containers (recommended)")
		} else if text == getServiceManagerLabel() {
			p.description.SetText(getServiceManagerDescription())
		} else {
			p.description.SetText("Run directly as a binary on your system")
		}
	})

	// Add a save button and ensure we validate the input.
	saveButton := tview.NewButton(tui.ButtonSaveSettings)
	saveButton.SetSelectedFunc(func() {
		validateAndUpdateContributoor(p)
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
	formFrame.SetTitle("Contributoor Settings")
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

	// Create a main layout both the flexes.
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(formDescriptionFlex, 0, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(buttonFlex, 1, 0, false).
		AddItem(nil, 1, 0, false)
	mainFlex.SetBackgroundColor(tui.ColorBackground)

	p.content = mainFlex
}

func validateAndUpdateContributoor(p *ContributoorSettingsPage) {
	logLevel, _ := p.form.GetFormItem(0).(*tview.DropDown)
	runMode, _ := p.form.GetFormItem(1).(*tview.DropDown)

	_, logLevelText := logLevel.GetCurrentOption()
	runModeIndex, _ := runMode.GetCurrentOption()

	if err := p.display.sidecarCfg.Update(func(cfg *config.Config) {
		cfg.LogLevel = logLevelText
		cfg.RunMethod = runModes[runModeIndex]
	}); err != nil {
		p.openErrorModal(err)

		return
	}

	p.display.markConfigChanged()
	p.display.setPage(p.display.homePage)
}

func (p *ContributoorSettingsPage) openErrorModal(err error) {
	p.display.app.SetRoot(tui.CreateErrorModal(
		p.display.app,
		err.Error(),
		func() {
			p.display.app.SetRoot(p.display.frame, true).EnableMouse(true)
			p.display.app.SetFocus(p.form)
		},
	), true).EnableMouse(true)
}

// getServiceManagerLabel returns the appropriate label based on platform.
func getServiceManagerLabel() string {
	if runtime.GOOS == "darwin" {
		return "launchd"
	}

	return "systemd"
}

// getServiceManagerDescription returns the appropriate description based on platform.
func getServiceManagerDescription() string {
	if runtime.GOOS == "darwin" {
		return "Run using macOS launchd service manager"
	}

	return "Run using Linux systemd service manager"
}
