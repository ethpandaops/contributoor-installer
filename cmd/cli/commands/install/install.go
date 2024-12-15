package install

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/commands/install/wizard"
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

	if !c.GlobalIsSet("config-path") {
		log.WithField("config_path", c.GlobalString("config-path")).Warnf("No config path provided, using default")
	}

	configService, err := service.NewConfigService(log, c.GlobalString("config-path"))
	if err != nil {
		if _, ok := err.(*service.ConfigNotFoundError); ok {
			return fmt.Errorf("%sMissing config file. Please run install.sh first.%s", terminal.ColorRed, terminal.ColorReset)
		}

		return fmt.Errorf("%sError loading config: %v%s", terminal.ColorRed, err, terminal.ColorReset)
	}

	// Update config if flags are set
	if c.IsSet("version") || c.IsSet("run-method") {
		if err := configService.Update(func(cfg *service.ContributoorConfig) {
			if c.IsSet("version") {
				cfg.Version = c.String("version")
			}

			if c.IsSet("run-method") {
				cfg.RunMethod = c.String("run-method")
			}
		}); err != nil {
			return fmt.Errorf("%sError updating config: %v%s", terminal.ColorRed, err, terminal.ColorReset)
		}
	}

	log.WithFields(logrus.Fields{
		"config_path": configService.Get().ContributoorDirectory,
		"version":     configService.Get().Version,
		"run_method":  configService.Get().RunMethod,
	}).Info("Running installation wizard")

	// Create and run the install wizard
	app := tview.NewApplication()
	wiz := wizard.NewInstallWizard(log, app, configService)

	if err := wiz.Start(); err != nil {
		return fmt.Errorf("%sWizard error: %w%s", terminal.ColorRed, err, terminal.ColorReset)
	}

	if err := app.Run(); err != nil {
		return fmt.Errorf("%sDisplay error: %w%s", terminal.ColorRed, err, terminal.ColorReset)
	}

	return wiz.OnComplete()
}
