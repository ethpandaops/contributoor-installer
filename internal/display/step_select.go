package display

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ChoiceStep is a generic wizard step that presents choices to the user
type ChoiceStep struct {
	Wizard      Wizard
	Modal       *tview.Modal
	Step, Total int
	choices     []string
	onSelect    func(index int)
}

// NewChoiceStep creates a new choice step
func NewChoiceStep(wizard Wizard, opts ChoiceStepOptions) *ChoiceStep {
	step := &ChoiceStep{
		Wizard:   wizard,
		Step:     opts.Step,
		Total:    opts.Total,
		choices:  opts.Choices,
		onSelect: opts.OnSelect,
	}

	title := fmt.Sprintf("[%d/%d] %s", opts.Step, opts.Total, opts.Title)

	modal := tview.NewModal()
	modal.SetText(opts.Text)
	modal.SetTitle(title)
	modal.AddButtons(opts.Choices)
	modal.SetButtonStyle(tcell.StyleDefault.
		Background(tcell.ColorDefault).
		Foreground(tcell.ColorLightGray))
	modal.SetButtonActivatedStyle(tcell.StyleDefault.
		Background(tcell.Color46).
		Foreground(tcell.ColorBlack))
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if opts.OnSelect != nil {
			opts.OnSelect(buttonIndex)
		}
	})

	step.Modal = modal
	return step
}

// ChoiceStepOptions configures a choice step
type ChoiceStepOptions struct {
	Step     int
	Total    int
	Title    string
	Text     string
	Choices  []string
	OnSelect func(index int)
}

// WizardStep interface implementation
func (s *ChoiceStep) Show() error {
	s.Wizard.GetApp().SetRoot(s.Modal, true)
	return nil
}

func (s *ChoiceStep) Next() (WizardStep, error) {
	return s.Wizard.GetCurrentStep(), nil
}

func (s *ChoiceStep) Previous() (WizardStep, error) {
	return nil, nil
}

func (s *ChoiceStep) GetTitle() string {
	return s.Modal.GetTitle()
}

func (s *ChoiceStep) GetProgress() (int, int) {
	return s.Step, s.Total
}
