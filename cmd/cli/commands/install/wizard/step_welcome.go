package wizard

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/internal/display"
)

// WelcomeStep is the first step of the installation wizard.
type WelcomeStep struct {
	*display.ChoiceStep
}

func NewWelcomeStep(w *InstallWizard) *WelcomeStep {
	intro := "We'll walk you through the basic setup of contributoor.\n\n"
	helperText := fmt.Sprintf("%s\n\nWelcome to the contributoor configuration wizard!\n\n%s\n\n", display.Logo, intro)
	step := display.NewChoiceStep(w, display.ChoiceStepOptions{
		Step:    1,
		Total:   3,
		Title:   "Welcome",
		Text:    helperText,
		Choices: []string{"Quit", "Next"},
		OnSelect: func(index int) {
			if index == 1 {
				if next, err := w.CurrentStep.Next(); err == nil {
					w.CurrentStep = next
					if err := w.CurrentStep.Show(); err != nil {
						w.Logger.Error(err)
					}
				}
			} else {
				w.GetApp().Stop()
			}
		},
	})

	return &WelcomeStep{step}
}

// Show displays the welcome step.
func (s *WelcomeStep) Show() error {
	s.ChoiceStep.Wizard.GetApp().SetRoot(s.ChoiceStep.Modal, true)

	return nil
}

// Next returns the next step.
func (s *WelcomeStep) Next() (display.WizardStep, error) {
	return s.ChoiceStep.Wizard.GetSteps()[1], nil
}

// Previous returns the previous step.
func (s *WelcomeStep) Previous() (display.WizardStep, error) {
	return nil, nil //nolint:nilnil // No previous step.
}
