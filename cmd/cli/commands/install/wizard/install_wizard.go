package wizard

import (
	"github.com/ethpandaops/contributoor-installer/cmd/cli/terminal"
	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

// InstallWizard is the wizard for the install command.
type InstallWizard struct {
	*display.BaseWizard
	configService *service.ConfigService
	completed     bool
}

// NewInstallWizard creates a new install wizard.
func NewInstallWizard(log *logrus.Logger, app *tview.Application, configService *service.ConfigService) *InstallWizard {
	w := &InstallWizard{
		BaseWizard:    display.NewBaseWizard(log, app),
		configService: configService,
		completed:     false,
	}

	w.Steps = []display.WizardStep{
		NewWelcomeStep(w),
		NewNetworkStep(w),
		NewBeaconNodeStep(w),
		NewOutputServerStep(w),
		NewOutputCredentialsStep(w),
		NewFinishStep(w),
	}

	w.CurrentStep = w.Steps[0]
	app.SetRoot(w.GetPages(), true)

	return w
}

// GetConfig returns the current configuraiton to the wizard.
func (w *InstallWizard) GetConfig() *service.ContributoorConfig {
	return w.configService.Get()
}

// UpdateConfig updates the current configuraiton via the wizard.
func (w *InstallWizard) UpdateConfig(updates func(*service.ContributoorConfig)) error {
	return w.configService.Update(updates)
}

// Start opens the first step of the wizard.
func (w *InstallWizard) Start() error {
	return w.CurrentStep.Show()
}

// SetCompleted sets the wizard as "completed"/"finished".
func (w *InstallWizard) SetCompleted() {
	w.completed = true
}

// OnComplete is called when the user has finished the installation wizard.
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
