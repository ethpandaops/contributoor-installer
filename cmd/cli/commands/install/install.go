package install

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/terminal"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func RegisterCommands(app *cli.App, opts *terminal.CommandOpts) {
	app.Commands = append(app.Commands, cli.Command{
		Name:      opts.Name(),
		Aliases:   opts.Aliases(),
		Usage:     "Install Contributoor",
		UsageText: "contributoor install [options]",
		Action: func(c *cli.Context) error {
			return installContributoor(c, opts)
		},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "version, v",
				Usage: "The contributoor version to install",
				Value: "latest",
			},
			cli.StringFlag{
				Name:  "run-method, r",
				Usage: "The method to run contributoor",
				Value: service.RunMethodDocker,
			},
		},
	})
}

func installContributoor(c *cli.Context, opts *terminal.CommandOpts) error {
	log := opts.Logger()
	log.SetLevel(logrus.DebugLevel)

	configService, err := service.NewConfigService(log, c.GlobalString("config-path"))
	if err != nil {
		return fmt.Errorf("%sError loading config: %v%s", terminal.ColorRed, err, terminal.ColorReset)
	}

	app := tview.NewApplication()
	display := NewInstallDisplay(log, app, configService)

	// Run the display
	if err := display.Run(); err != nil {
		log.Errorf("Error running display: %v", err)
		return fmt.Errorf("%sDisplay error: %w%s", terminal.ColorRed, err, terminal.ColorReset)
	}

	// Handle completion
	if err := display.OnComplete(); err != nil {
		log.Errorf("Error completing installation: %v", err)
		return fmt.Errorf("%sCompletion error: %w%s", terminal.ColorRed, err, terminal.ColorReset)
	}

	return nil
}
