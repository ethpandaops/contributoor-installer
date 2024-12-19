package start

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/installer"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func RegisterCommands(app *cli.App, opts *options.CommandOpts) error {
	app.Commands = append(app.Commands, cli.Command{
		Name:      opts.Name(),
		Aliases:   opts.Aliases(),
		Usage:     "Start Contributoor",
		UsageText: "contributoor start [options]",
		Action: func(c *cli.Context) error {
			log := opts.Logger()

			installerCfg := installer.NewConfig()

			sidecarConfig, err := sidecar.NewConfigService(log, c.GlobalString("config-path"))
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			dockerSidecar, err := sidecar.NewDockerSidecar(log, sidecarConfig, installerCfg)
			if err != nil {
				return fmt.Errorf("error creating docker sidecar service: %w", err)
			}

			binarySidecar := sidecar.NewBinarySidecar(log, sidecarConfig)

			return startContributoor(c, log, sidecarConfig, dockerSidecar, binarySidecar)
		},
	})

	return nil
}

func startContributoor(
	c *cli.Context,
	log *logrus.Logger,
	config sidecar.ConfigManager,
	docker sidecar.DockerSidecar,
	binary sidecar.BinarySidecar,
) error {
	var (
		runner sidecar.SidecarRunner
		cfg    = config.Get()
	)

	log.WithField("version", cfg.Version).Info("Starting Contributoor")

	// Start the sidecar via whatever method the user has configured (docker or binary).
	switch cfg.RunMethod {
	case sidecar.RunMethodDocker:
		runner = docker
	case sidecar.RunMethodBinary:
		runner = binary
	default:
		return fmt.Errorf("invalid sidecar run method: %s", cfg.RunMethod)
	}

	// Check if the sidecar is already running.
	running, err := runner.IsRunning()
	if err != nil {
		log.Errorf("could not check sidecar status: %v", err)

		return err
	}

	// If the sidecar is already running, we can just return.
	if running {
		return fmt.Errorf("%sContributoor is already running. Use 'contributoor stop' first if you want to restart it%s", tui.TerminalColorRed, tui.TerminalColorReset)
	}

	if err := runner.Start(); err != nil {
		return err
	}

	return nil
}
