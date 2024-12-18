package install

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// RegisterCommands registers the install command.
func RegisterCommands(app *cli.App, opts *options.CommandOpts) {
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

// installContributoor is the action for the install command.
func installContributoor(c *cli.Context, opts *options.CommandOpts) error {
	log := opts.Logger()
	log.SetLevel(logrus.DebugLevel)

	configService, err := service.NewConfigService(log, c.GlobalString("config-path"))
	if err != nil {
		return fmt.Errorf("%serror loading config: %v%s", tui.TerminalColorRed, err, tui.TerminalColorReset)
	}

	app := tview.NewApplication()
	d := NewInstallDisplay(log, app, configService)

	// Run the display.
	if err := d.Run(); err != nil {
		log.Errorf("error running display: %v", err)

		return fmt.Errorf("%sdisplay error: %w%s", tui.TerminalColorRed, err, tui.TerminalColorReset)
	}

	// Handle completion.
	if err := d.OnComplete(); err != nil {
		log.Errorf("error completing installation: %v", err)

		return fmt.Errorf("%scompletion error: %w%s", tui.TerminalColorRed, err, tui.TerminalColorReset)
	}

	return nil
}
