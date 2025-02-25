package stop

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func RegisterCommands(app *cli.App, opts *options.CommandOpts) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:      opts.Name(),
		Aliases:   opts.Aliases(),
		Usage:     "Stop Contributoor",
		UsageText: "contributoor stop [options]",
		Action: func(c *cli.Context) error {
			var (
				log          = opts.Logger()
				installerCfg = opts.InstallerConfig()
			)

			sidecarCfg, err := sidecar.NewConfigService(log, c.String("config-path"))
			if err != nil {
				return fmt.Errorf("%s%v%s", tui.TerminalColorRed, err, tui.TerminalColorReset)
			}

			dockerSidecar, err := sidecar.NewDockerSidecar(log, sidecarCfg, installerCfg)
			if err != nil {
				return fmt.Errorf("error creating docker sidecar service: %w", err)
			}

			binarySidecar, err := sidecar.NewBinarySidecar(log, sidecarCfg, installerCfg)
			if err != nil {
				return fmt.Errorf("error creating binary sidecar service: %w", err)
			}

			systemdSidecar, err := sidecar.NewSystemdSidecar(log, sidecarCfg, installerCfg)
			if err != nil {
				return fmt.Errorf("error creating systemd sidecar service: %w", err)
			}

			githubService, err := service.NewGitHubService(log, installerCfg)
			if err != nil {
				return fmt.Errorf("error creating github service: %w", err)
			}

			return stopContributoor(c, log, sidecarCfg, dockerSidecar, systemdSidecar, binarySidecar, githubService)
		},
	})
}

func stopContributoor(
	c *cli.Context,
	log *logrus.Logger,
	sidecarCfg sidecar.ConfigManager,
	docker sidecar.DockerSidecar,
	systemd sidecar.SystemdSidecar,
	binary sidecar.BinarySidecar,
	github service.GitHubService,
) error {
	var (
		runner sidecar.SidecarRunner
		cfg    = sidecarCfg.Get()
	)

	// Stop the sidecar via whatever method the user has configured (docker or binary).
	switch cfg.RunMethod {
	case config.RunMethod_RUN_METHOD_DOCKER:
		runner = docker
	case config.RunMethod_RUN_METHOD_SYSTEMD:
		runner = systemd
	case config.RunMethod_RUN_METHOD_BINARY:
		runner = binary
	default:
		return fmt.Errorf("invalid sidecar run method: %s", cfg.RunMethod)
	}

	// Check version and show upgrade warning if needed.
	current, latest, needsUpdate, err := sidecar.CheckVersion(runner, github, cfg.Version)
	if err == nil && needsUpdate {
		tui.UpgradeWarning(current, latest)
	}

	fmt.Printf("%sStopping Contributoor%s\n", tui.TerminalColorLightBlue, tui.TerminalColorReset)

	if err := runner.Stop(); err != nil {
		return err
	}

	return nil
}
