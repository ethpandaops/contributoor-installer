package wizard

import (
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/terminal"
	"github.com/ethpandaops/contributoor-installer-test/internal/display"
	"github.com/ethpandaops/contributoor-installer-test/internal/service"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

type InstallWizard struct {
	*display.BaseWizard
	configService *service.ConfigService
	completed     bool
}

func NewInstallWizard(log *logrus.Logger, app *tview.Application, configService *service.ConfigService) *InstallWizard {
	w := &InstallWizard{
		BaseWizard:    display.NewBaseWizard(log, app),
		configService: configService,
		completed:     false,
	}

	// Add install-specific steps
	w.Steps = []display.WizardStep{
		NewWelcomeStep(w),
		NewNetworkStep(w),
		NewFinishStep(w),
	}
	w.CurrentStep = w.Steps[0]

	// Set Pages as root
	app.SetRoot(w.GetPages(), true)

	return w
}

func (w *InstallWizard) GetConfig() *service.ContributoorConfig {
	return w.configService.Get()
}

func (w *InstallWizard) UpdateConfig(updates func(*service.ContributoorConfig)) error {
	return w.configService.Update(updates)
}

func (w *InstallWizard) Start() error {
	return w.CurrentStep.Show()
}

func (w *InstallWizard) SetCompleted() {
	w.completed = true
}

func (w *InstallWizard) OnComplete() error {
	// Don't save config if installation was interrupted
	if !w.completed {
		w.Logger.Info("Installation was interrupted")
		return nil
	}

	w.GetApp().Stop()

	w.Logger.Infof("%sInstallation complete. You can now run 'contributoor start'.%s", terminal.ColorGreen, terminal.ColorReset)

	return nil
}
