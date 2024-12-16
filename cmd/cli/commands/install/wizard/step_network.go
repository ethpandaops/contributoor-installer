package wizard

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var availableNetworks = []string{
	"mainnet",
	"sepolia",
	"holesky",
	"custom",
}

type NetworkStep struct {
	Wizard      *InstallWizard
	Modal       *tview.Frame
	Step, Total int
	values      map[string]string
	form        *tview.Form
}

func NewNetworkStep(w *InstallWizard) *NetworkStep {
	step := &NetworkStep{
		Wizard: w,
		Step:   2,
		Total:  3,
		values: make(map[string]string),
	}

	modal := tview.NewTextView().
		SetText("Please configure your network settings. Both fields are required.").
		SetTextAlign(tview.AlignCenter).
		SetWordWrap(true).
		SetTextColor(tcell.ColorLightGray).
		SetDynamicColors(true).
		SetBackgroundColor(tcell.ColorBlue).
		SetBorderPadding(1, 1, 0, 0)

	form := tview.NewForm()
	form.SetButtonsAlign(tview.AlignCenter)
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetBackgroundColor(tcell.ColorBlue)
	form.SetBorderPadding(0, 0, 0, 0)
	form.SetButtonTextColor(tcell.ColorLightGray)

	// Add network dropdown
	dropdown := tview.NewDropDown().
		SetLabel("Network Name [::b]▼[::-]: ").
		SetFieldWidth(40).
		SetListStyles(
			tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorLightGray),
			tcell.StyleDefault.Background(tcell.Color46).Foreground(tcell.ColorBlack),
		).
		SetOptions(availableNetworks, func(text string, index int) {
			step.values["Network Name"] = text
			// Show/hide custom network field based on selection.
			if text == "custom" {
				// Remove beacon node input temporarily.
				form.RemoveFormItem(1)

				// Add custom network input with minimal padding.
				customInput := tview.NewInputField().
					SetLabel("Custom Network Name:    ").
					SetFieldWidth(30)
				customInput.SetChangedFunc(func(text string) {
					step.values["Custom Network Name"] = text
				})
				form.AddFormItem(customInput)

				// Re-add beacon node input with minimal padding.
				beaconInput := tview.NewInputField().
					SetLabel("Beacon Node Address:    ").
					SetFieldWidth(30)
				beaconInput.SetChangedFunc(func(text string) {
					step.values["Beacon Node Address"] = text
				})
				form.AddFormItem(beaconInput)

				// Adjust form padding to be more compact.
				form.SetBorderPadding(0, 0, 0, 0)
			} else {
				// If not custom, ensure we only have dropdown and beacon node input.
				if form.GetFormItemCount() > 2 {
					form.RemoveFormItem(1) // Remove custom input if it exists
				}
				step.values["Network Name"] = text
			}
			form.SetFocus(0)
		}).
		SetCurrentOption(0)

	// Set the initial value since the callback won't fire automatically
	step.values["Network Name"] = availableNetworks[0]

	form.AddFormItem(dropdown)

	// Add beacon node input
	input := tview.NewInputField().
		SetLabel("Beacon Node Address:    ").
		SetFieldWidth(30)
	input.SetChangedFunc(func(text string) {
		step.values["Beacon Node Address"] = text
	})
	form.AddFormItem(input)

	form.AddButton("Next", func() {
		if err := w.UpdateConfig(func(cfg *service.ContributoorConfig) {
			cfg.Network = &service.NetworkConfig{
				Name:              step.values["Network Name"],
				CustomNetworkName: step.values["Custom Network Name"],
				BeaconNodeAddress: step.values["Beacon Node Address"],
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
	}).SetButtonStyle(tcell.StyleDefault.
		Background(tcell.ColorDefault).
		Foreground(tcell.ColorLightGray)).
		SetButtonActivatedStyle(tcell.StyleDefault.
			Background(tcell.Color46).
			Foreground(tcell.ColorBlack))

	form.SetFocus(0)

	// Create the control grid with the form.
	controlGrid := tview.NewGrid()
	controlGrid.SetRows(0)
	controlGrid.SetColumns(-2, -6, -2)
	controlGrid.SetBackgroundColor(tcell.ColorBlue)

	leftSpacer := tview.NewBox().SetBackgroundColor(tcell.ColorBlue)
	rightSpacer := tview.NewBox().SetBackgroundColor(tcell.ColorBlue)

	controlGrid.AddItem(leftSpacer, 0, 0, 1, 1, 0, 0, false)
	controlGrid.AddItem(form, 0, 1, 1, 1, 0, 0, true)
	controlGrid.AddItem(rightSpacer, 0, 2, 1, 1, 0, 0, false)

	// Create the content grid.
	contentGrid := tview.NewGrid()
	contentGrid.SetRows(0, 1, 0, -3, 0)
	contentGrid.SetColumns(0)
	contentGrid.SetBackgroundColor(tcell.ColorBlue)
	contentGrid.SetBorder(true)
	contentGrid.SetTitle(" Network Configuration ")

	spacer1 := tview.NewBox().SetBackgroundColor(tcell.ColorBlue)
	spacer2 := tview.NewBox().SetBackgroundColor(tcell.ColorBlue)
	spacer3 := tview.NewBox().SetBackgroundColor(tcell.ColorBlue)

	contentGrid.AddItem(spacer1, 0, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(modal, 1, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(spacer2, 2, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(controlGrid, 3, 0, 1, 1, 0, 0, true)
	contentGrid.AddItem(spacer3, 4, 0, 1, 1, 0, 0, false)

	// Create the centered layout.
	centeredModal := tview.NewFlex().
		AddItem(nil, 0, 2, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 2, false).
				AddItem(contentGrid, 0, 3, true).
				AddItem(nil, 0, 2, false),
			0, 2, true,
		).
		AddItem(nil, 0, 2, false)

	step.Modal = display.CreateWizardFrame(display.WizardFrameOptions{
		Content: centeredModal,
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

	step.form = form

	return step
}

// Show displays the network step.
func (s *NetworkStep) Show() error {
	s.Wizard.GetApp().SetRoot(s.Modal, true)

	// Store form reference during creation
	if s.form != nil {
		s.Wizard.GetApp().SetFocus(s.form)
		s.form.SetFocus(0)
	}
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

	if cfg.Network.Name == "custom" && cfg.Network.CustomNetworkName == "" {
		errorModal := tview.NewModal().
			SetText("Error: Custom network name is required\n\nPlease enter a name for your custom network (e.g. kurtosis, localtestnet, etc.)").
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				s.Wizard.GetApp().SetRoot(s.Modal, true)
			})

		s.Wizard.GetApp().SetRoot(errorModal, true)
		return nil, fmt.Errorf("custom network name is required")
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
