package install

import (
	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type FinishedPage struct {
	display *InstallDisplay
	page    *display.Page
	content tview.Primitive
	form    *tview.Form
}

func NewFinishedPage(id *InstallDisplay) *FinishedPage {
	finishedPage := &FinishedPage{
		display: id,
	}

	finishedPage.initPage()
	finishedPage.page = display.NewPage(
		id.outputServerCredentialsPage.GetPage(),
		"install-finished",
		"Installation Complete",
		"Contributoor has been configured successfully",
		finishedPage.content,
	)

	return finishedPage
}

func (p *FinishedPage) GetPage() *display.Page {
	return p.page
}

func (p *FinishedPage) initPage() {
	// Layout components
	var (
		modalWidth = 70 // Match other pages' width
		lines      = tview.WordWrap("Nice work, you're all done! Contributoor has been configured successfully.", modalWidth-4)
		height     = len(lines) + 4
	)

	// Initialize form
	form := tview.NewForm()
	form.SetButtonsAlign(tview.AlignCenter)
	form.SetBackgroundColor(display.ColorFormBackground)
	form.SetBorderPadding(0, 0, 0, 0)
	p.form = form

	// Add Finish button
	form.AddButton(display.ButtonClose, func() {
		p.display.app.Stop()
	})

	// Style the button
	if button := form.GetButton(0); button != nil {
		button.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		button.SetLabelColor(tcell.ColorLightGray)
		form.SetButtonStyle(tcell.StyleDefault.
			Background(tview.Styles.PrimitiveBackgroundColor).
			Foreground(tcell.ColorLightGray))
		form.SetButtonActivatedStyle(tcell.StyleDefault.
			Background(display.ColorButtonActivated).
			Foreground(tcell.ColorBlack))
	}

	// Create content grid
	contentGrid := tview.NewGrid()
	contentGrid.SetRows(2, 3, 1, 6, 1, 2)
	contentGrid.SetColumns(1, -4, 1)
	contentGrid.SetBackgroundColor(display.ColorFormBackground)

	// Create text view
	textView := tview.NewTextView()
	textView.SetText("Nice work, you're all done!\nContributoor has been configured successfully.")
	textView.SetTextAlign(tview.AlignCenter)
	textView.SetWordWrap(true)
	textView.SetTextColor(tview.Styles.PrimaryTextColor)
	textView.SetBackgroundColor(display.ColorFormBackground)
	textView.SetBorderPadding(0, 0, 0, 0)

	// Add items to content grid
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(display.ColorFormBackground), 0, 0, 1, 3, 0, 0, false)
	contentGrid.AddItem(textView, 1, 0, 1, 3, 0, 0, false)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(display.ColorFormBackground), 2, 0, 1, 3, 0, 0, false)
	contentGrid.AddItem(form, 3, 0, 1, 3, 0, 0, true)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(display.ColorFormBackground), 5, 0, 1, 3, 0, 0, false)

	// Create border grid
	borderGrid := tview.NewGrid()
	borderGrid.SetColumns(0, modalWidth, 0)
	borderGrid.SetRows(0, height+9, 0, 2)
	borderGrid.SetBackgroundColor(display.ColorFormBackground)

	contentGrid.SetBorder(true)
	contentGrid.SetTitle(" Installation Complete ")
	contentGrid.SetBackgroundColor(display.ColorFormBackground)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	p.content = borderGrid
}
