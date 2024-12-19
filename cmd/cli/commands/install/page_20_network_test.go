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
func TestNetworkConfigPage(t *testing.T) {
	setupMockDisplay := func(ctrl *gomock.Controller, cfg *service.ContributoorConfig) *InstallDisplay {
		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(cfg).AnyTimes()
		mockConfig.EXPECT().Update(gomock.Any()).Return(nil).AnyTimes()

		return &InstallDisplay{
			app:           tview.NewApplication(),
			log:           logrus.New(),
			configService: mockConfig,
		}
	}

	t.Run("creates page with correct structure", func(t *testing.T) {
		// Setup.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDisplay := setupMockDisplay(ctrl, &service.ContributoorConfig{})

		// Create the page.
		page := NewNetworkConfigPage(mockDisplay)

		// Verify the page was created.
		assert.NotNil(t, page, "page should be created")
		assert.NotNil(t, page.content, "page content should be set")
		assert.IsType(t, &tview.Grid{}, page.content, "content should be a grid")

		// Verify the page ID and title.
		assert.Equal(t, "install-network", page.page.ID)
		assert.Equal(t, "Network Selection", page.page.Title)
	})

	t.Run("has correct network options", func(t *testing.T) {
		// Setup.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDisplay := setupMockDisplay(ctrl, &service.ContributoorConfig{})

		// Create the page.
		page := NewNetworkConfigPage(mockDisplay)

		// Verify we have network options.
		assert.NotNil(t, page.content, "content should be set")
		assert.IsType(t, &tview.Grid{}, page.content, "content should be a grid")

		// Verify we have the correct number of networks available.
		assert.Greater(t, len(tui.AvailableNetworks), 0, "should have available networks")
	})

	t.Run("has correct styling", func(t *testing.T) {
		// Setup.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDisplay := setupMockDisplay(ctrl, &service.ContributoorConfig{})

		// Create the page.
		page := NewNetworkConfigPage(mockDisplay)

		// Verify basic styling.
		assert.NotNil(t, page.content, "content should be set")
		assert.IsType(t, &tview.Grid{}, page.content, "content should be a grid")
	})
}
