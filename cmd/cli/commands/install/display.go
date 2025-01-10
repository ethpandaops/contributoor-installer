package install

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

// InstallDisplay is the display for the install wizard.
type InstallDisplay struct {
	app               *tview.Application
	pages             *tview.Pages
	frame             *tview.Frame
	log               *logrus.Logger
	sidecarCfg        sidecar.ConfigManager
	installPages      []tui.PageInterface
	welcomePage       *WelcomePage
	networkConfigPage *NetworkConfigPage
	beaconPage        *BeaconNodePage
	//nolint:unused // Disabled for now.
	outputPage                  *OutputServerPage
	outputServerCredentialsPage *OutputServerCredentialsPage
	finishedPage                *FinishedPage
}

// NewInstallDisplay creates a new InstallDisplay.
func NewInstallDisplay(log *logrus.Logger, app *tview.Application, sidecarCfg sidecar.ConfigManager) *InstallDisplay {
	display := &InstallDisplay{
		app:        app,
		pages:      tview.NewPages(),
		log:        log,
		sidecarCfg: sidecarCfg,
	}

	// Create all of our install wizard pages.
	display.welcomePage = NewWelcomePage(display)
	display.networkConfigPage = NewNetworkConfigPage(display)
	display.beaconPage = NewBeaconNodePage(display)
	display.outputServerCredentialsPage = NewOutputServerCredentialsPage(display)
	display.finishedPage = NewFinishedPage(display)
	display.installPages = []tui.PageInterface{
		display.welcomePage,
		display.networkConfigPage,
		display.beaconPage,
		display.outputServerCredentialsPage,
		display.finishedPage,
	}

	// Add all of our pages to the pages stack.
	for _, page := range display.installPages {
		display.pages.AddPage(page.GetPage().ID, page.GetPage().Content, true, false)
	}

	// Create the initial page frame, this houses breadcrumbs, page stack, etc.
	frame := tui.CreatePageFrame(tui.PageFrameOptions{
		Content:  display.pages,
		Title:    display.welcomePage.GetPage().Title,
		HelpType: tui.HelpWizard,
		Step:     1,
		Total:    len(display.installPages),
	})
	display.frame = frame
	display.app.SetRoot(frame, true)

	return display
}

// Run starts the install wizard.
func (d *InstallDisplay) Run() error {
	d.setPage(d.welcomePage.GetPage())

	cfg := d.sidecarCfg.Get()

	d.log.WithFields(logrus.Fields{
		"config_path": cfg.ContributoorDirectory,
		"version":     cfg.Version,
		"run_method":  cfg.RunMethod,
	}).Debug("Running installation wizard")

	return d.app.Run()
}

// getCurrentStep returns the current step number based on the current page.
func (d *InstallDisplay) getCurrentStep() int {
	// Map pages to step numbers
	stepMap := map[string]int{
		"install-welcome":     1,
		"install-network":     2,
		"install-beacon":      3,
		"install-credentials": 4,
	}

	currentPage, _ := d.pages.GetFrontPage()
	if step, exists := stepMap[currentPage]; exists {
		return step
	}

	return 1
}

// setPage switches to the new page and updates the frame.
func (d *InstallDisplay) setPage(page *tui.Page) {
	// Switch to the new page first.
	d.pages.SwitchToPage(page.ID)

	// Then create the frame with the updated step number.
	d.frame.Clear()
	frame := tui.CreatePageFrame(tui.PageFrameOptions{
		Content:  d.pages,
		Title:    page.Title,
		HelpType: tui.HelpWizard,
		Step:     d.getCurrentStep(),
		Total:    len(d.installPages),
		OnEsc: func() {
			if page.Parent != nil {
				d.setPage(page.Parent)
			}
		},
	})

	d.frame = frame
	d.app.SetRoot(frame, true)
}

// OnComplete is called when the install wizard is complete.
func (d *InstallDisplay) OnComplete() error {
	cfg := d.sidecarCfg.Get()

	fmt.Printf("%sContributoor Status%s\n", tui.TerminalColorLightBlue, tui.TerminalColorReset)
	fmt.Printf("%-20s: %s\n", "Version", cfg.Version)
	fmt.Printf("%-20s: %s\n", "Run Method", cfg.RunMethod)
	fmt.Printf("%-20s: %s\n", "Network", cfg.NetworkName)
	fmt.Printf("%-20s: %s\n", "Beacon Node", cfg.BeaconNodeAddress)
	fmt.Printf("%-20s: %s\n", "Config Path", d.sidecarCfg.GetConfigPath())

	if cfg.OutputServer != nil {
		fmt.Printf("%-20s: %s\n", "Output Server", cfg.OutputServer.Address)
	}

	fmt.Printf("\n%sInstallation complete%s\n", tui.TerminalColorGreen, tui.TerminalColorReset)
	fmt.Printf("You can now manage contributoor using the following command(s):\n")
	fmt.Printf("    contributoor [start|stop|status|update|config]\n")

	return nil
}
