package install

import (
	"testing"

	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/ethpandaops/contributoor-installer/internal/service/mock"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// This is about the best we can do re testing TUI components.
// They're heavily dependent on the terminal state.
func TestFinishedPage(t *testing.T) {
	setupMockDisplay := func(ctrl *gomock.Controller) *InstallDisplay {
		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(&service.ContributoorConfig{}).AnyTimes()

		return &InstallDisplay{
			app:           tview.NewApplication(),
			log:           logrus.New(),
			configService: mockConfig,
			outputServerCredentialsPage: &OutputServerCredentialsPage{
				page: &tui.Page{ID: "output-server-credentials"},
			},
		}
	}

	t.Run("creates page with correct structure", func(t *testing.T) {
		// Setup.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDisplay := setupMockDisplay(ctrl)

		// Create the page.
		page := NewFinishedPage(mockDisplay)

		// Verify the page was created.
		assert.NotNil(t, page, "page should be created")
		assert.NotNil(t, page.content, "page content should be set")
		assert.IsType(t, &tview.Grid{}, page.content, "content should be a grid")
		assert.NotNil(t, page.form, "form should be set")

		// Verify the page ID and title.
		assert.Equal(t, "install-finished", page.page.ID)
		assert.Equal(t, "Installation Complete", page.page.Title)

		// Verify form structure.
		assert.Equal(t, 0, page.form.GetFormItemCount(), "should have no form items")
		assert.NotNil(t, page.form.GetButton(0), "should have Close button")
	})

	t.Run("has correct parent page", func(t *testing.T) {
		// Setup.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDisplay := setupMockDisplay(ctrl)

		// Create the page.
		page := NewFinishedPage(mockDisplay)

		// Verify parent page is set correctly.
		assert.Equal(t, "output-server-credentials", page.page.Parent.ID)
	})
}
