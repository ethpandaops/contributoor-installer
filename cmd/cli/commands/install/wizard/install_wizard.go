package wizard

import (
	"fmt"
	"path/filepath"

	config "github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal/display"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal/service"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

type InstallWizard struct {
	*display.BaseWizard
	Config    *config.ContributoorConfig
	completed bool
}

func NewInstallWizard(log *logrus.Logger, app *tview.Application, cfg *config.ContributoorConfig) *InstallWizard {
	w := &InstallWizard{
		BaseWizard: display.NewBaseWizard(log, app),
		Config:     cfg,
		completed:  false,
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

	// Save config before starting services.
	configPath := filepath.Join(w.Config.ContributoorDirectory, "contributoor.yaml")
	if err := w.Config.WriteToFile(configPath); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	switch w.Config.RunMethod {
	case config.RunMethodDocker:
		dockerService := service.NewDockerService(w.Logger, w.Config)
		if err := dockerService.Start(); err != nil {
			w.Logger.Errorf("could not start service: %v", err)
			return err
		}
	case config.RunMethodBinary:
		binaryService := service.NewBinaryService(w.Logger, w.Config)
		if err := binaryService.Start(); err != nil {
			w.Logger.Errorf("could not start service: %v", err)
			return err
		}
	}

	return nil
}
