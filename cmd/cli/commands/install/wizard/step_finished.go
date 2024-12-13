package wizard

import (
	"fmt"

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
		step:   3,
		total:  3,
	}

	cfg := w.GetConfig()
	helperText := fmt.Sprintf(`Nice work!

Your configuration:
Network: %s
Beacon Node: %s
Run Method: %s
Version: %s

You're all done and ready to run contributoor.`,
		cfg.Network.Name,
		cfg.Network.BeaconNodeAddress,
		cfg.RunMethod,
		cfg.Version)

	step.modal = tview.NewModal().
		SetText(helperText).
		AddButtons([]string{"Save and Exit"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			step.wizard.SetCompleted()
			step.wizard.GetApp().Stop()
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
	return s.wizard.Steps[1], nil
}

func (s *FinishStep) GetTitle() string {
	return "Finished"
}

func (s *FinishStep) GetProgress() (int, int) {
	return s.step, s.total
}
