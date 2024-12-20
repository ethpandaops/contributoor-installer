package restart

import (
	"errors"
	"testing"

	"github.com/ethpandaops/contributoor-installer/internal/sidecar/mock"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	"go.uber.org/mock/gomock"
)

func TestRestartContributoor(t *testing.T) {
	t.Run("restarts running docker sidecar", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock config with docker setup
		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(&config.Config{
			RunMethod: config.RunMethod_RUN_METHOD_DOCKER,
		}).AnyTimes()

		// Create mock docker sidecar that's running
		mockDocker := mock.NewMockDockerSidecar(ctrl)
		mockDocker.EXPECT().IsRunning().Return(true, nil)
		mockDocker.EXPECT().Stop().Return(nil)
		mockDocker.EXPECT().Start().Return(nil)

		err := restartContributoor(
			cli.NewContext(nil, nil, nil),
			logrus.New(),
			mockConfig,
			mockDocker,
			mock.NewMockSystemdSidecar(ctrl),
			mock.NewMockBinarySidecar(ctrl),
		)

		assert.NoError(t, err)
	})

	t.Run("starts stopped systemd sidecar", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock config with systemd setup
		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(&config.Config{
			RunMethod: config.RunMethod_RUN_METHOD_SYSTEMD,
		}).AnyTimes()

		// Create mock systemd sidecar that's not running
		mockSystemd := mock.NewMockSystemdSidecar(ctrl)
		mockSystemd.EXPECT().IsRunning().Return(false, nil)
		mockSystemd.EXPECT().Start().Return(nil)

		err := restartContributoor(
			cli.NewContext(nil, nil, nil),
			logrus.New(),
			mockConfig,
			mock.NewMockDockerSidecar(ctrl),
			mockSystemd,
			mock.NewMockBinarySidecar(ctrl),
		)

		assert.NoError(t, err)
	})

	t.Run("handles stop error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(&config.Config{
			RunMethod: config.RunMethod_RUN_METHOD_BINARY,
		}).AnyTimes()

		// Create mock binary sidecar that fails to stop
		mockBinary := mock.NewMockBinarySidecar(ctrl)
		mockBinary.EXPECT().IsRunning().Return(true, nil)
		mockBinary.EXPECT().Stop().Return(errors.New("test error"))

		err := restartContributoor(
			cli.NewContext(nil, nil, nil),
			logrus.New(),
			mockConfig,
			mock.NewMockDockerSidecar(ctrl),
			mock.NewMockSystemdSidecar(ctrl),
			mockBinary,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to stop service")
	})

	t.Run("handles start error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(&config.Config{
			RunMethod: config.RunMethod_RUN_METHOD_BINARY,
		}).AnyTimes()

		// Create mock binary sidecar that fails to start
		mockBinary := mock.NewMockBinarySidecar(ctrl)
		mockBinary.EXPECT().IsRunning().Return(false, nil)
		mockBinary.EXPECT().Start().Return(errors.New("test error"))

		err := restartContributoor(
			cli.NewContext(nil, nil, nil),
			logrus.New(),
			mockConfig,
			mock.NewMockDockerSidecar(ctrl),
			mock.NewMockSystemdSidecar(ctrl),
			mockBinary,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start service")
	})

	t.Run("handles invalid run method", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfig := mock.NewMockConfigManager(ctrl)
		mockConfig.EXPECT().Get().Return(&config.Config{
			RunMethod: config.RunMethod_RUN_METHOD_UNSPECIFIED,
		}).AnyTimes()

		err := restartContributoor(
			cli.NewContext(nil, nil, nil),
			logrus.New(),
			mockConfig,
			mock.NewMockDockerSidecar(ctrl),
			mock.NewMockSystemdSidecar(ctrl),
			mock.NewMockBinarySidecar(ctrl),
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid sidecar run method")
	})
}
