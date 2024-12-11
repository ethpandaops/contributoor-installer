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
	freshInstall bool
	Config       *config.ContributoorConfig
}

func NewInstallWizard(log *logrus.Logger, app *tview.Application, cfg *config.ContributoorConfig, freshInstall bool) *InstallWizard {
	w := &InstallWizard{
		BaseWizard:   display.NewBaseWizard(log, app),
		Config:       cfg,
		freshInstall: freshInstall,
	}

	// Add install-specific steps
	w.Steps = []display.WizardStep{
		NewWelcomeStep(w),
		NewModeStep(w),
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

func (w *InstallWizard) OnComplete() error {
	w.GetApp().Stop()

	// Save config before starting services.
	configPath := filepath.Join(w.Config.ContributoorDirectory, "contributoor.yaml")
	if err := w.Config.WriteToFile(configPath); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	w.Logger.Info("Installation wizard complete")

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
