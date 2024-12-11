package install

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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

const (
	InstallerURL string = "https://gist.githubusercontent.com/mattevans/5cba10243f8fa5fa547aa8cfd96ff888/raw/%s/test.sh"
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
		installerURL = fmt.Sprintf(InstallerURL, version)
	)

	log := c.App.Metadata["logger"].(*logrus.Logger)

	if c.GlobalIsSet("config-path") {
		log.Infof("Custom config path provided: %s", configDir)
	}

	if c.IsSet("version") {
		log.Infof("Using provided version: %s", version)
	}

	log.WithFields(logrus.Fields{
		"url": installerURL,
	}).Info("Downloading installation script")

	// Download the installation script
	resp, err := http.Get(installerURL)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%sUnexpected http status downloading installation script: %d%s", colorRed, resp.StatusCode, colorReset)
	}

	// Sanity check that the script octet length matches content-length
	script, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if fmt.Sprint(len(script)) != resp.Header.Get("content-length") {
		return fmt.Errorf("%sDownloaded script length %d did not match content-length header %s%s", colorRed, len(script), resp.Header.Get("content-length"), colorReset)
	}

	// Expand the home directory if necessary. Takes care of paths provided with `~`.
	expandedDir, err := homedir.Expand(configDir)
	if err != nil {
		return fmt.Errorf("%sFailed to expand config path: %w%s", colorRed, err, colorReset)
	}

	log.Info("Running installation script")

	// Execute the script and capture output
	scriptPath := filepath.Join(os.TempDir(), "install.sh")
	if err := os.WriteFile(scriptPath, script, 0755); err != nil {
		return fmt.Errorf("%sFailed to write installation script: %w%s", colorRed, err, colorReset)
	}
	defer os.Remove(scriptPath)

	cmd := exec.Command(scriptPath, "-p", expandedDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%sFailed to execute installation script: %w\nOutput: %s%s", colorRed, err, output, colorReset)
	}

	log.Info("Installation script completed")

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
