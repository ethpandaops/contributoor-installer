package wizard

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/rivo/tview"
)

type OutputCredentialsStep struct {
	Wizard      *InstallWizard
	Modal       *tview.Frame
	Step, Total int
}

func NewOutputCredentialsStep(w *InstallWizard) *OutputCredentialsStep {
	step := &OutputCredentialsStep{
		Wizard: w,
		Step:   5,
		Total:  5,
	}

	// Farm this out into a separate function which we can call here in
	// the constructor and in the Show() method. This is important because
	// steps before this one might have modified the config, which this
	// step conditionally uses.
	step.setupModal()

	return step
}

func (s *OutputCredentialsStep) Show() error {
	s.setupModal()
	s.Wizard.GetApp().SetRoot(s.Modal, true)

	return nil
}

func (s *OutputCredentialsStep) Next() (display.WizardStep, error) {
	return nil, nil // This is the last step
}

func (s *OutputCredentialsStep) Previous() (display.WizardStep, error) {
	return s.Wizard.GetSteps()[3], nil
}

func (s *OutputCredentialsStep) setupModal() {
	var (
		cfg        = s.Wizard.GetConfig()
		helpText   = "Please enter your custom output server address below, with credentials if they are required."
		labels     = []string{"Server Address", "Username", "Password"}
		maxLengths = []int{256, 256, 256}
		isPassword = []bool{false, false, true}
	)

	if cfg.OutputServer == nil {
		cfg.OutputServer = &service.OutputServerConfig{}
	}

	pandaOutputServer := strings.Contains(cfg.OutputServer.Address, "platform.ethpandaops.io")
	if pandaOutputServer {
		helpText = "The ethPandaOps team will have provided you with a username and password. Please enter them below."
		labels = []string{"Username", "Password"}
		maxLengths = []int{256, 256}
		isPassword = []bool{false, true}
	}

	modal := display.NewTextBoxModal(s.Wizard.GetApp(), display.TextBoxModalOptions{
		Title:      fmt.Sprintf("[%d/%d] Output Server Credentials", s.Step, s.Total),
		Width:      70,
		Text:       helpText,
		Labels:     labels,
		MaxLengths: maxLengths,
		IsPassword: isPassword,
		OnDone: func(values map[string]string, setError func(string)) {
			var address string
			username := values["Username"]
			password := values["Password"]

			if pandaOutputServer {
				if username == "" || password == "" {
					setError("Username and password are required")
					return
				}
				address = cfg.OutputServer.Address
			} else {
				address = values["Server Address"]
				if address == "" {
					setError("Server address is required")
					return
				}
				if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
					setError("Server address must start with http:// or https://")
					return
				}
			}

			// Update config with base64 encoded credentials
			if username != "" && password != "" {
				credentials := fmt.Sprintf("%s:%s", username, password)
				encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
				s.Wizard.UpdateConfig(func(cfg *service.ContributoorConfig) {
					if cfg.OutputServer == nil {
						cfg.OutputServer = &service.OutputServerConfig{}
					}
					cfg.OutputServer.Address = address
					cfg.OutputServer.Credentials = encodedCredentials
				})
			}

			// Complete the wizard
			s.Wizard.SetCompleted()
			s.Wizard.GetApp().Stop()
		},
		OnBack: func() {
			prev, _ := s.Previous()
			prev.Show()
		},
		OnEsc: func() {
			prev, _ := s.Previous()
			prev.Show()
		},
	})

	s.Modal = display.CreateWizardFrame(display.WizardFrameOptions{
		Content: modal.BorderGrid,
		Step:    s.Step,
		Total:   s.Total,
		Title:   "Output Server Credentials",
		OnEsc: func() {
			prev, _ := s.Previous()
			prev.Show()
		},
	})
}
