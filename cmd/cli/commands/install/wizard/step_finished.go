package wizard

import (
	"github.com/ethpandaops/contributoor-installer-test/internal/display"
	"github.com/rivo/tview"
)

// FinishStep is the last step of the installation wizard.
type FinishStep struct {
	wizard      *InstallWizard
	modal       *tview.Modal
	step, total int
}

// NewFinishStep creates a new finish step.
func NewFinishStep(w *InstallWizard) *FinishStep {
	step := &FinishStep{
		wizard: w,
		step:   3,
		total:  3,
	}

	helperText := `Nice work!
You're all done and ready to run contributoor.`

	step.modal = tview.NewModal().
		SetText(helperText).
		AddButtons([]string{"Save and Exit"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			step.wizard.SetCompleted()
			step.wizard.GetApp().Stop()
		})

	return step
}

// Show displays the finish step.
func (s *FinishStep) Show() error {
	s.wizard.GetApp().SetRoot(s.modal, true)

	return nil
}

// Next returns the next step.
func (s *FinishStep) Next() (display.WizardStep, error) {
	return nil, nil //nolint:nilnil // No next step.
}

// Previous returns the previous step.
func (s *FinishStep) Previous() (display.WizardStep, error) {
	return s.wizard.Steps[1], nil
}
