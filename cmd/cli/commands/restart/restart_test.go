package restart

import (
	"errors"
	"testing"

	servicemock "github.com/ethpandaops/contributoor-installer/internal/service/mock"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar/mock"
	"github.com/ethpandaops/contributoor-installer/internal/test"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestRestartContributoor(t *testing.T) {
	t.Run("restarts running docker sidecar", func(t *testing.T) {
		cleanup := test.SuppressOutput(t)
		defer cleanup()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock config with docker setup
		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(&config.Config{
			RunMethod: config.RunMethod_RUN_METHOD_DOCKER,
			Version:   "latest",
		}).AnyTimes()

		// Create mock docker sidecar that's running
		mockDocker := mock.NewMockDockerSidecar(ctrl)
		mockDocker.EXPECT().Version().Return("1.0.0", nil)
		mockDocker.EXPECT().IsRunning().Return(true, nil)
		mockDocker.EXPECT().Stop().Return(nil)
		mockDocker.EXPECT().Start().Return(nil)

		// Create mock GitHub service
		mockGitHub := servicemock.NewMockGitHubService(ctrl)
		mockGitHub.EXPECT().GetLatestVersion().Return("v1.0.0", nil)

		err := restartContributoor(
			cli.NewContext(nil, nil, nil),
			logrus.New(),
			mockConfig,
			mockDocker,
			mock.NewMockSystemdSidecar(ctrl),
			mock.NewMockBinarySidecar(ctrl),
			mockGitHub,
		)

		assert.NoError(t, err)
	})

	t.Run("starts stopped systemd sidecar", func(t *testing.T) {
		cleanup := test.SuppressOutput(t)
		defer cleanup()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock config with systemd setup
		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(&config.Config{
			RunMethod: config.RunMethod_RUN_METHOD_SYSTEMD,
			Version:   "latest",
		}).AnyTimes()

		// Create mock systemd sidecar that's not running
		mockSystemd := mock.NewMockSystemdSidecar(ctrl)
		mockSystemd.EXPECT().Version().Return("1.0.0", nil)
		mockSystemd.EXPECT().IsRunning().Return(false, nil)
		mockSystemd.EXPECT().Start().Return(nil)

		// Create mock GitHub service
		mockGitHub := servicemock.NewMockGitHubService(ctrl)
		mockGitHub.EXPECT().GetLatestVersion().Return("v1.0.0", nil)

		err := restartContributoor(
			cli.NewContext(nil, nil, nil),
			logrus.New(),
			mockConfig,
			mock.NewMockDockerSidecar(ctrl),
			mockSystemd,
			mock.NewMockBinarySidecar(ctrl),
			mockGitHub,
		)

		assert.NoError(t, err)
	})

	t.Run("handles stop error", func(t *testing.T) {
		cleanup := test.SuppressOutput(t)
		defer cleanup()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(&config.Config{
			RunMethod: config.RunMethod_RUN_METHOD_BINARY,
			Version:   "latest",
		}).AnyTimes()

		// Create mock binary sidecar that fails to stop
		mockBinary := mock.NewMockBinarySidecar(ctrl)
		mockBinary.EXPECT().Version().Return("1.0.0", nil)
		mockBinary.EXPECT().IsRunning().Return(true, nil)
		mockBinary.EXPECT().Stop().Return(errors.New("test error"))

		// Create mock GitHub service
		mockGitHub := servicemock.NewMockGitHubService(ctrl)
		mockGitHub.EXPECT().GetLatestVersion().Return("v1.0.0", nil)

		err := restartContributoor(
			cli.NewContext(nil, nil, nil),
			logrus.New(),
			mockConfig,
			mock.NewMockDockerSidecar(ctrl),
			mock.NewMockSystemdSidecar(ctrl),
			mockBinary,
			mockGitHub,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to stop service")
	})

	t.Run("handles start error", func(t *testing.T) {
		cleanup := test.SuppressOutput(t)
		defer cleanup()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(&config.Config{
			RunMethod: config.RunMethod_RUN_METHOD_BINARY,
			Version:   "latest",
		}).AnyTimes()

		// Create mock binary sidecar that fails to start
		mockBinary := mock.NewMockBinarySidecar(ctrl)
		mockBinary.EXPECT().Version().Return("1.0.0", nil)
		mockBinary.EXPECT().IsRunning().Return(false, nil)
		mockBinary.EXPECT().Start().Return(errors.New("test error"))

		// Create mock GitHub service
		mockGitHub := servicemock.NewMockGitHubService(ctrl)
		mockGitHub.EXPECT().GetLatestVersion().Return("v1.0.0", nil)

		err := restartContributoor(
			cli.NewContext(nil, nil, nil),
			logrus.New(),
			mockConfig,
			mock.NewMockDockerSidecar(ctrl),
			mock.NewMockSystemdSidecar(ctrl),
			mockBinary,
			mockGitHub,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start service")
	})

	t.Run("handles invalid run method", func(t *testing.T) {
		cleanup := test.SuppressOutput(t)
		defer cleanup()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(&config.Config{
			RunMethod: config.RunMethod_RUN_METHOD_UNSPECIFIED,
			Version:   "latest",
		}).AnyTimes()

		err := restartContributoor(
			cli.NewContext(nil, nil, nil),
			logrus.New(),
			mockConfig,
			mock.NewMockDockerSidecar(ctrl),
			mock.NewMockSystemdSidecar(ctrl),
			mock.NewMockBinarySidecar(ctrl),
			servicemock.NewMockGitHubService(ctrl),
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid sidecar run method")
	})

	t.Run("handles github error gracefully", func(t *testing.T) {
		cleanup := test.SuppressOutput(t)
		defer cleanup()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(&config.Config{
			RunMethod: config.RunMethod_RUN_METHOD_DOCKER,
			Version:   "latest",
		}).AnyTimes()

		mockDocker := mock.NewMockDockerSidecar(ctrl)
		mockDocker.EXPECT().IsRunning().Return(true, nil)
		mockDocker.EXPECT().Stop().Return(nil)
		mockDocker.EXPECT().Start().Return(nil)

		// Create mock GitHub service that returns an error
		mockGitHub := servicemock.NewMockGitHubService(ctrl)
		mockGitHub.EXPECT().GetLatestVersion().Return("", errors.New("github error"))

		err := restartContributoor(
			cli.NewContext(nil, nil, nil),
			logrus.New(),
			mockConfig,
			mockDocker,
			mock.NewMockSystemdSidecar(ctrl),
			mock.NewMockBinarySidecar(ctrl),
			mockGitHub,
		)

		// The restart should still succeed even if GitHub check fails
		assert.NoError(t, err)
	})
}
