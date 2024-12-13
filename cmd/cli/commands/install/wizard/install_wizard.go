package wizard

import (
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal/display"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal/service"
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

	switch w.configService.Get().RunMethod {
	case service.RunMethodDocker:
		dockerService, err := service.NewDockerService(w.Logger, w.configService)
		if err != nil {
			w.Logger.Errorf("could not create docker service: %v", err)
			return err
		}

		if err := dockerService.Start(); err != nil {
			w.Logger.Errorf("could not start service: %v", err)
			return err
		}
	case service.RunMethodBinary:
		binaryService := service.NewBinaryService(w.Logger, w.configService)
		if err := binaryService.Start(); err != nil {
			w.Logger.Errorf("could not start service: %v", err)
			return err
		}
	}

	return nil
}
