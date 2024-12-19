package config

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func RegisterCommands(app *cli.App, opts *options.CommandOpts) {
	app.Commands = append(app.Commands, cli.Command{
		Name:      opts.Name(),
		Usage:     "Configure Contributoor settings",
		UsageText: "contributoor config",
		Action: func(c *cli.Context) error {
			log := opts.Logger()

			sidecarConfig, err := sidecar.NewConfigService(log, c.GlobalString("config-path"))
			if err != nil {
				return fmt.Errorf("%serror loading config: %v%s", tui.TerminalColorRed, err, tui.TerminalColorReset)
			}

			return configureContributoor(c, log, sidecarConfig)
		},
	})
}

func configureContributoor(c *cli.Context, log *logrus.Logger, sidecarConfig sidecar.ConfigManager) error {
	var (
		app     = tview.NewApplication()
		display = NewConfigDisplay(log, app, sidecarConfig)
	)

	if err := display.Run(); err != nil {
		return fmt.Errorf("%sdisplay error: %w%s", tui.TerminalColorRed, err, tui.TerminalColorReset)
	}

	return nil
}
