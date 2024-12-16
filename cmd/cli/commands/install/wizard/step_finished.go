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

	// Farm this out into a separate function which we can call here in
	// the constructor and in the Show() method. This is important because
	// steps before this one might have modified the config, which this
	// step conditionally uses.
	step.setupModal()

	return step
}

// Show displays the finish step.
func (s *FinishStep) Show() error {
	s.setupModal()
	s.Wizard.GetApp().SetRoot(s.Modal, true)
	return nil
}

// Next returns the next step.
func (s *FinishStep) Next() (display.WizardStep, error) {
	return nil, nil //nolint:nilnil // No next step.
}

// Previous returns the previous step.
func (s *FinishStep) Previous() (display.WizardStep, error) {
	return s.Wizard.Steps[3], nil
}

func (s *FinishStep) setupModal() {
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
			s.Wizard.SetCompleted()
			s.Wizard.GetApp().Stop()
		})

	s.Modal = display.CreateWizardFrame(display.WizardFrameOptions{
		Content: modal,
		Step:    s.Step,
		Total:   s.Total,
		Title:   "Finished",
		OnEsc: func() {
			if prev, err := s.Previous(); err == nil {
				s.Wizard.CurrentStep = prev
				if err := prev.Show(); err != nil {
					s.Wizard.Logger.Error(err)
				}
			}
		},
	})
}
