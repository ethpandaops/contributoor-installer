package stop

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func RegisterCommands(app *cli.App, opts *options.CommandOpts) error {
	app.Commands = append(app.Commands, cli.Command{
		Name:      opts.Name(),
		Aliases:   opts.Aliases(),
		Usage:     "Stop Contributoor",
		UsageText: "contributoor stop [options]",
		Action: func(c *cli.Context) error {
			log := opts.Logger()

			configService, err := service.NewConfigService(log, c.GlobalString("config-path"))
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			dockerService, err := service.NewDockerService(log, configService)
			if err != nil {
				return fmt.Errorf("error creating docker service: %w", err)
			}

			binaryService := service.NewBinaryService(log, configService)

			return stopContributoor(c, log, configService, dockerService, binaryService)
		},
	})

	return nil
}

func stopContributoor(
	c *cli.Context,
	log *logrus.Logger,
	config service.ConfigManager,
	docker service.DockerService,
	binary service.BinaryService,
) error {
	var (
		runner service.ServiceRunner
		cfg    = config.Get()
	)

	log.WithField("version", cfg.Version).Info("Stopping Contributoor")

	// Stop the service via whatever method the user has configured (docker or binary).
	switch cfg.RunMethod {
	case service.RunMethodDocker:
		runner = docker
	case service.RunMethodBinary:
		runner = binary
	default:
		return fmt.Errorf("invalid run method: %s", cfg.RunMethod)
	}

	// Check if running before attempting to stop.
	running, err := runner.IsRunning()
	if err != nil {
		log.Errorf("could not check service status: %v", err)

		return err
	}

	// If the service is not running, we can just return.
	if !running {
		return fmt.Errorf("%sContributoor is not running. Use 'contributoor start' to start it%s", tui.TerminalColorRed, tui.TerminalColorReset)
	}

	if err := runner.Stop(); err != nil {
		return err
	}

	return nil
}
