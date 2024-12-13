package update

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal/service"
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
		Usage:     "Update Contributoor to the latest version",
		UsageText: "contributoor update [options]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "version, v",
				Usage: "The contributoor version to update to",
				Value: "latest",
			},
		},
		Action: func(c *cli.Context) error {
			return updateContributoor(c)
		},
	})
}

func promptYesNo(prompt string) bool {
	var response string
	fmt.Printf("\n%s%s [y/N]: %s\n", colorYellow, prompt, colorReset)
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y"
}

func updateContributoor(c *cli.Context) error {
	configPath := c.GlobalString("config-path")
	path, err := homedir.Expand(configPath)
	if err != nil {
		return fmt.Errorf("error expanding config path [%s]: %w", configPath, err)
	}

	// Check directory exists
	dirInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("%sYour configured contributoor directory [%s] does not exist. Please run 'contributoor install' first%s", colorRed, path, colorReset)
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("%s[%s] is not a directory%s", colorRed, path, colorReset)
	}

	// Check config file exists
	configFile := filepath.Join(path, "config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("%sConfig file not found at [%s]. Please run 'contributoor install' first%s", colorRed, configFile, colorReset)
	}

	logger := c.App.Metadata["logger"].(*logrus.Logger)

	cfg, err := internal.LoadConfig(configFile)
	if err != nil {
		return err
	}

	github := service.NewGitHubService("ethpandaops", "contributoor-test")

	// Update version in config if specified
	if c.IsSet("version") {
		requestedVersion := c.String("version")
		logger.WithField("version", requestedVersion).Info("Update version provided")

		exists, err := github.VersionExists(requestedVersion)
		if err != nil {
			return fmt.Errorf("failed to check version: %w", err)
		}
		if !exists {
			return fmt.Errorf(
				"%sVersion %s not found. Use 'contributoor update' without --version to get the latest version%s",
				colorRed,
				requestedVersion,
				colorReset,
			)
		}

		cfg.Version = requestedVersion
	} else {
		tag, err := github.GetLatestVersion()
		if err != nil {
			return fmt.Errorf("failed to get latest version: %w", err)
		}

		logger.WithField("version", tag).Info("Latest version detected")
		cfg.Version = tag
	}

	// Save the updated config
	if err := cfg.WriteToFile(configFile); err != nil {
		logger.Errorf("could not save updated config: %v", err)
		return err
	}

	switch cfg.RunMethod {
	case internal.RunMethodDocker:
		dockerService, err := service.NewDockerService(logger, cfg)
		if err != nil {
			logger.Errorf("could not create docker service: %v", err)
			return err
		}

		logger.WithField("version", cfg.Version).Info("Updating Contributoor")
		if err := dockerService.Update(); err != nil {
			logger.Errorf("could not update service: %v", err)
			return err
		}

		// Check if service is running
		running, err := dockerService.IsRunning()
		if err != nil {
			logger.Errorf("could not check service status: %v", err)
			return err
		}

		if running {
			if promptYesNo("Service is running. Would you like to restart it with the new version?") {
				if err := dockerService.Stop(); err != nil {
					return fmt.Errorf("failed to stop service: %w", err)
				}
				if err := dockerService.Start(); err != nil {
					return fmt.Errorf("failed to start service: %w", err)
				}
			} else {
				logger.Info("Service will continue running with the previous version until next restart")
			}
		} else {
			if promptYesNo("Service is not running. Would you like to start it?") {
				if err := dockerService.Start(); err != nil {
					return fmt.Errorf("failed to start service: %w", err)
				}
			}
		}

		logger.Infof("%sContributoor updated successfully to version %s%s", colorGreen, cfg.Version, colorReset)
	case internal.RunMethodBinary:
		binaryService := service.NewBinaryService(logger, cfg)
		if err := binaryService.Update(); err != nil {
			logger.Errorf("could not update service: %v", err)
			return err
		}

		// Save the updated config back to file
		if err := cfg.WriteToFile(configFile); err != nil {
			logger.Errorf("could not save updated config: %v", err)
			return err
		}
	}

	return nil
}
