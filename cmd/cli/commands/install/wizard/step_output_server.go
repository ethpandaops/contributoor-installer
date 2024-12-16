package wizard

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/rivo/tview"
)

type outputServerOption struct {
	Label       string
	Value       string
	Description string
}

var availableOutputServers = []outputServerOption{
	{
		Label:       "ethPandaOps Production",
		Value:       "https://xatu.primary.production.platform.ethpandaops.io",
		Description: "The production server provided by ethPandaOps.",
	},
	{
		Label:       "ethPandaOps Staging",
		Value:       "https://xatu.primary.staging.platform.ethpandaops.io",
		Description: "The staging server provided by ethPandaOps.",
	},
	{
		Label:       "Custom",
		Value:       "custom",
		Description: "Use your own custom output server.",
	},
}

type OutputServerStep struct {
	Wizard      *InstallWizard
	Modal       *tview.Frame
	Step, Total int
}

func NewOutputServerStep(w *InstallWizard) *OutputServerStep {
	step := &OutputServerStep{
		Wizard: w,
		Step:   4,
		Total:  5,
	}

	// Farm this out into a separate function which we can call here in
	// the constructor and in the Show() method. This is important because
	// steps before this one might have modified the config, which this
	// step conditionally uses.
	step.setupModal()

	return step
}

func (s *OutputServerStep) Show() error {
	s.setupModal()
	s.Wizard.GetApp().SetRoot(s.Modal, true)

	return nil
}

func (s *OutputServerStep) Next() (display.WizardStep, error) {
	return s.Wizard.GetSteps()[4], nil
}

func (s *OutputServerStep) Previous() (display.WizardStep, error) {
	return s.Wizard.GetSteps()[2], nil
}

func (s *OutputServerStep) setupModal() {
	labels := make([]string, len(availableOutputServers))
	descriptions := make([]string, len(availableOutputServers))
	for i, server := range availableOutputServers {
		labels[i] = server.Label
		descriptions[i] = server.Description
	}

	modal := display.NewChoiceModal(s.Wizard.GetApp(), display.ChoiceModalOptions{
		Title:        fmt.Sprintf("[%d/%d] Output Server", s.Step, s.Total),
		Width:        80,
		Text:         "Select the output server you'd like to use. This is the server where your events will be sent to.",
		Labels:       labels,
		Descriptions: descriptions,
		OnSelect: func(index int) {
			s.Wizard.UpdateConfig(func(cfg *service.ContributoorConfig) {
				cfg.OutputServer = &service.OutputServerConfig{
					Address:     availableOutputServers[index].Value,
					Credentials: "",
				}
			})

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
		Title:   "Output Server",
		OnEsc: func() {
			prev, _ := s.Previous()
			prev.Show()
		},
	})
}
