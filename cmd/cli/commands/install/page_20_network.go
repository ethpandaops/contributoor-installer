package install

import (
	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type NetworkPage struct {
	display *InstallDisplay
	page    *display.Page
	content tview.Primitive
}

func NewNetworkPage(id *InstallDisplay) *NetworkPage {
	networkPage := &NetworkPage{
		display: id,
	}

	networkPage.initPage()
	networkPage.page = display.NewPage(
		nil,
		"install-network",
		"Network Selection",
		"Select which network you're using",
		networkPage.content,
	)

	return networkPage
}

func (p *NetworkPage) GetPage() *display.Page {
	return p.page
}

func (p *NetworkPage) initPage() {
	// Layout components
	var (
		// Selectable options
		labels       = make([]string, len(display.AvailableNetworks))
		descriptions = make([]string, len(display.AvailableNetworks))

		// Calculate dimensions
		modalWidth     = 70
		lines          = tview.WordWrap("Select which network you're using", modalWidth-4)
		textViewHeight = len(lines) + 4
		buttonHeight   = len(labels)*2 + 1

		// Main grids
		buttonGrid  = tview.NewGrid()
		contentGrid = tview.NewGrid()
		borderGrid  = tview.NewGrid().SetColumns(0, modalWidth, 0)

		// Flex containers
		formsFlex = tview.NewFlex().SetDirection(tview.FlexRow)

		// Form components
		forms   = make([]*tview.Form, 0)
		descBox *tview.TextView

		// Spacers
		leftSpacer  = tview.NewBox().SetBackgroundColor(display.ColorFormBackground)
		midSpacer   = tview.NewBox().SetBackgroundColor(display.ColorFormBackground)
		rightSpacer = tview.NewBox().SetBackgroundColor(display.ColorFormBackground)
	)

	// Populate selectable options
	for i, network := range display.AvailableNetworks {
		labels[i] = network.Label
		descriptions[i] = network.Description
	}

	// Initialize buttonGrid
	buttonGrid.SetRows(0)
	buttonGrid.SetBackgroundColor(display.ColorFormBackground)

	// Add initial spacing for description box alignment
	formsFlex.AddItem(tview.NewBox().SetBackgroundColor(display.ColorFormBackground), 1, 1, false)

	// Create forms for each button
	for i, label := range labels {
		form := tview.NewForm()
		form.SetButtonsAlign(tview.AlignCenter)
		form.SetBackgroundColor(display.ColorFormBackground)
		form.SetBorderPadding(0, 0, 0, 0)

		// Add button to form
		index := i // Capture index for closure
		form.AddButton(label, func() {
			p.display.configService.Update(func(cfg *service.ContributoorConfig) {
				cfg.NetworkName = display.AvailableNetworks[index].Value
			})
			p.display.setPage(p.display.beaconPage.GetPage())
		})

		// Style the button
		button := form.GetButton(0)
		button.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		button.SetLabelColor(tcell.ColorLightGray)
		form.SetButtonStyle(tcell.StyleDefault.
			Background(tview.Styles.PrimitiveBackgroundColor).
			Foreground(tcell.ColorLightGray)).
			SetButtonActivatedStyle(tcell.StyleDefault.
				Background(display.ColorButtonActivated).
				Foreground(tcell.ColorBlack))

		// Set up navigation
		button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyDown, tcell.KeyTab:
				nextIndex := (index + 1) % len(labels)
				descBox.SetText(descriptions[nextIndex])
				p.display.app.SetFocus(forms[nextIndex])
				return nil
			case tcell.KeyUp:
				nextIndex := index - 1
				if nextIndex < 0 {
					nextIndex = len(labels) - 1
				}
				descBox.SetText(descriptions[nextIndex])
				p.display.app.SetFocus(forms[nextIndex])
				return nil
			case tcell.KeyEsc:
				if p.page.Parent != nil {
					p.display.setPage(p.page.Parent)
				}
				return nil
			}
			return event
		})

		forms = append(forms, form)
		formsFlex.AddItem(form, 1, 1, true)
		formsFlex.AddItem(tview.NewBox().SetBackgroundColor(display.ColorFormBackground), 0, 1, false)
	}

	// Create description box.
	descBox = tview.NewTextView()
	descBox.SetDynamicColors(true)
	descBox.SetText(descriptions[0])
	descBox.SetBackgroundColor(display.ColorFormBackground)
	descBox.SetBorder(true)
	descBox.SetTitle(display.TitleDescription)
	descBox.SetBorderPadding(0, 0, 1, 1)

	// Set up the grids
	buttonGrid.SetColumns(1, -3, 1, -4, 1)
	buttonGrid.AddItem(leftSpacer, 0, 0, 1, 1, 0, 0, false)
	buttonGrid.AddItem(formsFlex, 0, 1, 1, 1, 0, 0, true)
	buttonGrid.AddItem(midSpacer, 0, 2, 1, 1, 0, 0, false)
	buttonGrid.AddItem(descBox, 0, 3, 1, 1, 0, 0, false)
	buttonGrid.AddItem(rightSpacer, 0, 4, 1, 1, 0, 0, false)

	// Create the main text view
	textView := tview.NewTextView()
	textView.SetText("Select which network you're using")
	textView.SetTextAlign(tview.AlignCenter)
	textView.SetWordWrap(true)
	textView.SetTextColor(tview.Styles.PrimaryTextColor)
	textView.SetBackgroundColor(display.ColorFormBackground)
	textView.SetBorderPadding(0, 0, 0, 0)

	// Content grid
	contentGrid.SetRows(2, 2, 1, 0, 1)
	contentGrid.SetBackgroundColor(display.ColorFormBackground)
	contentGrid.SetBorder(true)
	contentGrid.SetTitle(" Network ")

	// Add items to content grid
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(display.ColorFormBackground), 0, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(textView, 1, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(display.ColorFormBackground), 2, 0, 1, 1, 0, 0, false)
	contentGrid.AddItem(buttonGrid, 3, 0, 1, 1, 0, 0, true)
	contentGrid.AddItem(tview.NewBox().SetBackgroundColor(display.ColorFormBackground), 4, 0, 1, 1, 0, 0, false)

	// Border grid
	borderGrid.SetRows(0, textViewHeight+buttonHeight+3, 0, 2)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	// Set initial focus
	p.display.app.SetFocus(forms[0])
	descBox.SetText(descriptions[0])
	p.content = borderGrid
}
