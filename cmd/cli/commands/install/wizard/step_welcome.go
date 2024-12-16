package wizard

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// WelcomeStep is the first step of the installation wizard.
type WelcomeStep struct {
	Wizard      *InstallWizard
	Modal       *tview.Frame
	Step, Total int
}

func NewWelcomeStep(w *InstallWizard) *WelcomeStep {
	step := &WelcomeStep{
		Wizard: w,
		Step:   1,
		Total:  3,
	}

	intro := "We'll walk you through the basic setup of contributoor.\n\n"
	helperText := fmt.Sprintf("%s\n\nWelcome to the contributoor configuration wizard!\n\n%s\n\n", display.Logo, intro)

	modal := tview.NewModal().
		SetText(helperText).
		AddButtons([]string{"Quit", "Next"}).
		SetButtonStyle(tcell.StyleDefault.
			Background(tcell.ColorDefault).
			Foreground(tcell.ColorLightGray)).
		SetButtonActivatedStyle(tcell.StyleDefault.
			Background(tcell.Color46).
			Foreground(tcell.ColorBlack)).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 1 {
				if next, err := step.Next(); err == nil {
					step.Wizard.CurrentStep = next
					if err := next.Show(); err != nil {
						step.Wizard.Logger.Error(err)
					}
				}
			} else {
				step.Wizard.GetApp().Stop()
			}
		})

	step.Modal = display.CreateWizardFrame(display.WizardFrameOptions{
		Content: modal,
		Step:    step.Step,
		Total:   step.Total,
		Title:   "Welcome",
		OnEsc: func() {
			step.Wizard.GetApp().Stop()
		},
	})

	return step
}

// Show displays the welcome step.
func (s *WelcomeStep) Show() error {
	s.Wizard.GetApp().SetRoot(s.Modal, true)
	return nil
}

// Next returns the next step.
func (s *WelcomeStep) Next() (display.WizardStep, error) {
	return s.Wizard.GetSteps()[1], nil
}

// Previous returns the previous step.
func (s *WelcomeStep) Previous() (display.WizardStep, error) {
	return nil, nil //nolint:nilnil // No previous step.
}
