package start

import (
	"errors"
	"flag"
	"testing"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	servicemock "github.com/ethpandaops/contributoor-installer/internal/service/mock"
	sidecarmock "github.com/ethpandaops/contributoor-installer/internal/sidecar/mock"
	"github.com/ethpandaops/contributoor-installer/internal/test"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
	"go.uber.org/mock/gomock"
)

func TestStartContributoor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		runMethod     config.RunMethod
		setupMocks    func(*sidecarmock.MockConfigManager, *sidecarmock.MockDockerSidecar, *sidecarmock.MockBinarySidecar, *servicemock.MockGitHubService)
		expectedError string
	}{
		{
			name:      "docker - starts service successfully",
			runMethod: config.RunMethod_RUN_METHOD_DOCKER,
			setupMocks: func(cfg *sidecarmock.MockConfigManager, d *sidecarmock.MockDockerSidecar, b *sidecarmock.MockBinarySidecar, gh *servicemock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_DOCKER,
					Version:   "latest",
				}).Times(1)
				gh.EXPECT().GetLatestVersion().Return("v1.0.0", nil)
				d.EXPECT().Version().Return("1.0.0", nil)
				d.EXPECT().IsRunning().Return(false, nil)
				d.EXPECT().Start().Return(nil)
			},
		},
		{
			name:      "docker - service already running",
			runMethod: config.RunMethod_RUN_METHOD_DOCKER,
			setupMocks: func(cfg *sidecarmock.MockConfigManager, d *sidecarmock.MockDockerSidecar, b *sidecarmock.MockBinarySidecar, gh *servicemock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_DOCKER,
					Version:   "latest",
				}).Times(1)
				gh.EXPECT().GetLatestVersion().Return("v1.0.0", nil)
				d.EXPECT().Version().Return("1.0.0", nil)
				d.EXPECT().IsRunning().Return(true, nil)
			},
		},
		{
			name:      "docker - start fails",
			runMethod: config.RunMethod_RUN_METHOD_DOCKER,
			setupMocks: func(cfg *sidecarmock.MockConfigManager, d *sidecarmock.MockDockerSidecar, b *sidecarmock.MockBinarySidecar, gh *servicemock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_DOCKER,
					Version:   "latest",
				}).Times(1)
				gh.EXPECT().GetLatestVersion().Return("v1.0.0", nil)
				d.EXPECT().Version().Return("1.0.0", nil)
				d.EXPECT().IsRunning().Return(false, nil)
				d.EXPECT().Start().Return(errors.New("start failed"))
			},
			expectedError: "start failed",
		},
		{
			name:      "binary - starts service successfully",
			runMethod: config.RunMethod_RUN_METHOD_BINARY,
			setupMocks: func(cfg *sidecarmock.MockConfigManager, d *sidecarmock.MockDockerSidecar, b *sidecarmock.MockBinarySidecar, gh *servicemock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_BINARY,
					Version:   "latest",
				}).Times(1)
				gh.EXPECT().GetLatestVersion().Return("v1.0.0", nil)
				b.EXPECT().Version().Return("1.0.0", nil)
				b.EXPECT().IsRunning().Return(false, nil)
				b.EXPECT().Start().Return(nil)
			},
		},
		{
			name:      "binary - service already running",
			runMethod: config.RunMethod_RUN_METHOD_BINARY,
			setupMocks: func(cfg *sidecarmock.MockConfigManager, d *sidecarmock.MockDockerSidecar, b *sidecarmock.MockBinarySidecar, gh *servicemock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_BINARY,
					Version:   "latest",
				}).Times(1)
				gh.EXPECT().GetLatestVersion().Return("v1.0.0", nil)
				b.EXPECT().Version().Return("1.0.0", nil)
				b.EXPECT().IsRunning().Return(true, nil)
			},
		},
		{
			name:      "invalid sidecar run method",
			runMethod: config.RunMethod_RUN_METHOD_UNSPECIFIED,
			setupMocks: func(cfg *sidecarmock.MockConfigManager, d *sidecarmock.MockDockerSidecar, b *sidecarmock.MockBinarySidecar, gh *servicemock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_UNSPECIFIED,
					Version:   "latest",
				}).Times(1)
			},
			expectedError: "invalid sidecar run method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := test.SuppressOutput(t)
			defer cleanup()

			mockConfig := sidecarmock.NewMockConfigManager(ctrl)
			mockDocker := sidecarmock.NewMockDockerSidecar(ctrl)
			mockBinary := sidecarmock.NewMockBinarySidecar(ctrl)
			mockSystemd := sidecarmock.NewMockSystemdSidecar(ctrl)
			mockGitHub := servicemock.NewMockGitHubService(ctrl)

			tt.setupMocks(mockConfig, mockDocker, mockBinary, mockGitHub)

			app := cli.NewApp()
			ctx := cli.NewContext(app, nil, nil)

			err := startContributoor(ctx, logrus.New(), mockConfig, mockDocker, mockSystemd, mockBinary, mockGitHub)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)

				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestRegisterCommands(t *testing.T) {
	cleanup := test.SuppressOutput(t)
	defer cleanup()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		configPath    string
		expectedError string
	}{
		{
			name:       "successfully registers command",
			configPath: "testdata/valid", // "testdata" is an ancillary dir provided by go-test.
		},
		{
			name:          "fails when config service fails",
			configPath:    "/invalid/path/that/doesnt/exist",
			expectedError: "directory [/invalid/path/that/doesnt/exist] does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create CLI app, with the config flag.
			app := cli.NewApp()
			app.Flags = []cli.Flag{
				cli.StringFlag{
					Name: "config-path",
				},
			}

			// Ensure we set the config path flag.
			globalSet := flag.NewFlagSet("test", flag.ContinueOnError)
			globalSet.String("config-path", "", "")
			err := globalSet.Set("config-path", tt.configPath)
			require.NoError(t, err)

			// Create the cmd context.
			globalCtx := cli.NewContext(app, globalSet, nil)
			app.Metadata = map[string]interface{}{
				"flagContext": globalCtx,
			}

			// Now test!
			RegisterCommands(
				app,
				options.NewCommandOpts(
					options.WithName("start"),
					options.WithLogger(logrus.New()),
					options.WithAliases([]string{"s"}),
				),
			)

			if tt.expectedError != "" {
				// Ensure the command registration succeeded.
				assert.NoError(t, err)

				// Assert that the action execution fails as expected.
				cmd := app.Commands[0]
				ctx := cli.NewContext(app, nil, globalCtx)

				// Assert that the action is the func we expect, mainly because the linter is having a fit otherwise.
				action, ok := cmd.Action.(func(*cli.Context) error)
				require.True(t, ok, "expected action to be func(*cli.Context) error")

				// Execute the action and assert the error.
				actionErr := action(ctx)
				assert.Error(t, actionErr)
				assert.ErrorContains(t, actionErr, tt.expectedError)
			} else {
				// Ensure the command registration succeeded.
				assert.NoError(t, err)
				assert.Len(t, app.Commands, 1)

				// Ensure the command is registered as expected by dumping the command.
				cmd := app.Commands[0]
				assert.Equal(t, "start", cmd.Name)
				assert.Equal(t, []string{"s"}, cmd.Aliases)
				assert.Equal(t, "Start Contributoor", cmd.Usage)
				assert.Equal(t, "contributoor start [options]", cmd.UsageText)
				assert.NotNil(t, cmd.Action)
			}
		})
	}
}
