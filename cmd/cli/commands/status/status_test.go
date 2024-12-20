package status

import (
	"fmt"
	"testing"

	servicemock "github.com/ethpandaops/contributoor-installer/internal/service/mock"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar/mock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	"go.uber.org/mock/gomock"
)

func TestShowStatus(t *testing.T) {
	t.Run("shows status for running docker sidecar", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock config with docker setup
		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(&sidecar.Config{
			Version:           "1.0.0",
			RunMethod:         sidecar.RunMethodDocker,
			NetworkName:       "mainnet",
			BeaconNodeAddress: "http://localhost:5052",
			OutputServer: &sidecar.OutputServerConfig{
				Address: "https://output.server",
			},
		}).AnyTimes()
		mockConfig.EXPECT().GetConfigPath().Return("/path/to/config.yaml")

		// Create mock docker sidecar that's running
		mockDocker := mock.NewMockDockerSidecar(ctrl)
		mockDocker.EXPECT().IsRunning().Return(true, nil)

		// Create mock binary sidecar (shouldn't be used)
		mockBinary := mock.NewMockBinarySidecar(ctrl)

		// Create mock systemd sidecar (shouldn't be used)
		mockSystemd := mock.NewMockSystemdSidecar(ctrl)

		// Create mock GitHub service
		mockGithub := servicemock.NewMockGitHubService(ctrl)
		mockGithub.EXPECT().GetLatestVersion().Return("1.0.1", nil)

		err := showStatus(
			cli.NewContext(nil, nil, nil),
			logrus.New(),
			mockConfig,
			mockDocker,
			mockSystemd,
			mockBinary,
			mockGithub,
		)

		assert.NoError(t, err)
	})

	t.Run("shows status for stopped binary sidecar", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock config with binary setup
		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(&sidecar.Config{
			Version:           "1.0.0",
			RunMethod:         sidecar.RunMethodBinary,
			NetworkName:       "mainnet",
			BeaconNodeAddress: "http://localhost:5052",
		}).AnyTimes()
		mockConfig.EXPECT().GetConfigPath().Return("/path/to/config.yaml")

		// Create mock docker sidecar (shouldn't be used)
		mockDocker := mock.NewMockDockerSidecar(ctrl)

		// Create mock binary sidecar that's stopped
		mockBinary := mock.NewMockBinarySidecar(ctrl)
		mockBinary.EXPECT().IsRunning().Return(false, nil)

		// Create mock systemd sidecar (shouldn't be used)
		mockSystemd := mock.NewMockSystemdSidecar(ctrl)

		// Create mock GitHub service with same version (shouldn't show update)
		mockGithub := servicemock.NewMockGitHubService(ctrl)
		mockGithub.EXPECT().GetLatestVersion().Return("1.0.0", nil)

		err := showStatus(
			cli.NewContext(nil, nil, nil),
			logrus.New(),
			mockConfig,
			mockDocker,
			mockSystemd,
			mockBinary,
			mockGithub,
		)

		assert.NoError(t, err)
	})

	t.Run("handles github service error gracefully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(&sidecar.Config{
			Version:     "1.0.0",
			RunMethod:   sidecar.RunMethodDocker,
			NetworkName: "mainnet",
		}).AnyTimes()
		mockConfig.EXPECT().GetConfigPath().Return("/path/to/config.yaml")

		mockDocker := mock.NewMockDockerSidecar(ctrl)
		mockDocker.EXPECT().IsRunning().Return(true, nil)
		mockSystemd := mock.NewMockSystemdSidecar(ctrl)

		// Create mock GitHub service that returns an error
		mockGithub := servicemock.NewMockGitHubService(ctrl)
		mockGithub.EXPECT().GetLatestVersion().Return("", fmt.Errorf("github error"))

		err := showStatus(
			cli.NewContext(nil, nil, nil),
			logrus.New(),
			mockConfig,
			mockDocker,
			mockSystemd,
			mock.NewMockBinarySidecar(ctrl),
			mockGithub,
		)

		assert.NoError(t, err) // Should still succeed even with GitHub error
	})

	t.Run("handles invalid run method", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock config with invalid run method
		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(&sidecar.Config{
			RunMethod: "invalid",
		})

		err := showStatus(
			cli.NewContext(nil, nil, nil),
			logrus.New(),
			mockConfig,
			nil,
			nil,
			nil,
			nil,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid sidecar run method")
	})
}
