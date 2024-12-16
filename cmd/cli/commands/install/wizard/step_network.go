package wizard

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/rivo/tview"
)

type networkOption struct {
	Label       string
	Value       string
	Description string
}

var availableNetworks = []networkOption{
	{
		Label:       "Ethereum Mainnet",
		Value:       "mainnet",
		Description: "This is the real Ethereum main network.",
	},
	{
		Label:       "Holesky Testnet",
		Value:       "holesky",
		Description: "The Holesky test network.",
	},
	{
		Label:       "Sepolia Testnet",
		Value:       "sepolia",
		Description: "The Sepolia test network.",
	},
}

type NetworkStep struct {
	Wizard          *InstallWizard
	Modal           *tview.Frame
	Step, Total     int
	selectedNetwork string
}

func NewNetworkStep(w *InstallWizard) *NetworkStep {
	step := &NetworkStep{
		Wizard:          w,
		Step:            2,
		Total:           3,
		selectedNetwork: availableNetworks[0].Value,
	}

	// Extract labels and descriptions for the modal
	labels := make([]string, len(availableNetworks))
	descriptions := make([]string, len(availableNetworks))
	for i, network := range availableNetworks {
		labels[i] = network.Label
		descriptions[i] = network.Description
	}

	// Create modal layout
	modal := display.NewChoiceModal(w.GetApp(), display.ChoiceModalOptions{
		Title:        fmt.Sprintf("[%d/%d] Network", step.Step, step.Total),
		Width:        70,
		Text:         "Let's start by choosing which network you'd like to use.",
		Labels:       labels,
		Descriptions: descriptions,
		OnSelect: func(index int) {
			step.selectedNetwork = availableNetworks[index].Value
			// Update config with selected network value
			w.UpdateConfig(func(cfg *service.ContributoorConfig) {
				cfg.NetworkName = step.selectedNetwork
			})
			// Move to next step
			next, _ := step.Next()
			next.Show()
		},
		OnBack: func() {
			prev, _ := step.Previous()
			prev.Show()
		},
	})

	step.Modal = display.CreateWizardFrame(display.WizardFrameOptions{
		Content: modal.BorderGrid,
		Step:    step.Step,
		Total:   step.Total,
		Title:   "Network",
		OnEsc: func() {
			prev, _ := step.Previous()
			prev.Show()
		},
	})

	return step
}

func (s *NetworkStep) Show() error {
	s.Wizard.GetApp().SetRoot(s.Modal, true)
	return nil
}

func (s *NetworkStep) Next() (display.WizardStep, error) {
	return s.Wizard.GetSteps()[2], nil
}

func (s *NetworkStep) Previous() (display.WizardStep, error) {
	return s.Wizard.GetSteps()[0], nil
}

func (s *NetworkStep) GetSelectedNetwork() string {
	return s.selectedNetwork
}
