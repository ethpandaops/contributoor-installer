package display

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

// WizardStep represents a single step in a wizard.
type WizardStep interface {
	Show() error
	Next() (WizardStep, error)
	Previous() (WizardStep, error)
	GetTitle() string
	GetProgress() (current, total int)
}

// Wizard defines the interface for command wizards.
type Wizard interface {
	// Start begins the wizard flow
	Start() error
	// GetApp returns the tview application
	GetApp() *tview.Application
	// GetPages returns the tview pages
	GetPages() *tview.Pages
	// GetCurrentStep returns the current step
	GetCurrentStep() WizardStep
	// GetSteps returns the steps in the wizard
	GetSteps() []WizardStep
	// OnComplete is called when wizard finishes
	OnComplete() error
}

// BaseWizard provides common wizard functionality.
type BaseWizard struct {
	App         *tview.Application
	Pages       *tview.Pages
	RootPages   *tview.Pages
	CurrentStep WizardStep
	Steps       []WizardStep
	Logger      *logrus.Logger
}

func NewBaseWizard(log *logrus.Logger, app *tview.Application) *BaseWizard {
	w := &BaseWizard{
		App:       app,
		Pages:     tview.NewPages(),
		RootPages: tview.NewPages(),
		Logger:    log,
	}

	// Create the main grid with padding
	grid := tview.NewGrid().
		SetColumns(1, 0, 1).
		SetRows(1, 1, 1, 0, 1)

	grid.SetBackgroundColor(tcell.ColorBlack)
	grid.SetBorder(true).
		SetTitle(fmt.Sprintf(" Contributoor %s Configuration ", "0.0.1")).
		SetBorderColor(tcell.ColorOrange).
		SetTitleColor(tcell.ColorOrange)

	// Add padding
	padding := tview.NewBox().SetBackgroundColor(tcell.ColorBlack)

	for i := 0; i < 3; i++ {
		for j := 0; j < 5; j++ {
			grid.AddItem(padding, j, i, 1, 1, 0, 0, false)
		}
	}

	// Add pages to the center of the grid
	grid.AddItem(w.Pages, 3, 1, 1, 1, 0, 0, true)

	// Set up pages
	w.RootPages.AddPage("main", grid, true, true)
	app.SetRoot(w.RootPages, true)

	return w
}

func (w *BaseWizard) GetApp() *tview.Application {
	return w.App
}

func (w *BaseWizard) GetCurrentStep() WizardStep {
	return w.CurrentStep
}

func (w *BaseWizard) GetPages() *tview.Pages {
	return w.Pages
}

func (w *BaseWizard) GetSteps() []WizardStep {
	return w.Steps
}
