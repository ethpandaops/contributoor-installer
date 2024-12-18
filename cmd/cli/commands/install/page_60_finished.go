package install

import (
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FinishedPage is the page for displaying the installation complete message.
type FinishedPage struct {
	display *InstallDisplay
	page    *tui.Page
	content tview.Primitive
	form    *tview.Form
}

// NewFinishedPage creates a new FinishedPage.
func NewFinishedPage(display *InstallDisplay) *FinishedPage {
	finishedPage := &FinishedPage{
		display: display,
	}

	finishedPage.initPage()
	finishedPage.page = tui.NewPage(
		display.outputServerCredentialsPage.GetPage(),
		"install-finished",
		"Installation Complete",
		"Contributoor has been configured successfully",
		finishedPage.content,
	)

	return finishedPage
}

// GetPage returns the page.
func (p *FinishedPage) GetPage() *tui.Page {
	return p.page
}

// initPage initializes the page.
func (p *FinishedPage) initPage() {
	var (
		modalWidth = 70
		lines      = tview.WordWrap("Nice work, you're all done! Contributoor has been configured successfully.", modalWidth-4)
		height     = len(lines) + 4
	)

	// We need a form to house our input fields.
	form := tview.NewForm()
	form.SetButtonsAlign(tview.AlignCenter)
	form.SetBackgroundColor(tui.ColorFormBackground)
	form.SetBorderPadding(0, 0, 0, 0)
	p.form = form

	// Add a 'Finish' button.
	form.AddButton(tui.ButtonClose, func() {
		p.display.app.Stop()
	})

	if button := form.GetButton(0); button != nil {
		button.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		button.SetLabelColor(tcell.ColorLightGray)
		form.SetButtonStyle(tcell.StyleDefault.
			Background(tview.Styles.PrimitiveBackgroundColor).
			Foreground(tcell.ColorLightGray))
		form.SetButtonActivatedStyle(tcell.StyleDefault.
			Background(tui.ColorButtonActivated).
			Foreground(tcell.ColorBlack))
	}

	// Create content grid.
	contentGrid := tview.NewGrid()
	contentGrid.SetRows(2, 3, 1, 6, 1, 2)
	contentGrid.SetColumns(1, -4, 1)
	contentGrid.SetBackgroundColor(tui.ColorFormBackground)

	// Create text view.
	textView := tview.NewTextView()
	textView.SetText("Nice work, you're all done!\nContributoor has been configured successfully.")
	textView.SetTextAlign(tview.AlignCenter)
	textView.SetWordWrap(true)
	textView.SetTextColor(tview.Styles.PrimaryTextColor)
	textView.SetBackgroundColor(tui.ColorFormBackground)
	textView.SetBorderPadding(0, 0, 0, 0)

	// Add items to content grid
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 0, 0, 1, 3, 0, 0, false)
	contentGrid.AddItem(textView, 1, 0, 1, 3, 0, 0, false)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 2, 0, 1, 3, 0, 0, false)
	contentGrid.AddItem(form, 3, 0, 1, 3, 0, 0, true)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(tui.ColorFormBackground), 5, 0, 1, 3, 0, 0, false)
	contentGrid.SetBorder(true)
	contentGrid.SetTitle(" Installation Complete ")
	contentGrid.SetBackgroundColor(tui.ColorFormBackground)

	// Create border grid.
	borderGrid := tview.NewGrid()
	borderGrid.SetColumns(0, modalWidth, 0)
	borderGrid.SetRows(0, height+9, 0, 2)
	borderGrid.SetBackgroundColor(tui.ColorFormBackground)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	p.content = borderGrid
}
