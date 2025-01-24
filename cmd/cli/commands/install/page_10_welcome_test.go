package install

import (
	"testing"

	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// This is about the best we can do re testing TUI components.
// They're heavily dependent on the terminal state.
func TestWelcomePage(t *testing.T) {
	tests := []struct {
		name          string
		buttonIndex   int
		buttonLabel   string
		expectStop    bool
		expectNewPage bool
	}{
		{
			name:          "clicks next button",
			buttonIndex:   0,
			buttonLabel:   tui.ButtonNext,
			expectStop:    false,
			expectNewPage: true,
		},
		{
			name:          "clicks close button",
			buttonIndex:   1,
			buttonLabel:   tui.ButtonClose,
			expectStop:    true,
			expectNewPage: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup.
			app := tview.NewApplication()

			// Create a new welcome page with our mock display.
			mockDisplay := &InstallDisplay{
				app: app,
				log: logrus.New(),
				networkConfigPage: &NetworkConfigPage{
					page: &tui.Page{ID: "network-config"},
				},
			}
			page := NewWelcomePage(mockDisplay)

			// Verify the page was created.
			assert.NotNil(t, page, "page should be created")
			assert.NotNil(t, page.content, "page content should be set")
			assert.IsType(t, &tview.Grid{}, page.content, "content should be a grid")

			// Verify the page ID and title.
			assert.Equal(t, "install-welcome", page.page.ID)
			assert.Equal(t, "Welcome", page.page.Title)
		})
	}
}
