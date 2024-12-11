package wizard

import (
	"fmt"

	config "github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal/display"
)

type ModeStep struct {
	*display.ChoiceStep
}

func NewModeStep(w *InstallWizard) *ModeStep {
	helperText := fmt.Sprintf("Now lets select how you want to run contributoor\n\nYou can run contributoor using a binary or docker container.\n\n")
	step := display.NewChoiceStep(w, display.ChoiceStepOptions{
		Step:    2,
		Total:   4,
		Title:   "Run Mode",
		Text:    helperText,
		Choices: []string{"Back", "Docker", "Binary"},
		OnSelect: func(index int) {
			if index == 0 {
				w.CurrentStep = w.GetSteps()[0]
				w.CurrentStep.Show()
				return
			} else {
				switch index {
				case 1:
					w.Config.RunMethod = config.RunMethodDocker
				case 2:
					w.Config.RunMethod = config.RunMethodBinary
				}

				if next, err := w.CurrentStep.Next(); err == nil {
					w.CurrentStep = next
					w.CurrentStep.Show()
				}
			}
		},
	})

	return &ModeStep{step}
}

func (s *ModeStep) Show() error {
	s.ChoiceStep.Wizard.GetApp().SetRoot(s.ChoiceStep.Modal, true)
	return nil
}

func (s *ModeStep) Next() (display.WizardStep, error) {
	return s.ChoiceStep.Wizard.GetSteps()[2], nil
}

func (s *ModeStep) Previous() (display.WizardStep, error) {
	return s.Wizard.GetSteps()[1], nil
}

func (s *ModeStep) GetTitle() string {
	return "Run Mode"
}

func (s *ModeStep) GetProgress() (int, int) {
	return s.ChoiceStep.Step, s.ChoiceStep.Total
}
