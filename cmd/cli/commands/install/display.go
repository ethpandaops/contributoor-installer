package install

import (
	"github.com/ethpandaops/contributoor-installer/cmd/cli/terminal"
	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

type installPage interface {
	GetPage() *display.Page
}

type InstallDisplay struct {
	app                         *tview.Application
	pages                       *tview.Pages
	frame                       *tview.Frame
	log                         *logrus.Logger
	configService               *service.ConfigService
	homePage                    *page
	content                     tview.Primitive
	installPages                []installPage
	networkPage                 *NetworkPage
	beaconPage                  *BeaconNodePage
	outputPage                  *OutputServerPage
	description                 *tview.TextView
	welcomePage                 *WelcomePage
	outputServerCredentialsPage *OutputServerCredentialsPage
	finishedPage                *FinishedPage
}

func NewInstallDisplay(log *logrus.Logger, app *tview.Application, configService *service.ConfigService) *InstallDisplay {
	installDisplay := &InstallDisplay{
		app:           app,
		pages:         tview.NewPages(),
		log:           log,
		configService: configService,
	}

	// Create pages
	installDisplay.welcomePage = NewWelcomePage(installDisplay)
	installDisplay.networkPage = NewNetworkPage(installDisplay)
	installDisplay.beaconPage = NewBeaconNodePage(installDisplay)
	installDisplay.outputPage = NewOutputServerPage(installDisplay)
	installDisplay.outputServerCredentialsPage = NewOutputServerCredentialsPage(installDisplay)
	installDisplay.finishedPage = NewFinishedPage(installDisplay)
	installDisplay.installPages = []installPage{
		installDisplay.welcomePage,
		installDisplay.networkPage,
		installDisplay.beaconPage,
		installDisplay.outputPage,
		installDisplay.outputServerCredentialsPage,
		installDisplay.finishedPage,
	}

	// Setup pages
	for _, page := range installDisplay.installPages {
		installDisplay.pages.AddPage(page.GetPage().ID, page.GetPage().Content, true, false)
	}

	// Create initial frame
	frame := display.CreatePageFrame(display.PageFrameOptions{
		Content:  installDisplay.pages,
		Title:    installDisplay.welcomePage.GetPage().Title,
		HelpType: display.HelpWizard,
		Step:     1,
		Total:    len(installDisplay.installPages),
	})
	installDisplay.frame = frame
	installDisplay.app.SetRoot(frame, true)

	return installDisplay
}

func (d *InstallDisplay) Run() error {
	d.setPage(d.welcomePage.GetPage())

	d.log.WithFields(logrus.Fields{
		"config_path": d.configService.Get().ContributoorDirectory,
		"version":     d.configService.Get().Version,
		"run_method":  d.configService.Get().RunMethod,
	}).Info("Running installation wizard")

	return d.app.Run()
}

func (d *InstallDisplay) getCurrentStep() int {
	// Map pages to step numbers
	stepMap := map[string]int{
		"install-welcome":     1,
		"install-network":     2,
		"install-beacon":      3,
		"install-output":      4,
		"install-credentials": 5,
	}

	currentPage, _ := d.pages.GetFrontPage()
	if step, exists := stepMap[currentPage]; exists {
		return step
	}
	return 1
}

func (d *InstallDisplay) setPage(page *page) {
	// Switch to the new page first
	d.pages.SwitchToPage(page.ID)

	// Then create the frame with the updated step number
	d.frame.Clear()
	frame := display.CreatePageFrame(display.PageFrameOptions{
		Content:  d.pages,
		Title:    page.Title,
		HelpType: display.HelpWizard,
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

func (d *InstallDisplay) OnComplete() error {
	d.log.Infof("%sInstallation complete%s", terminal.ColorGreen, terminal.ColorReset)
	d.log.Info("You can now manage contributoor using the following commands:")
	d.log.Info("    contributoor start")
	d.log.Info("    contributoor stop")
	d.log.Info("    contributoor update")
	d.log.Info("    contributoor config")

	return nil
}
