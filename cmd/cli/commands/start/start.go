package start

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
)

func RegisterCommands(app *cli.App, opts *options.CommandOpts) {
	app.Commands = append(app.Commands, cli.Command{
		Name:      opts.Name(),
		Aliases:   opts.Aliases(),
		Usage:     "Start Contributoor",
		UsageText: "contributoor start [options]",
		Action: func(c *cli.Context) error {
			return startContributoor(c, opts)
		},
	})
}

func startContributoor(c *cli.Context, opts *options.CommandOpts) error {
	log := opts.Logger()

	configService, err := service.NewConfigService(log, c.GlobalString("config-path"))
	if err != nil {
		return fmt.Errorf("%sError loading config: %v%s", tui.TerminalColorRed, err, tui.TerminalColorReset)
	}

	// Start the service via whatever method the user has configured (docker or binary).
	switch configService.Get().RunMethod {
	case service.RunMethodDocker:
		log.WithField("version", configService.Get().Version).Info("Starting Contributoor")

		dockerService, err := service.NewDockerService(log, configService)
		if err != nil {
			log.Errorf("could not create docker service: %v", err)

			return err
		}

		// Check if the service is already running.
		running, err := dockerService.IsRunning()
		if err != nil {
			log.Errorf("could not check service status: %v", err)

			return err
		}

		// If the service is already running, we can just return.
		if running {
			return fmt.Errorf("%sContributoor is already running. Use 'contributoor stop' first if you want to restart it%s", tui.TerminalColorRed, tui.TerminalColorReset)
		}

		if err := dockerService.Start(); err != nil {
			log.Errorf("could not start service: %v", err)

			return err
		}

	case service.RunMethodBinary:
		binaryService := service.NewBinaryService(log, configService)

		// Check if the service is currently running.
		running, err := binaryService.IsRunning()
		if err != nil {
			return fmt.Errorf("failed to check service status: %v", err)
		}

		// If the service is already running, we can just return.
		if running {
			return fmt.Errorf("%sContributoor is already running%s", tui.TerminalColorRed, tui.TerminalColorReset)
		}

		if err := binaryService.Start(); err != nil {
			return err
		}
	}

	return nil
}
