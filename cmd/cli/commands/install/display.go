package install

import (
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

type InstallDisplay struct {
	app                         *tview.Application
	pages                       *tview.Pages
	frame                       *tview.Frame
	log                         *logrus.Logger
	configService               *service.ConfigService
	homePage                    *tui.Page
	content                     tview.Primitive
	installPages                []tui.PageInterface
	networkConfigPage           *networkConfigPage
	beaconPage                  *BeaconNodePage
	outputPage                  *OutputServerPage
	description                 *tview.TextView
	welcomePage                 *WelcomePage
	outputServerCredentialsPage *OutputServerCredentialsPage
	finishedPage                *FinishedPage
}

func NewInstallDisplay(log *logrus.Logger, app *tview.Application, configService *service.ConfigService) *InstallDisplay {
	id := &InstallDisplay{
		app:           app,
		pages:         tview.NewPages(),
		log:           log,
		configService: configService,
	}

	// Create pages
	id.welcomePage = NewWelcomePage(id)
	id.networkConfigPage = NewnetworkConfigPage(id)
	id.beaconPage = NewBeaconNodePage(id)
	id.outputPage = NewOutputServerPage(id)
	id.outputServerCredentialsPage = NewOutputServerCredentialsPage(id)
	id.finishedPage = NewFinishedPage(id)
	id.installPages = []tui.PageInterface{
		id.welcomePage,
		id.networkConfigPage,
		id.beaconPage,
		id.outputPage,
		id.outputServerCredentialsPage,
		id.finishedPage,
	}

	// Setup pages
	for _, page := range id.installPages {
		id.pages.AddPage(page.GetPage().ID, page.GetPage().Content, true, false)
	}

	// Create initial frame
	frame := tui.CreatePageFrame(tui.PageFrameOptions{
		Content:  id.pages,
		Title:    id.welcomePage.GetPage().Title,
		HelpType: tui.HelpWizard,
		Step:     1,
		Total:    len(id.installPages),
	})
	id.frame = frame
	id.app.SetRoot(frame, true)

	return id
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

func (d *InstallDisplay) setPage(page *tui.Page) {
	// Switch to the new page first
	d.pages.SwitchToPage(page.ID)

	// Then create the frame with the updated step number
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

func (d *InstallDisplay) OnComplete() error {
	d.log.Infof("%sInstallation complete%s", tui.TerminalColorGreen, tui.TerminalColorReset)
	d.log.Info("You can now manage contributoor using the following commands:")
	d.log.Info("    contributoor start")
	d.log.Info("    contributoor stop")
	d.log.Info("    contributoor update")
	d.log.Info("    contributoor config")

	return nil
}
