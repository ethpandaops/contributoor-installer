package install

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/commands/install/wizard"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal"
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
				Value: internal.RunMethodDocker,
			},
		},
	})
}

func installContributoor(c *cli.Context) error {
	var (
		configDir = c.GlobalString("config-path")
	)

	log := c.App.Metadata["logger"].(*logrus.Logger)

	// Expand the home directory if necessary. Takes care of paths provided with `~`.
	expandedDir, err := homedir.Expand(configDir)
	if err != nil {
		return fmt.Errorf("%sFailed to expand config path: %w%s", colorRed, err, colorReset)
	}

	// Log a warning if no config path is provided.
	if !c.GlobalIsSet("config-path") {
		log.Warnf("No config path provided, using default: %s", expandedDir)
	}

	var (
		cfg        *internal.ContributoorConfig
		configPath = filepath.Join(expandedDir, "config.yaml")
	)

	exists, err := fileExists(configPath)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("%sMissing config file. Please run install.sh first.%s", colorRed, colorReset)
	}

	if cfg, err = internal.LoadConfig(configPath); err != nil {
		return err
	}

	// If we've been given a version explicitly via flag, use that.
	if c.IsSet("version") {
		cfg.Version = c.String("version")
	}

	// If we've been given a run method explicitly via flag, use that.
	if c.IsSet("run-method") {
		cfg.RunMethod = c.String("run-method")
	}

	log.WithFields(logrus.Fields{
		"config_path": cfg.ContributoorDirectory,
		"version":     cfg.Version,
		"run_method":  cfg.RunMethod,
	}).Info("Running installation wizard")

	// Create and run the install wizard
	app := tview.NewApplication()
	w := wizard.NewInstallWizard(log, app, cfg)

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
