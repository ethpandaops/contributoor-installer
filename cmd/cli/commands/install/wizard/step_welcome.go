package wizard

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal/display"
)

type WelcomeStep struct {
	*display.ChoiceStep
}

func NewWelcomeStep(w *InstallWizard) *WelcomeStep {
	intro := getIntroText(w.freshInstall)
	helperText := fmt.Sprintf("%s\n\nWelcome to the contributoor configuration wizard!\n\n%s\n\n", display.Logo, intro)
	step := display.NewChoiceStep(w, display.ChoiceStepOptions{
		Step:    1,
		Total:   4,
		Title:   "Welcome",
		Text:    helperText,
		Choices: []string{"Quit", "Next"},
		OnSelect: func(index int) {
			if index == 1 {
				// Get next step and show it
				if next, err := w.CurrentStep.Next(); err == nil {
					w.CurrentStep = next
					w.CurrentStep.Show()
				}
			} else {
				w.GetApp().Stop()
			}
		},
	})

	return &WelcomeStep{step}
}

func (s *WelcomeStep) Show() error {
	s.ChoiceStep.Wizard.GetApp().SetRoot(s.ChoiceStep.Modal, true)
	return nil
}

func (s *WelcomeStep) Next() (display.WizardStep, error) {
	return s.ChoiceStep.Wizard.GetSteps()[1], nil
}

func (s *WelcomeStep) Previous() (display.WizardStep, error) {
	// This is the first step, so no previous
	return nil, nil
}

func (s *WelcomeStep) GetTitle() string {
	return "Welcome"
}

func (s *WelcomeStep) GetProgress() (int, int) {
	return s.ChoiceStep.Step, s.ChoiceStep.Total
}

func getIntroText(freshInstall bool) string {
	if freshInstall {
		return "Since this is your first time configuring the contributoor, we'll walk you through the basic setup.\n\n"
	}
	return "You've already configured contributoor, so any changes you make will update your existing configuration.\n\n"
}
