package wizard

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/rivo/tview"
)

// NetworkStep is the network step of the installation wizard.
type NetworkStep struct {
	Wizard      *InstallWizard
	Modal       *tview.Frame
	Step, Total int
}

func NewNetworkStep(w *InstallWizard) *NetworkStep {
	step := &NetworkStep{
		Wizard: w,
		Step:   2,
		Total:  3,
	}

	modalLayout := display.NewTextBoxModal(w.GetApp(), display.TextBoxModalOptions{
		Title:      "Network Configuration",
		Width:      99,
		Text:       "Please configure your network settings. Both fields are required.",
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

			if next, err := step.Next(); err == nil {
				w.CurrentStep = next
				if err := next.Show(); err != nil {
					w.Logger.Error(err)
				}
			}
		},
		OnBack: func() {
			if prev, err := step.Previous(); err == nil {
				w.CurrentStep = prev
				if err := prev.Show(); err != nil {
					w.Logger.Error(err)
				}
			}
		},
		OnEsc: func() {
			if prev, err := step.Previous(); err == nil {
				w.CurrentStep = prev
				if err := prev.Show(); err != nil {
					w.Logger.Error(err)
				}
			}
		},
	})

	step.Modal = display.CreateWizardFrame(display.WizardFrameOptions{
		Content: modalLayout.BorderGrid,
		Step:    step.Step,
		Total:   step.Total,
		Title:   "Network Configuration",
		OnEsc: func() {
			if prev, err := step.Previous(); err == nil {
				w.CurrentStep = prev
				if err := prev.Show(); err != nil {
					w.Logger.Error(err)
				}
			}
		},
	})

	return step
}

// Show displays the network step.
func (s *NetworkStep) Show() error {
	s.Wizard.GetApp().SetRoot(s.Modal, true)
	return nil
}

// Next returns the next step.
func (s *NetworkStep) Next() (display.WizardStep, error) {
	cfg := s.Wizard.GetConfig()

	if cfg.Network == nil {
		if err := s.Wizard.UpdateConfig(func(cfg *service.ContributoorConfig) {
			cfg.Network = &service.NetworkConfig{}
		}); err != nil {
			return nil, err
		}
		cfg = s.Wizard.GetConfig()
	}

	if cfg.Network.Name == "" {
		errorModal := tview.NewModal().
			SetText("Error: Network name is required\n\nPlease enter a name for your network (e.g. mainnet, sepolia, etc.)").
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				s.Wizard.GetApp().SetRoot(s.Modal, true)
			})

		s.Wizard.GetApp().SetRoot(errorModal, true)
		return nil, fmt.Errorf("network name is required")
	}

	if cfg.Network.BeaconNodeAddress == "" {
		errorModal := tview.NewModal().
			SetText("Error: Beacon node address is required\n\nPlease enter the address of your beacon node\n(e.g. http://localhost:5052)").
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				s.Wizard.GetApp().SetRoot(s.Modal, true)
			})

		s.Wizard.GetApp().SetRoot(errorModal, true)
		return nil, fmt.Errorf("beacon node address is required")
	}

	return s.Wizard.GetSteps()[2], nil
}

// Previous returns the previous step.
func (s *NetworkStep) Previous() (display.WizardStep, error) {
	return s.Wizard.GetSteps()[0], nil
}
