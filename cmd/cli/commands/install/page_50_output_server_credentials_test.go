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
func TestOutputServerCredentialsPage(t *testing.T) {
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
			outputPage: &OutputServerPage{
				page: &tui.Page{ID: "output-server"},
			},
		}
	}

	t.Run("creates page with correct structure", func(t *testing.T) {
		// Setup.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDisplay := setupMockDisplay(ctrl, &sidecar.Config{})

		// Create the page.
		page := NewOutputServerCredentialsPage(mockDisplay)

		// Verify the page was created.
		assert.NotNil(t, page, "page should be created")
		assert.NotNil(t, page.content, "page content should be set")
		assert.IsType(t, &tview.Grid{}, page.content, "content should be a grid")
		assert.NotNil(t, page.form, "form should be set")

		// Verify the page ID and title.
		assert.Equal(t, "install-credentials", page.page.ID)
		assert.Equal(t, "Output Server Credentials", page.page.Title)

		// Verify form structure.
		assert.Equal(t, 2, page.form.GetFormItemCount(), "should have two form items") // Username and Password fields
		assert.NotNil(t, page.form.GetButton(0), "should have Next button")
	})

	t.Run("has correct parent page", func(t *testing.T) {
		// Setup.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDisplay := setupMockDisplay(ctrl, &sidecar.Config{})

		// Create the page.
		page := NewOutputServerCredentialsPage(mockDisplay)

		// Verify parent page is set correctly.
		assert.Equal(t, "output-server", page.page.Parent.ID)
	})

	t.Run("loads existing credentials", func(t *testing.T) {
		// Setup.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create config with existing credentials.
		cfg := &sidecar.Config{
			OutputServer: &sidecar.OutputServerConfig{
				Credentials: "dGVzdHVzZXI6dGVzdHBhc3M=", // base64 encoded "testuser:testpass"
			},
		}

		mockDisplay := setupMockDisplay(ctrl, cfg)

		// Create the page.
		page := NewOutputServerCredentialsPage(mockDisplay)

		// Verify credentials were loaded.
		assert.Equal(t, "testuser", page.username, "should load existing username")
		assert.Equal(t, "testpass", page.password, "should load existing password")
	})

	t.Run("handles invalid credentials gracefully", func(t *testing.T) {
		// Setup.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create config with invalid credentials.
		cfg := &sidecar.Config{
			OutputServer: &sidecar.OutputServerConfig{
				Credentials: "invalid-base64",
			},
		}

		mockDisplay := setupMockDisplay(ctrl, cfg)

		// Create the page.
		page := NewOutputServerCredentialsPage(mockDisplay)

		// Verify invalid credentials don't cause issues.
		assert.Empty(t, page.username, "should have empty username for invalid credentials")
		assert.Empty(t, page.password, "should have empty password for invalid credentials")
	})
}
