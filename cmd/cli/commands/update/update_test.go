package update

import (
	"errors"
	"flag"
	"testing"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	smock "github.com/ethpandaops/contributoor-installer/internal/service/mock"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar/mock"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
	"go.uber.org/mock/gomock"
)

var confirmResponse bool

func init() {
	tui.Confirm = func(string) bool {
		return confirmResponse
	}
}

func TestUpdateContributoor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	confirmResponse = true

	tests := []struct {
		name          string
		runMethod     string
		version       string
		confirmPrompt bool
		setupMocks    func(*mock.MockConfigManager, *mock.MockDockerSidecar, *mock.MockBinarySidecar, *smock.MockGitHubService)
		expectedError string
	}{
		{
			name:          "docker - updates service successfully",
			runMethod:     sidecar.RunMethodDocker,
			confirmPrompt: true,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, b *mock.MockBinarySidecar, g *smock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&sidecar.Config{
					RunMethod: sidecar.RunMethodDocker,
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
			name:      "docker - already at latest version",
			runMethod: sidecar.RunMethodDocker,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, b *mock.MockBinarySidecar, g *smock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&sidecar.Config{
					RunMethod: sidecar.RunMethodDocker,
					Version:   "v1.0.0",
				}).Times(1)
				g.EXPECT().GetLatestVersion().Return("v1.0.0", nil)
			},
		},
		{
			name:      "docker - update fails",
			runMethod: sidecar.RunMethodDocker,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, b *mock.MockBinarySidecar, g *smock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&sidecar.Config{
					RunMethod: sidecar.RunMethodDocker,
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
			name:          "specific version - exists",
			version:       "v1.1.0",
			confirmPrompt: true,
			runMethod:     sidecar.RunMethodDocker,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, b *mock.MockBinarySidecar, g *smock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&sidecar.Config{
					RunMethod: sidecar.RunMethodDocker,
					Version:   "v1.0.0",
				}).Times(2)
				g.EXPECT().VersionExists("v1.1.0").Return(true, nil)

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
			name:      "specific version - does not exist",
			version:   "v999.0.0",
			runMethod: sidecar.RunMethodDocker,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, b *mock.MockBinarySidecar, g *smock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&sidecar.Config{
					RunMethod: sidecar.RunMethodDocker,
					Version:   "v1.0.0",
				}).Times(1)
				g.EXPECT().VersionExists("v999.0.0").Return(false, nil)
			},
			expectedError: "version v999.0.0 not found",
		},
		{
			name:          "binary - updates service successfully",
			runMethod:     sidecar.RunMethodBinary,
			confirmPrompt: true,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, b *mock.MockBinarySidecar, g *smock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&sidecar.Config{
					RunMethod: sidecar.RunMethodBinary,
					Version:   "v1.0.0",
				}).Times(2)
				g.EXPECT().GetLatestVersion().Return("v1.1.0", nil)

				// Expect a call to update, which in-turn updates + saves the config.
				b.EXPECT().Update().Return(nil)
				cfg.EXPECT().Update(gomock.Any()).Return(nil)
				cfg.EXPECT().Save().Return(nil)

				// Finally, a call is made to see if the service is running.
				b.EXPECT().IsRunning().Return(true, nil)

				// If it is, we expect it to be stopped and started.
				b.EXPECT().Stop().Return(nil)
				b.EXPECT().Start().Return(nil)
			},
		},
		{
			name:      "binary - already at latest version",
			runMethod: sidecar.RunMethodBinary,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, b *mock.MockBinarySidecar, g *smock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&sidecar.Config{
					RunMethod: sidecar.RunMethodBinary,
					Version:   "v1.0.0",
				}).Times(1)
				g.EXPECT().GetLatestVersion().Return("v1.0.0", nil)
			},
		},
		{
			name:          "binary - update fails",
			runMethod:     sidecar.RunMethodBinary,
			confirmPrompt: true,
			setupMocks: func(cfg *mock.MockConfigManager, d *mock.MockDockerSidecar, b *mock.MockBinarySidecar, g *smock.MockGitHubService) {
				cfg.EXPECT().Get().Return(&sidecar.Config{
					RunMethod: sidecar.RunMethodBinary,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confirmResponse = tt.confirmPrompt

			mockConfig := mock.NewMockConfigManager(ctrl)
			mockDocker := mock.NewMockDockerSidecar(ctrl)
			mockBinary := mock.NewMockBinarySidecar(ctrl)
			mockGithub := smock.NewMockGitHubService(ctrl)

			tt.setupMocks(mockConfig, mockDocker, mockBinary, mockGithub)

			app := cli.NewApp()
			app.Flags = []cli.Flag{
				cli.StringFlag{
					Name: "version, v",
				},
			}
			set := flag.NewFlagSet("test", 0)
			set.String("version", "", "")
			if tt.version != "" {
				err := set.Set("version", tt.version)
				require.NoError(t, err)
			}
			context := cli.NewContext(app, set, nil)

			err := updateContributoor(context, logrus.New(), mockConfig, mockDocker, mockBinary, mockGithub)

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
			expectedError: "error loading config",
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
			err = RegisterCommands(
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
				assert.Equal(t, "update", cmd.Name)
				assert.Equal(t, []string{"u"}, cmd.Aliases)
				assert.Equal(t, "Update Contributoor to the latest version", cmd.Usage)
				assert.Equal(t, "contributoor update [options]", cmd.UsageText)
				assert.NotNil(t, cmd.Action)
			}
		})
	}
}
