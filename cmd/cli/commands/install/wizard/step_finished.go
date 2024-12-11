package wizard

import (
	"fmt"
	"path/filepath"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal/display"
	"github.com/rivo/tview"
)

type FinishStep struct {
	wizard      *InstallWizard
	modal       *tview.Modal
	step, total int
}

func NewFinishStep(w *InstallWizard) *FinishStep {
	step := &FinishStep{
		wizard: w,
		step:   4,
		total:  4,
	}

	helperText := "All done! You're ready to run."

	step.modal = tview.NewModal().
		SetText(helperText).
		AddButtons([]string{"Save and Exit"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 0 { // Save and Exit
				// Process config first
				if err := step.processConfigAfterQuit(); err != nil {
					step.wizard.Logger.Error("failed to process config after quit: %w", err)
					return
				}

				step.wizard.GetApp().Stop()
			}
		})

	return step
}

func (s *FinishStep) Show() error {
	s.wizard.GetApp().SetRoot(s.modal, true)
	return nil
}

func (s *FinishStep) Next() (display.WizardStep, error) {
	return nil, nil // Last step
}

func (s *FinishStep) Previous() (display.WizardStep, error) {
	return s.wizard.Steps[2], nil
}

func (s *FinishStep) GetTitle() string {
	return "Finished"
}

func (s *FinishStep) GetProgress() (int, int) {
	return s.step, s.total
}

func (s *FinishStep) processConfigAfterQuit() error {
	// Write config to file
	configPath := filepath.Join(s.wizard.Config.ContributoorDirectory, "contributoor.yaml")
	if err := s.wizard.Config.WriteToFile(configPath); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
