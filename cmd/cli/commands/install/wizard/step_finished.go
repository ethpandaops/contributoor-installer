package wizard

import (
	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FinishStep is the last step of the installation wizard.
type FinishStep struct {
	Wizard      *InstallWizard
	Modal       *tview.Frame
	Step, Total int
}

// NewFinishStep creates a new finish step.
func NewFinishStep(w *InstallWizard) *FinishStep {
	step := &FinishStep{
		Wizard: w,
		Step:   3,
		Total:  3,
	}

	helperText := `Nice work!
You're all done and ready to run contributoor.`

	// Create the modal
	modal := tview.NewModal().
		SetText(helperText).
		AddButtons([]string{"Save and Exit"}).
		SetButtonStyle(tcell.StyleDefault.
			Background(tcell.ColorDefault).
			Foreground(tcell.ColorLightGray)).
		SetButtonActivatedStyle(tcell.StyleDefault.
			Background(tcell.Color46).
			Foreground(tcell.ColorBlack)).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			step.Wizard.SetCompleted()
			step.Wizard.GetApp().Stop()
		})

	step.Modal = display.CreateWizardFrame(display.WizardFrameOptions{
		Content: modal,
		Step:    step.Step,
		Total:   step.Total,
		Title:   "Finished",
		OnEsc: func() {
			if prev, err := step.Previous(); err == nil {
				step.Wizard.CurrentStep = prev
				if err := prev.Show(); err != nil {
					step.Wizard.Logger.Error(err)
				}
			}
		},
	})

	return step
}

// Show displays the finish step.
func (s *FinishStep) Show() error {
	s.Wizard.GetApp().SetRoot(s.Modal, true)
	return nil
}

// Next returns the next step.
func (s *FinishStep) Next() (display.WizardStep, error) {
	return nil, nil //nolint:nilnil // No next step.
}

// Previous returns the previous step.
func (s *FinishStep) Previous() (display.WizardStep, error) {
	return s.Wizard.Steps[1], nil
}
