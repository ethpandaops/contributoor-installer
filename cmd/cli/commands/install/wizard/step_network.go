package wizard

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/rivo/tview"
)

type NetworkStep struct {
	Wizard      *InstallWizard
	Modal       *tview.Frame
	Step, Total int
}

func NewNetworkStep(w *InstallWizard) *NetworkStep {
	step := &NetworkStep{
		Wizard: w,
		Step:   2,
		Total:  5,
	}

	// Farm this out into a separate function which we can call here in
	// the constructor and in the Show() method. This is important because
	// steps before this one might have modified the config, which this
	// step conditionally uses.
	step.setupModal()

	return step
}

func (s *NetworkStep) Show() error {
	s.setupModal()
	s.Wizard.GetApp().SetRoot(s.Modal, true)
	return nil
}

func (s *NetworkStep) Next() (display.WizardStep, error) {
	return s.Wizard.GetSteps()[2], nil
}

func (s *NetworkStep) Previous() (display.WizardStep, error) {
	return s.Wizard.GetSteps()[0], nil
}

func (s *NetworkStep) setupModal() {
	// Extract labels and descriptions for the modal
	labels := make([]string, len(display.AvailableNetworks))
	descriptions := make([]string, len(display.AvailableNetworks))
	for i, network := range display.AvailableNetworks {
		labels[i] = network.Label
		descriptions[i] = network.Description
	}

	// Create modal layout
	modal := display.NewChoiceModal(s.Wizard.GetApp(), display.ChoiceModalOptions{
		Title:        fmt.Sprintf("[%d/%d] Network", s.Step, s.Total),
		Width:        70,
		Text:         "Let's start by selecting which network you're using.",
		Labels:       labels,
		Descriptions: descriptions,
		OnSelect: func(index int) {
			// Update config with selected network value
			s.Wizard.UpdateConfig(func(cfg *service.ContributoorConfig) {
				cfg.NetworkName = display.AvailableNetworks[index].Value
			})
			// Move to next step
			next, _ := s.Next()
			next.Show()
		},
		OnBack: func() {
			prev, _ := s.Previous()
			prev.Show()
		},
	})

	s.Modal = display.CreateWizardFrame(display.WizardFrameOptions{
		Content: modal.BorderGrid,
		Step:    s.Step,
		Total:   s.Total,
		Title:   "Network",
		OnEsc: func() {
			prev, _ := s.Previous()
			prev.Show()
		},
	})
}
