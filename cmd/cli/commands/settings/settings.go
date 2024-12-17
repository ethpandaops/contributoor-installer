package settings

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/terminal"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/rivo/tview"
	"github.com/urfave/cli"
)

func RegisterCommands(app *cli.App, opts *terminal.CommandOpts) {
	app.Commands = append(app.Commands, cli.Command{
		Name:      opts.Name(),
		Usage:     "Configure Contributoor settings",
		UsageText: "contributoor settings",
		Action: func(c *cli.Context) error {
			return showSettings(c, opts)
		},
	})
}

func showSettings(c *cli.Context, opts *terminal.CommandOpts) error {
	log := opts.Logger()

	configService, err := service.NewConfigService(log, c.GlobalString("config-path"))
	if err != nil {
		return fmt.Errorf("%sError loading config: %v%s", terminal.ColorRed, err, terminal.ColorReset)
	}

	app := tview.NewApplication()

	if err := NewSettingsDisplay(log, app, configService).Run(); err != nil {
		return fmt.Errorf("%sDisplay error: %w%s", terminal.ColorRed, err, terminal.ColorReset)
	}

	return nil
}
