package stop

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/terminal"
	"github.com/ethpandaops/contributoor-installer-test/internal/service"
	"github.com/urfave/cli"
)

func RegisterCommands(app *cli.App, opts *terminal.CommandOpts) {
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

func stopContributoor(c *cli.Context, opts *terminal.CommandOpts) error {
	log := opts.Logger()

	configService, err := service.NewConfigService(log, c.GlobalString("config-path"))
	if err != nil {
		if _, ok := err.(*service.ConfigNotFoundError); ok {
			return fmt.Errorf("%s%v%s", terminal.ColorRed, err, terminal.ColorReset)
		}

		return fmt.Errorf("%sError loading config: %v%s", terminal.ColorRed, err, terminal.ColorReset)
	}

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

		if !running {
			return fmt.Errorf("%sContributoor is not running. Use 'contributoor start' to start it%s", terminal.ColorRed, terminal.ColorReset)
		}

		if err := dockerService.Stop(); err != nil {
			log.Errorf("could not stop service: %v", err)

			return err
		}
	case service.RunMethodBinary:
		binaryService := service.NewBinaryService(log, configService)
		if err := binaryService.Stop(); err != nil {
			log.Errorf("could not stop service: %v", err)

			return err
		}
	}

	return nil
}
