package status

import (
	"errors"
	"testing"

	servicemock "github.com/ethpandaops/contributoor-installer/internal/service/mock"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar/mock"
	"github.com/ethpandaops/contributoor-installer/internal/test"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	"go.uber.org/mock/gomock"
)

func TestShowStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		runMethod     config.RunMethod
		setupMocks    func(*mock.MockConfigManager, *mock.MockDockerSidecar, *mock.MockSystemdSidecar, *mock.MockBinarySidecar, *servicemock.MockGitHubService)
		expectedError string
	}{
		{
			name:      "docker - shows status successfully",
			runMethod: config.RunMethod_RUN_METHOD_DOCKER,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, s *mock.MockSystemdSidecar, b *mock.MockBinarySidecar, g *servicemock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod:         config.RunMethod_RUN_METHOD_DOCKER,
					Version:           "latest",
					NetworkName:       config.NetworkName_NETWORK_NAME_MAINNET,
					BeaconNodeAddress: "http://test:4444",
				}).AnyTimes()
				cfg.EXPECT().GetConfigPath().Return("/test/config.yaml")
				g.EXPECT().GetLatestVersion().Return("v1.0.0", nil)
				d.EXPECT().Version().Return("1.0.0", nil)
				d.EXPECT().IsRunning().Return(true, nil)
				d.EXPECT().Status().Return("running", nil)
			},
		},
		{
			name:      "binary - shows status successfully",
			runMethod: config.RunMethod_RUN_METHOD_BINARY,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, s *mock.MockSystemdSidecar, b *mock.MockBinarySidecar, g *servicemock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod:         config.RunMethod_RUN_METHOD_BINARY,
					Version:           "latest",
					NetworkName:       config.NetworkName_NETWORK_NAME_MAINNET,
					BeaconNodeAddress: "http://test:4444",
				}).AnyTimes()
				cfg.EXPECT().GetConfigPath().Return("/test/config.yaml")
				g.EXPECT().GetLatestVersion().Return("v1.0.0", nil)
				b.EXPECT().Version().Return("1.0.0", nil)
				b.EXPECT().IsRunning().Return(true, nil)
				b.EXPECT().Status().Return("running", nil)
			},
		},
		{
			name:      "systemd - shows status successfully",
			runMethod: config.RunMethod_RUN_METHOD_SYSTEMD,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, s *mock.MockSystemdSidecar, b *mock.MockBinarySidecar, g *servicemock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod:         config.RunMethod_RUN_METHOD_SYSTEMD,
					Version:           "latest",
					NetworkName:       config.NetworkName_NETWORK_NAME_MAINNET,
					BeaconNodeAddress: "http://test:4444",
				}).AnyTimes()
				cfg.EXPECT().GetConfigPath().Return("/test/config.yaml")
				g.EXPECT().GetLatestVersion().Return("v1.0.0", nil)
				s.EXPECT().Version().Return("1.0.0", nil)
				s.EXPECT().IsRunning().Return(true, nil)
				s.EXPECT().Status().Return("active", nil)
			},
		},
		{
			name:      "handles invalid run method",
			runMethod: config.RunMethod_RUN_METHOD_UNSPECIFIED,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, s *mock.MockSystemdSidecar, b *mock.MockBinarySidecar, g *servicemock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_UNSPECIFIED,
					Version:   "latest",
				}).AnyTimes()
			},
			expectedError: "invalid sidecar run method",
		},
		{
			name:      "handles github error gracefully",
			runMethod: config.RunMethod_RUN_METHOD_DOCKER,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, s *mock.MockSystemdSidecar, b *mock.MockBinarySidecar, g *servicemock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod:         config.RunMethod_RUN_METHOD_DOCKER,
					Version:           "latest",
					NetworkName:       config.NetworkName_NETWORK_NAME_MAINNET,
					BeaconNodeAddress: "http://test:4444",
				}).AnyTimes()
				cfg.EXPECT().GetConfigPath().Return("/test/config.yaml")
				g.EXPECT().GetLatestVersion().Return("", errors.New("github error"))
				d.EXPECT().IsRunning().Return(true, nil)
				d.EXPECT().Status().Return("running", nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := test.SuppressOutput(t)
			defer cleanup()

			mockConfig := mock.NewMockConfigManager(ctrl)
			mockDocker := mock.NewMockDockerSidecar(ctrl)
			mockSystemd := mock.NewMockSystemdSidecar(ctrl)
			mockBinary := mock.NewMockBinarySidecar(ctrl)
			mockGitHub := servicemock.NewMockGitHubService(ctrl)

			tt.setupMocks(mockConfig, mockDocker, mockSystemd, mockBinary, mockGitHub)

			app := cli.NewApp()
			ctx := cli.NewContext(app, nil, nil)

			err := showStatus(ctx, logrus.New(), mockConfig, mockDocker, mockSystemd, mockBinary, mockGitHub)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)

				return
			}

			assert.NoError(t, err)
		})
	}
}
