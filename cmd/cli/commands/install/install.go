package install

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/commands/install/wizard"
	config "github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal"
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
				Value: fmt.Sprintf("%s", config.ContributorVersion),
			},
			cli.StringFlag{
				Name:  "run-method, r",
				Usage: "The method to run contributoor",
				Value: config.RunMethodDocker,
			},
		},
	})
}

func installContributoor(c *cli.Context) error {
	var (
		freshInstall bool
		configDir    = c.GlobalString("config-path")
		version      = c.String("version")
	)

	log := c.App.Metadata["logger"].(*logrus.Logger)

	if c.GlobalIsSet("config-path") {
		log.Infof("Custom config path provided: %s", configDir)
	}

	if c.IsSet("version") {
		log.Infof("Using provided version: %s", version)
	}

	// Expand the home directory if necessary. Takes care of paths provided with `~`.
	expandedDir, err := homedir.Expand(configDir)
	if err != nil {
		return fmt.Errorf("%sFailed to expand config path: %w%s", colorRed, err, colorReset)
	}

	// Create empty config or load existing
	cfg := config.NewContributoorConfig(expandedDir)

	configPath := filepath.Join(expandedDir, "contributoor.yaml")
	if exists, err := fileExists(configPath); err != nil {
		return err
	} else {
		freshInstall = !exists
		if !freshInstall {
			log.WithFields(logrus.Fields{
				"path": configPath,
			}).Infof("Existing config found")
			if cfg, err = config.LoadConfig(configPath); err != nil {
				return err
			}
		} else {
			log.WithFields(logrus.Fields{
				"path": configPath,
			}).Info("Fresh install detected, creating new config")
		}
	}

	log.Info("Running installation wizard")

	// Create and run the install wizard
	app := tview.NewApplication()
	w := wizard.NewInstallWizard(log, app, cfg, freshInstall)

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
