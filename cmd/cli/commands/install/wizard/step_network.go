package wizard

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer-test/internal/display"
	"github.com/ethpandaops/contributoor-installer-test/internal/service"
	"github.com/rivo/tview"
)

// NetworkStep is the network step of the installation wizard.
type NetworkStep struct {
	*display.TextBoxStep
}

// NewNetworkStep creates a new network step.
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
			if err := w.UpdateConfig(func(cfg *service.ContributoorConfig) {
				cfg.Network = &service.NetworkConfig{
					Name:              values["Network Name"],
					BeaconNodeAddress: values["Beacon Node Address"],
				}
			}); err != nil {
				w.Logger.Error(err)

				return
			}

			if next, err := w.CurrentStep.Next(); err == nil {
				w.CurrentStep = next

				if err := w.CurrentStep.Show(); err != nil {
					w.Logger.Error(err)
				}
			}
		},
		OnBack: func() {
			if prev, err := w.CurrentStep.Previous(); err == nil {
				w.CurrentStep = prev

				if err := w.CurrentStep.Show(); err != nil {
					w.Logger.Error(err)
				}
			}
		},
		PageID: "network",
	})

	return &NetworkStep{step}
}

// Show displays the network step.
func (s *NetworkStep) Show() error {
	s.Wizard.GetApp().SetRoot(s.Modal.BorderGrid, true)

	return nil
}

// Next returns the next step.
func (s *NetworkStep) Next() (display.WizardStep, error) {
	w, ok := s.TextBoxStep.Wizard.(*InstallWizard)
	if !ok {
		return nil, fmt.Errorf("invalid wizard instance")
	}

	cfg := w.GetConfig()

	if cfg.Network == nil {
		if err := w.UpdateConfig(func(cfg *service.ContributoorConfig) {
			cfg.Network = &service.NetworkConfig{}
		}); err != nil {
			return nil, err
		}

		cfg = w.GetConfig()
	}

	if cfg.Network.Name == "" {
		errorModal := tview.NewModal().
			SetText("Error: Network name is required\n\nPlease enter a name for your network (e.g. mainnet, sepolia, etc.)").
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				s.TextBoxStep.Wizard.GetApp().SetRoot(s.Modal.BorderGrid, true)
			})

		s.TextBoxStep.Wizard.GetApp().SetRoot(errorModal, true)

		return nil, fmt.Errorf("network name is required")
	}

	if cfg.Network.BeaconNodeAddress == "" {
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

// Previous returns the previous step.
func (s *NetworkStep) Previous() (display.WizardStep, error) {
	return s.Wizard.GetSteps()[0], nil
}
