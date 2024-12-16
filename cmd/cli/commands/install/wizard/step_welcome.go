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
		Total:  5,
	}

	// Farm this out into a separate function which we can call here in
	// the constructor and in the Show() method. This is important because
	// steps before this one might have modified the config, which this
	// step conditionally uses.
	step.setupModal()

	return step
}

func (s *WelcomeStep) Show() error {
	s.setupModal()
	s.Wizard.GetApp().SetRoot(s.Modal, true)
	return nil
}

func (s *WelcomeStep) Next() (display.WizardStep, error) {
	return s.Wizard.GetSteps()[1], nil
}

func (s *WelcomeStep) Previous() (display.WizardStep, error) {
	return nil, nil //nolint:nilnil // No previous step.
}

func (s *WelcomeStep) setupModal() {
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
				if next, err := s.Next(); err == nil {
					s.Wizard.CurrentStep = next
					if err := next.Show(); err != nil {
						s.Wizard.Logger.Error(err)
					}
				}
			} else {
				s.Wizard.GetApp().Stop()
			}
		})

	s.Modal = display.CreateWizardFrame(display.WizardFrameOptions{
		Content: modal,
		Step:    s.Step,
		Total:   s.Total,
		Title:   "Welcome",
		OnEsc: func() {
			s.Wizard.GetApp().Stop()
		},
	})
}
