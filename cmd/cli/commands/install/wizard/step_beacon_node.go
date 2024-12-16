package wizard

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/rivo/tview"
)

type BeaconNodeStep struct {
	Wizard      *InstallWizard
	Modal       *tview.Frame
	Step, Total int
}

func NewBeaconNodeStep(w *InstallWizard) *BeaconNodeStep {
	step := &BeaconNodeStep{
		Wizard: w,
		Step:   3,
		Total:  3,
	}

	// Create modal layout
	modal := display.NewTextBoxModal(w.GetApp(), display.TextBoxModalOptions{
		Title: fmt.Sprintf("[%d/%d] Beacon Node", step.Step, step.Total),
		Width: 70,
		Text:  "Please enter the address of your Beacon Node.\nFor example: http://localhost:5052",
		Labels: []string{
			"Beacon Node Address",
		},
		MaxLengths: []int{
			256, // reasonable max length for a URL
		},
		Regexes: []string{
			`^https?://[^\s/$.?#].[^\s]*$`, // Basic URL validation
		},
		OnDone: func(values map[string]string, setError func(string)) {
			address := values["Beacon Node Address"]

			// Validate beacon node
			if err := validateBeaconNode(address); err != nil {
				setError(err.Error())
				return
			}

			// Update config with beacon node address
			w.UpdateConfig(func(cfg *service.ContributoorConfig) {
				cfg.BeaconNodeAddress = address
			})

			// Complete the wizard
			w.SetCompleted()
			w.GetApp().Stop()
		},
		OnBack: func() {
			prev, _ := step.Previous()
			prev.Show()
		},
		OnEsc: func() {
			prev, _ := step.Previous()
			prev.Show()
		},
	})

	step.Modal = display.CreateWizardFrame(display.WizardFrameOptions{
		Content: modal.BorderGrid,
		Step:    step.Step,
		Total:   step.Total,
		Title:   "Beacon Node",
		OnEsc: func() {
			prev, _ := step.Previous()
			prev.Show()
		},
	})

	return step
}

func validateBeaconNode(address string) error {
	// Check if URL is valid
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		return fmt.Errorf("Beacon node address must start with http:// or https://")
	}

	// Try to connect to the beacon node
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(fmt.Sprintf("%s/eth/v1/node/health", address))
	if err != nil {
		return fmt.Errorf("We're unable to connect to your beacon node: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Beacon node returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *BeaconNodeStep) Show() error {
	s.Wizard.GetApp().SetRoot(s.Modal, true)
	return nil
}

func (s *BeaconNodeStep) Next() (display.WizardStep, error) {
	return nil, nil // This is the last step
}

func (s *BeaconNodeStep) Previous() (display.WizardStep, error) {
	return s.Wizard.GetSteps()[1], nil
}
