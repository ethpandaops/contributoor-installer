package logs

import (
	"errors"
	"flag"
	"testing"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	sidecarmock "github.com/ethpandaops/contributoor-installer/internal/sidecar/mock"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestShowLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		runMethod     config.RunMethod
		tailLines     int
		follow        bool
		setupMocks    func(*sidecarmock.MockConfigManager, *sidecarmock.MockDockerSidecar, *sidecarmock.MockBinarySidecar, *sidecarmock.MockSystemdSidecar)
		expectedError string
	}{
		{
			name:      "docker - shows logs successfully",
			runMethod: config.RunMethod_RUN_METHOD_DOCKER,
			tailLines: 100,
			follow:    false,
			setupMocks: func(cfg *sidecarmock.MockConfigManager, d *sidecarmock.MockDockerSidecar, b *sidecarmock.MockBinarySidecar, s *sidecarmock.MockSystemdSidecar) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_DOCKER,
				}).Times(1)
				d.EXPECT().Logs(100, false).Return(nil)
			},
		},
		{
			name:      "docker - logs fail",
			runMethod: config.RunMethod_RUN_METHOD_DOCKER,
			tailLines: 100,
			follow:    false,
			setupMocks: func(cfg *sidecarmock.MockConfigManager, d *sidecarmock.MockDockerSidecar, b *sidecarmock.MockBinarySidecar, s *sidecarmock.MockSystemdSidecar) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_DOCKER,
				}).Times(1)
				d.EXPECT().Logs(100, false).Return(errors.New("logs failed"))
			},
			expectedError: "logs failed",
		},
		{
			name:      "binary - shows logs successfully",
			runMethod: config.RunMethod_RUN_METHOD_BINARY,
			tailLines: 50,
			follow:    true,
			setupMocks: func(cfg *sidecarmock.MockConfigManager, d *sidecarmock.MockDockerSidecar, b *sidecarmock.MockBinarySidecar, s *sidecarmock.MockSystemdSidecar) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_BINARY,
				}).Times(1)
				b.EXPECT().Logs(50, true).Return(nil)
			},
		},
		{
			name:      "systemd - shows logs successfully",
			runMethod: config.RunMethod_RUN_METHOD_SYSTEMD,
			tailLines: 200,
			follow:    false,
			setupMocks: func(cfg *sidecarmock.MockConfigManager, d *sidecarmock.MockDockerSidecar, b *sidecarmock.MockBinarySidecar, s *sidecarmock.MockSystemdSidecar) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_SYSTEMD,
				}).Times(1)
				s.EXPECT().Logs(200, false).Return(nil)
			},
		},
		{
			name:      "invalid sidecar run method",
			runMethod: config.RunMethod_RUN_METHOD_UNSPECIFIED,
			tailLines: 100,
			follow:    false,
			setupMocks: func(cfg *sidecarmock.MockConfigManager, d *sidecarmock.MockDockerSidecar, b *sidecarmock.MockBinarySidecar, s *sidecarmock.MockSystemdSidecar) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_UNSPECIFIED,
				}).Times(1)
			},
			expectedError: "invalid sidecar run method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				mockConfig  = sidecarmock.NewMockConfigManager(ctrl)
				mockDocker  = sidecarmock.NewMockDockerSidecar(ctrl)
				mockBinary  = sidecarmock.NewMockBinarySidecar(ctrl)
				mockSystemd = sidecarmock.NewMockSystemdSidecar(ctrl)
			)

			tt.setupMocks(mockConfig, mockDocker, mockBinary, mockSystemd)

			var (
				app = cli.NewApp()
				set = flag.NewFlagSet("test", flag.ContinueOnError)
			)

			set.Int("tail", tt.tailLines, "")
			set.Bool("follow", tt.follow, "")
			ctx := cli.NewContext(app, set, nil)

			err := showLogs(ctx, logrus.New(), mockConfig, mockDocker, mockSystemd, mockBinary)

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
			configPath: "testdata/valid",
		},
		{
			name:          "fails when config service fails",
			configPath:    "/invalid/path/that/doesnt/exist",
			expectedError: "directory [/invalid/path/that/doesnt/exist] does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create dummy CLI app, with the config flag.
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

			RegisterCommands(
				app,
				options.NewCommandOpts(
					options.WithName("logs"),
					options.WithLogger(logrus.New()),
				),
			)

			if tt.expectedError != "" {
				// Ensure the command registration succeeded
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

				// Ensure the command is registered as expected.
				cmd := app.Commands[0]
				assert.Equal(t, "logs", cmd.Name)
				assert.Equal(t, "View Contributoor logs", cmd.Usage)
				assert.Equal(t, "contributoor logs [options]", cmd.UsageText)
				assert.NotNil(t, cmd.Action)

				// Verify flags.
				assert.Len(t, cmd.Flags, 2)
				tailFlag, _ := cmd.Flags[0].(*cli.IntFlag)
				followFlag, _ := cmd.Flags[1].(*cli.BoolFlag)

				assert.Equal(t, "tail", tailFlag.Name)
				assert.Equal(t, 100, tailFlag.Value)
				assert.Equal(t, "follow, f", followFlag.Name)
			}
		})
	}
}
