package install

import (
	"testing"

	"github.com/ethpandaops/contributoor-installer/internal/sidecar/mock"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// This is about the best we can do re testing TUI components.
// They're heavily dependent on the terminal state.
func TestBeaconNodePage(t *testing.T) {
	setupMockDisplay := func(ctrl *gomock.Controller, cfg *config.Config) *InstallDisplay {
		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(cfg).AnyTimes()

		return &InstallDisplay{
			app:        tview.NewApplication(),
			log:        logrus.New(),
			sidecarCfg: mockConfig,
			networkConfigPage: &NetworkConfigPage{
				page: &tui.Page{ID: "network-config"},
			},
		}
	}

	t.Run("creates page with correct structure", func(t *testing.T) {
		// Setup.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDisplay := setupMockDisplay(ctrl, &config.Config{
			BeaconNodeAddress: "http://localhost:5052",
		})

		// Create the page.
		page := NewBeaconNodePage(mockDisplay)

		// Verify the page was created.
		assert.NotNil(t, page, "page should be created")
		assert.NotNil(t, page.content, "page content should be set")
		assert.IsType(t, &tview.Grid{}, page.content, "content should be a grid")
		assert.NotNil(t, page.form, "form should be set")

		// Verify the page ID and title.
		assert.Equal(t, "install-beacon", page.page.ID)
		assert.Equal(t, "Beacon Node", page.page.Title)

		// Verify form structure.
		assert.Equal(t, 1, page.form.GetFormItemCount(), "should have one form item initially")
		assert.NotNil(t, page.form.GetButton(0), "should have Next button")
	})

	t.Run("has correct parent page", func(t *testing.T) {
		// Setup.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDisplay := setupMockDisplay(ctrl, &config.Config{
			BeaconNodeAddress: "http://localhost:5052",
		})

		// Create the page.
		page := NewBeaconNodePage(mockDisplay)

		// Verify parent page is set correctly.
		assert.Equal(t, "network-config", page.page.Parent.ID)
	})
}
