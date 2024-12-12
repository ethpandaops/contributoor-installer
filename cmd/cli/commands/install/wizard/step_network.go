package wizard

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal/display"
	"github.com/rivo/tview"
)

type NetworkStep struct {
	*display.TextBoxStep
}

func NewNetworkStep(w *InstallWizard) *NetworkStep {
	step := display.NewTextBoxStep(w, display.TextBoxStepOptions{
		Step:       2,
		Total:      3,
		Title:      "Network Configuration",
		HelperText: "Please configure your network settings. Both fields are required.",
		Width:      99,
		Labels:     []string{"Network Name", "Beacon Node Address"},
		MaxLengths: []int{20, 100},
		Regexes:    []string{"", ""},
		OnDone: func(values map[string]string) {
			w.Config.Network = &internal.NetworkConfig{
				Name:              values["Network Name"],
				BeaconNodeAddress: values["Beacon Node Address"],
			}

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

	if w.Config.Network == nil {
		w.Config.Network = &internal.NetworkConfig{}
	}

	// Validate network settings
	if w.Config.Network.Name == "" {
		errorModal := tview.NewModal().
			SetText("Error: Network name is required\n\nPlease enter a name for your network (e.g. mainnet, sepolia, etc.)").
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				s.TextBoxStep.Wizard.GetApp().SetRoot(s.Modal.BorderGrid, true)
			})
		s.TextBoxStep.Wizard.GetApp().SetRoot(errorModal, true)
		return nil, fmt.Errorf("network name is required")
	}

	if w.Config.Network.BeaconNodeAddress == "" {
		errorModal := tview.NewModal().
			SetText("Error: Beacon node address is required\n\nPlease enter the address of your beacon node\n(e.g. http://localhost:5052)").
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				s.TextBoxStep.Wizard.GetApp().SetRoot(s.Modal.BorderGrid, true)
			})
		s.TextBoxStep.Wizard.GetApp().SetRoot(errorModal, true)
		return nil, fmt.Errorf("beacon node address is required")
	}

	return w.GetSteps()[2], nil
}

func (s *NetworkStep) Previous() (display.WizardStep, error) {
	return s.Wizard.GetSteps()[0], nil
}

func (s *NetworkStep) GetTitle() string {
	return "Network Configuration"
}

func (s *NetworkStep) GetProgress() (int, int) {
	return s.Step, s.Total
}
