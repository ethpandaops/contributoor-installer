package install

import (
	"testing"

	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar/mock"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// This is about the best we can do re testing TUI components.
// They're heavily dependent on the terminal state.
func TestOutputServerPage(t *testing.T) {
	setupMockDisplay := func(ctrl *gomock.Controller, cfg *sidecar.Config) *InstallDisplay {
		if cfg.OutputServer == nil {
			cfg.OutputServer = &sidecar.OutputServerConfig{}
		}

		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(cfg).AnyTimes()
		mockConfig.EXPECT().Update(gomock.Any()).Return(nil).AnyTimes()

		return &InstallDisplay{
			app:        tview.NewApplication(),
			log:        logrus.New(),
			sidecarCfg: mockConfig,
			beaconPage: &BeaconNodePage{
				page: &tui.Page{ID: "beacon-node"},
			},
		}
	}

	t.Run("creates page with correct structure", func(t *testing.T) {
		// Setup.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDisplay := setupMockDisplay(ctrl, &sidecar.Config{})

		// Create the page.
		page := NewOutputServerPage(mockDisplay)

		// Verify the page was created.
		assert.NotNil(t, page, "page should be created")
		assert.NotNil(t, page.content, "page content should be set")
		assert.IsType(t, &tview.Grid{}, page.content, "content should be a grid")
		assert.NotNil(t, page.form, "form should be set")

		// Verify the page ID and title.
		assert.Equal(t, "install-output", page.page.ID)
		assert.Equal(t, "Output Server", page.page.Title)

		// Verify form structure - initially just has dropdown and button
		assert.Equal(t, 1, page.form.GetFormItemCount(), "should have one form item initially")
		assert.NotNil(t, page.form.GetButton(0), "should have Next button")
	})

	t.Run("has correct parent page", func(t *testing.T) {
		// Setup.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDisplay := setupMockDisplay(ctrl, &sidecar.Config{})

		// Create the page.
		page := NewOutputServerPage(mockDisplay)

		// Verify parent page is set correctly.
		assert.Equal(t, "beacon-node", page.page.Parent.ID)
	})

	t.Run("initializes with default output server config", func(t *testing.T) {
		// Setup.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDisplay := setupMockDisplay(ctrl, &sidecar.Config{})

		// Create the page.
		page := NewOutputServerPage(mockDisplay)

		// Verify the page was created with default config.
		assert.NotNil(t, page, "page should be created")
		assert.NotNil(t, page.form, "form should be set")
	})
}
