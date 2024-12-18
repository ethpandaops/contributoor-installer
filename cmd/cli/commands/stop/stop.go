package stop

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/urfave/cli"
)

func RegisterCommands(app *cli.App, opts *options.CommandOpts) {
	app.Commands = append(app.Commands, cli.Command{
		Name:      opts.Name(),
		Aliases:   opts.Aliases(),
		Usage:     "Stop Contributoor",
		UsageText: "contributoor stop [options]",
		Action: func(c *cli.Context) error {
			return stopContributoor(c, opts)
		},
	})
}

func stopContributoor(c *cli.Context, opts *options.CommandOpts) error {
	log := opts.Logger()

	configService, err := service.NewConfigService(log, c.GlobalString("config-path"))
	if err != nil {
		if _, ok := err.(*service.ConfigNotFoundError); ok {
			return fmt.Errorf("%s%v%s", tui.TerminalColorRed, err, tui.TerminalColorReset)
		}

		return fmt.Errorf("%sError loading config: %v%s", tui.TerminalColorRed, err, tui.TerminalColorReset)
	}

	// Stop the service via whatever method the user has configured (docker or binary).
	switch configService.Get().RunMethod {
	case service.RunMethodDocker:
		log.WithField("version", configService.Get().Version).Info("Stopping Contributoor")

		dockerService, err := service.NewDockerService(log, configService)
		if err != nil {
			log.Errorf("could not create docker service: %v", err)

			return err
		}

		// Check if running before attempting to stop.
		running, err := dockerService.IsRunning()
		if err != nil {
			log.Errorf("could not check service status: %v", err)

			return err
		}

		// If the service is not running, we can just return.
		if !running {
			return fmt.Errorf("%sContributoor is not running. Use 'contributoor start' to start it%s", tui.TerminalColorRed, tui.TerminalColorReset)
		}

		if err := dockerService.Stop(); err != nil {
			log.Errorf("could not stop service: %v", err)

			return err
		}
	case service.RunMethodBinary:
		binaryService := service.NewBinaryService(log, configService)

		// Check if the service is currently running.
		running, err := binaryService.IsRunning()
		if err != nil {
			return fmt.Errorf("failed to check service status: %v", err)
		}

		// If the service is not running, we can just return.
		if !running {
			return fmt.Errorf("%sContributoor is not running%s", tui.TerminalColorRed, tui.TerminalColorReset)
		}

		if err := binaryService.Stop(); err != nil {
			return err
		}
	}

	return nil
}
