package stop

import (
	"errors"
	"flag"
	"testing"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	servicemock "github.com/ethpandaops/contributoor-installer/internal/service/mock"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar/mock"
	"github.com/ethpandaops/contributoor-installer/internal/test"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestStopContributoor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		runMethod     config.RunMethod
		setupMocks    func(*mock.MockConfigManager, *mock.MockDockerSidecar, *mock.MockBinarySidecar, *servicemock.MockGitHubService)
		expectedError string
	}{
		{
			name:      "docker - stops service successfully",
			runMethod: config.RunMethod_RUN_METHOD_DOCKER,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, b *mock.MockBinarySidecar, g *servicemock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_DOCKER,
					Version:   "latest",
				}).Times(1)
				g.EXPECT().GetLatestVersion().Return("v1.0.0", nil)
				d.EXPECT().Version().Return("1.0.0", nil)
				d.EXPECT().Stop().Return(nil)
			},
		},
		{
			name:      "docker - stop fails",
			runMethod: config.RunMethod_RUN_METHOD_DOCKER,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, b *mock.MockBinarySidecar, g *servicemock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_DOCKER,
					Version:   "latest",
				}).Times(1)
				g.EXPECT().GetLatestVersion().Return("v1.0.0", nil)
				d.EXPECT().Version().Return("1.0.0", nil)
				d.EXPECT().Stop().Return(errors.New("stop failed"))
			},
			expectedError: "stop failed",
		},
		{
			name:      "binary - stops service successfully",
			runMethod: config.RunMethod_RUN_METHOD_BINARY,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, b *mock.MockBinarySidecar, g *servicemock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_BINARY,
					Version:   "latest",
				}).Times(1)
				g.EXPECT().GetLatestVersion().Return("v1.0.0", nil)
				b.EXPECT().Version().Return("1.0.0", nil)
				b.EXPECT().Stop().Return(nil)
			},
		},
		{
			name:      "invalid sidecar run method",
			runMethod: config.RunMethod_RUN_METHOD_UNSPECIFIED,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, b *mock.MockBinarySidecar, g *servicemock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_UNSPECIFIED,
					Version:   "latest",
				}).Times(1)
			},
			expectedError: "invalid sidecar run method",
		},
		{
			name:      "github error is handled gracefully",
			runMethod: config.RunMethod_RUN_METHOD_DOCKER,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, b *mock.MockBinarySidecar, g *servicemock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_DOCKER,
					Version:   "latest",
				}).Times(1)
				g.EXPECT().GetLatestVersion().Return("", errors.New("github error"))
				d.EXPECT().Stop().Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := test.SuppressOutput(t)
			defer cleanup()

			mockConfig := mock.NewMockConfigManager(ctrl)
			mockDocker := mock.NewMockDockerSidecar(ctrl)
			mockBinary := mock.NewMockBinarySidecar(ctrl)
			mockSystemd := mock.NewMockSystemdSidecar(ctrl)
			mockGitHub := servicemock.NewMockGitHubService(ctrl)

			tt.setupMocks(mockConfig, mockDocker, mockBinary, mockGitHub)

			app := cli.NewApp()
			ctx := cli.NewContext(app, nil, nil)

			err := stopContributoor(ctx, logrus.New(), mockConfig, mockDocker, mockSystemd, mockBinary, mockGitHub)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)

				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestRegisterCommands(t *testing.T) {
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
			cleanup := test.SuppressOutput(t)
			defer cleanup()

			// Create CLI app, with the config flag.
			app := cli.NewApp()
			app.Flags = []cli.Flag{
				&cli.StringFlag{
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
					options.WithName("stop"),
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

				// Execute the action and assert the error.
				actionErr := cmd.Action(ctx)
				assert.Error(t, actionErr)
				assert.ErrorContains(t, actionErr, tt.expectedError)
			} else {
				// Ensure the command registration succeeded.
				assert.NoError(t, err)
				assert.Len(t, app.Commands, 1)

				// Ensure the command is registered as expected by dumping the command.
				cmd := app.Commands[0]
				assert.Equal(t, "stop", cmd.Name)
				assert.Equal(t, []string{"s"}, cmd.Aliases)
				assert.Equal(t, "Stop Contributoor", cmd.Usage)
				assert.Equal(t, "contributoor stop [options]", cmd.UsageText)
				assert.NotNil(t, cmd.Action)
			}
		})
	}
}
