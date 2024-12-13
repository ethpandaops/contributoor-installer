package install

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/commands/install/wizard"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal/service"
	"github.com/mitchellh/go-homedir"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	colorReset     string = "\033[0m"
	colorBold      string = "\033[1m"
	colorRed       string = "\033[31m"
	colorYellow    string = "\033[33m"
	colorGreen     string = "\033[32m"
	colorLightBlue string = "\033[36m"
	clearLine      string = "\033[2K"
)

func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, cli.Command{
		Name:      name,
		Aliases:   aliases,
		Usage:     "Install Contributoor",
		UsageText: "contributoor install [options]",
		Action:    installContributoor,
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

func installContributoor(c *cli.Context) error {
	var (
		configDir = c.GlobalString("config-path")
	)

	log := c.App.Metadata["logger"].(*logrus.Logger)

	// Expand the home directory if necessary
	expandedDir, err := homedir.Expand(configDir)
	if err != nil {
		return fmt.Errorf("%sFailed to expand config path: %w%s", colorRed, err, colorReset)
	}

	if !c.GlobalIsSet("config-path") {
		log.Warnf("No config path provided, using default: %s", expandedDir)
	}

	configPath := filepath.Join(expandedDir, "config.yaml")
	exists, err := fileExists(configPath)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("%sMissing config file. Please run install.sh first.%s", colorRed, colorReset)
	}

	configService, err := service.NewConfigService(log, configPath)
	if err != nil {
		return err
	}

	// Update config if flags are set
	if c.IsSet("version") || c.IsSet("run-method") {
		err = configService.Update(func(cfg *service.ContributoorConfig) {
			if c.IsSet("version") {
				cfg.Version = c.String("version")
			}
			if c.IsSet("run-method") {
				cfg.RunMethod = c.String("run-method")
			}
		})
		if err != nil {
			return err
		}
	}

	log.WithFields(logrus.Fields{
		"config_path": configService.Get().ContributoorDirectory,
		"version":     configService.Get().Version,
		"run_method":  configService.Get().RunMethod,
	}).Info("Running installation wizard")

	// Create and run the install wizard
	app := tview.NewApplication()
	w := wizard.NewInstallWizard(log, app, configService)

	if err := w.Start(); err != nil {
		return fmt.Errorf("%sWizard error: %w%s", colorRed, err, colorReset)
	}

	if err := app.Run(); err != nil {
		return fmt.Errorf("%sDisplay error: %w%s", colorRed, err, colorReset)
	}

	return w.OnComplete()
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}
