package config

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/rivo/tview"
	"github.com/urfave/cli"
)

func RegisterCommands(app *cli.App, opts *options.CommandOpts) {
	app.Commands = append(app.Commands, cli.Command{
		Name:      opts.Name(),
		Usage:     "Configure Contributoor settings",
		UsageText: "contributoor config",
		Action: func(c *cli.Context) error {
			return showConfig(c, opts)
		},
	})
}

func showConfig(c *cli.Context, opts *options.CommandOpts) error {
	log := opts.Logger()

	configService, err := service.NewConfigService(log, c.GlobalString("config-path"))
	if err != nil {
		return fmt.Errorf("%sError loading config: %v%s", tui.TerminalColorRed, err, tui.TerminalColorReset)
	}

	app := tview.NewApplication()

	if err := NewConfigDisplay(log, app, configService).Run(); err != nil {
		return fmt.Errorf("%sDisplay error: %w%s", tui.TerminalColorRed, err, tui.TerminalColorReset)
	}

	return nil
}
