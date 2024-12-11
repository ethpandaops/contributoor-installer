package wizard

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal/display"
	"github.com/rivo/tview"
)

type NetworkStep struct {
	*display.TextBoxStep
}

func NewNetworkStep(w *InstallWizard) *NetworkStep {
	step := display.NewTextBoxStep(w, display.TextBoxStepOptions{
		Step:       3,
		Total:      4,
		Title:      "Network Configuration",
		HelperText: "Please configure your network settings:",
		Width:      60,
		Labels:     []string{"Network Name", "Beacon Node Address"},
		MaxLengths: []int{20, 100},
		Regexes:    []string{"", ""},
		OnDone: func(values map[string]string) {
			w.Config.Network.Name.Value = values["Network Name"]
			w.Config.Network.BeaconNodeAddress.Value = values["Beacon Node Address"]

			if next, err := w.CurrentStep.Next(); err == nil {
				w.CurrentStep = next
				w.CurrentStep.Show()
			}
		},
		OnBack: func() {
			if prev, err := w.CurrentStep.Previous(); err == nil {
				w.CurrentStep = prev
				w.CurrentStep.Show()
			}
		},
		PageID: "network",
	})
	return &NetworkStep{step}
}

func (s *NetworkStep) Show() error {
	s.Wizard.GetApp().SetRoot(s.Modal.BorderGrid, true)
	return nil
}

func (s *NetworkStep) Next() (display.WizardStep, error) {
	// Get InstallWizard instance
	w := s.TextBoxStep.Wizard.(*InstallWizard)

	// Validate network settings
	if w.Config.Network.Name.Value == "" {
		errorModal := tview.NewModal().
			SetText("Error: Network name is required").
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				s.TextBoxStep.Wizard.GetApp().SetRoot(s.Modal.BorderGrid, true)
			})
		s.TextBoxStep.Wizard.GetApp().SetRoot(errorModal, true)
		return nil, fmt.Errorf("network name is required")
	}

	if w.Config.Network.BeaconNodeAddress.Value == "" {
		errorModal := tview.NewModal().
			SetText("Error: Beacon node address is required").
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				s.TextBoxStep.Wizard.GetApp().SetRoot(s.Modal.BorderGrid, true)
			})
		s.TextBoxStep.Wizard.GetApp().SetRoot(errorModal, true)
		return nil, fmt.Errorf("beacon node address is required")
	}

	return w.GetSteps()[3], nil
}

func (s *NetworkStep) Previous() (display.WizardStep, error) {
	return s.Wizard.GetSteps()[2], nil
}

func (s *NetworkStep) GetTitle() string {
	return "Network Configuration"
}

func (s *NetworkStep) GetProgress() (int, int) {
	return s.Step, s.Total
}
