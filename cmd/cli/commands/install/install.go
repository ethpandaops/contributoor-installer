package install

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
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
			log := opts.Logger()

			sidecarConfig, err := sidecar.NewConfigService(log, c.GlobalString("config-path"))
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			return installContributoor(c, log, sidecarConfig)
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
				Value: sidecar.RunMethodDocker,
			},
		},
	})
}

func installContributoor(c *cli.Context, log *logrus.Logger, sidecarConfig sidecar.ConfigManager) error {
	var (
		app     = tview.NewApplication()
		display = NewInstallDisplay(log, app, sidecarConfig)
	)

	// Run the display.
	if err := display.Run(); err != nil {
		log.Errorf("error running display: %v", err)

		return fmt.Errorf("%sdisplay error: %w%s", tui.TerminalColorRed, err, tui.TerminalColorReset)
	}

	// Handle completion.
	if err := display.OnComplete(); err != nil {
		log.Errorf("error completing installation: %v", err)

		return fmt.Errorf("%scompletion error: %w%s", tui.TerminalColorRed, err, tui.TerminalColorReset)
	}

	return nil
}
