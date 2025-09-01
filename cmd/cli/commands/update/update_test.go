package update

import (
	"errors"
	"flag"
	"testing"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	smock "github.com/ethpandaops/contributoor-installer/internal/service/mock"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar/mock"
	"github.com/ethpandaops/contributoor-installer/internal/test"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

var confirmResponse bool

// For obvious reasons, we need to mock the confirm prompt. Tests can't be interactive.
func init() {
	tui.Confirm = func(string) bool {
		return confirmResponse
	}
}

func TestUpdateContributoor(t *testing.T) {
	cleanup := test.SuppressOutput(t)
	defer cleanup()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	confirmResponse = true

	tests := []struct {
		name           string
		runMethod      config.RunMethod
		version        string
		confirmPrompt  bool
		nonInteractive bool
		setupMocks     func(*mock.MockConfigManager, *mock.MockDockerSidecar, *mock.MockSystemdSidecar, *mock.MockBinarySidecar, *smock.MockGitHubService)
		expectedError  string
	}{
		{
			name:           "docker - updates service successfully",
			runMethod:      config.RunMethod_RUN_METHOD_DOCKER,
			confirmPrompt:  true,
			nonInteractive: false,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, s *mock.MockSystemdSidecar, b *mock.MockBinarySidecar, g *smock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_DOCKER,
					Version:   "v1.0.0",
				}).Times(2)
				g.EXPECT().GetLatestVersion().Return("v1.1.0", nil)

				// Expect a call to update, which in-turn updates + saves the config.
				d.EXPECT().Update().Return(nil)
				cfg.EXPECT().Update(gomock.Any()).Return(nil)
				cfg.EXPECT().Save().Return(nil)

				// Finally, a call is made to see if the service is running.
				d.EXPECT().IsRunning().Return(true, nil)

				// If it is, we expect it to be stopped and started.
				d.EXPECT().Stop().Return(nil)
				d.EXPECT().Start().Return(nil)
			},
		},
		{
			name:           "docker - already at latest version",
			runMethod:      config.RunMethod_RUN_METHOD_DOCKER,
			nonInteractive: false,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, s *mock.MockSystemdSidecar, b *mock.MockBinarySidecar, g *smock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_DOCKER,
					Version:   "v1.0.0",
				}).Times(1)
				g.EXPECT().GetLatestVersion().Return("v1.0.0", nil)
			},
		},
		{
			name:           "docker - update fails",
			runMethod:      config.RunMethod_RUN_METHOD_DOCKER,
			confirmPrompt:  true,
			nonInteractive: false,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, s *mock.MockSystemdSidecar, b *mock.MockBinarySidecar, g *smock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_DOCKER,
					Version:   "v1.0.0",
				}).Times(2)
				g.EXPECT().GetLatestVersion().Return("v1.1.0", nil)

				// Expect a call to update, which in-turn updates + saves the config.
				d.EXPECT().Update().Return(errors.New("update failed"))
				cfg.EXPECT().Update(gomock.Any()).Return(nil)
				cfg.EXPECT().Save().Return(nil)

				// Because the update failed, expect config to be rolled back.
				cfg.EXPECT().Update(gomock.Any()).Return(nil)
				cfg.EXPECT().Save().Return(nil)
			},
			expectedError: "update failed",
		},
		{
			name:           "binary - already at latest version",
			runMethod:      config.RunMethod_RUN_METHOD_BINARY,
			nonInteractive: false,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, s *mock.MockSystemdSidecar, b *mock.MockBinarySidecar, g *smock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_BINARY,
					Version:   "v1.0.0",
				}).Times(1)
				g.EXPECT().GetLatestVersion().Return("v1.0.0", nil)
			},
		},
		{
			name:           "binary - update fails",
			runMethod:      config.RunMethod_RUN_METHOD_BINARY,
			confirmPrompt:  true,
			nonInteractive: false,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, s *mock.MockSystemdSidecar, b *mock.MockBinarySidecar, g *smock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_BINARY,
					Version:   "v1.0.0",
				}).Times(2)
				g.EXPECT().GetLatestVersion().Return("v1.1.0", nil)

				// Expect check if service is running.
				b.EXPECT().IsRunning().Return(false, nil)

				// Expect a call to update, which in-turn updates + saves the config.
				b.EXPECT().Update().Return(errors.New("update failed"))
				cfg.EXPECT().Update(gomock.Any()).Return(nil)
				cfg.EXPECT().Save().Return(nil)

				// Because the update failed, expect config to be rolled back.
				cfg.EXPECT().Update(gomock.Any()).Return(nil)
				cfg.EXPECT().Save().Return(nil)
			},
			expectedError: "update failed",
		},
		{
			name:           "docker - non-interactive mode auto-restarts",
			runMethod:      config.RunMethod_RUN_METHOD_DOCKER,
			nonInteractive: true,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, s *mock.MockSystemdSidecar, b *mock.MockBinarySidecar, g *smock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_DOCKER,
					Version:   "v1.0.0",
				}).Times(2)
				g.EXPECT().GetLatestVersion().Return("v1.1.0", nil)
				d.EXPECT().Update().Return(nil)
				cfg.EXPECT().Update(gomock.Any()).Return(nil)
				cfg.EXPECT().Save().Return(nil)

				// Check if service is running.
				d.EXPECT().IsRunning().Return(true, nil)

				// In non-interactive mode, should auto-restart.
				d.EXPECT().Stop().Return(nil)
				d.EXPECT().Start().Return(nil)
			},
		},
		{
			name:           "systemd - non-interactive mode auto-restarts",
			runMethod:      config.RunMethod_RUN_METHOD_SYSTEMD,
			nonInteractive: true,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, s *mock.MockSystemdSidecar, b *mock.MockBinarySidecar, g *smock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_SYSTEMD,
					Version:   "v1.0.0",
				}).Times(2)
				g.EXPECT().GetLatestVersion().Return("v1.1.0", nil)

				// Check if service is running.
				s.EXPECT().IsRunning().Return(true, nil)

				// Should stop before update.
				s.EXPECT().Stop().Return(nil)

				// Expect update.
				s.EXPECT().Update().Return(nil)
				cfg.EXPECT().Update(gomock.Any()).Return(nil)
				cfg.EXPECT().Save().Return(nil)

				// In non-interactive mode, should auto-restart.
				s.EXPECT().Start().Return(nil)
			},
		},
		{
			name:           "binary - non-interactive mode auto-restarts",
			runMethod:      config.RunMethod_RUN_METHOD_BINARY,
			nonInteractive: true,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, s *mock.MockSystemdSidecar, b *mock.MockBinarySidecar, g *smock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&config.Config{
					RunMethod: config.RunMethod_RUN_METHOD_BINARY,
					Version:   "v1.0.0",
				}).Times(2)
				g.EXPECT().GetLatestVersion().Return("v1.1.0", nil)

				// Check if service is running.
				b.EXPECT().IsRunning().Return(true, nil)

				// Should stop before update.
				b.EXPECT().Stop().Return(nil)

				// Expect update.
				b.EXPECT().Update().Return(nil)
				cfg.EXPECT().Update(gomock.Any()).Return(nil)
				cfg.EXPECT().Save().Return(nil)

				// In non-interactive mode, should auto-restart.
				b.EXPECT().Start().Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confirmResponse = tt.confirmPrompt

			mockConfig := mock.NewMockConfigManager(ctrl)
			mockDocker := mock.NewMockDockerSidecar(ctrl)
			mockSystemd := mock.NewMockSystemdSidecar(ctrl)
			mockBinary := mock.NewMockBinarySidecar(ctrl)
			mockGithub := smock.NewMockGitHubService(ctrl)

			tt.setupMocks(mockConfig, mockDocker, mockSystemd, mockBinary, mockGithub)

			app := cli.NewApp()
			app.Flags = []cli.Flag{
				&cli.StringFlag{
					Name: "version, v",
				},
				&cli.BoolFlag{
					Name: "non-interactive",
				},
			}
			set := flag.NewFlagSet("test", 0)
			set.String("version", "", "")
			set.Bool("non-interactive", tt.nonInteractive, "")

			if tt.version != "" {
				err := set.Set("version", tt.version)

				require.NoError(t, err)
			}

			context := cli.NewContext(app, set, nil)

			err := updateContributoor(context, logrus.New(), mockConfig, mockDocker, mockSystemd, mockBinary, mockGithub)

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
					options.WithName("update"),
					options.WithLogger(logrus.New()),
					options.WithAliases([]string{"u"}),
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
				assert.Equal(t, "update", cmd.Name)
				assert.Equal(t, []string{"u"}, cmd.Aliases)
				assert.Equal(t, "Update Contributoor to the latest version", cmd.Usage)
				assert.Equal(t, "contributoor update [options]", cmd.UsageText)
				assert.NotNil(t, cmd.Action)
			}
		})
	}
}
