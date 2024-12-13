package update

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal/service"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/utils"
)

func RegisterCommands(app *cli.App, opts *utils.CommandOpts) {
	app.Commands = append(app.Commands, cli.Command{
		Name:      opts.Name(),
		Aliases:   opts.Aliases(),
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
			return updateContributoor(c, opts)
		},
	})
}

func updateContributoor(c *cli.Context, opts *utils.CommandOpts) error {
	log := opts.Logger()

	configPath := c.GlobalString("config-path")
	path, err := homedir.Expand(configPath)
	if err != nil {
		return fmt.Errorf("error expanding config path [%s]: %w", configPath, err)
	}

	// Check directory exists
	dirInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("%sYour configured contributoor directory [%s] does not exist. Please run 'contributoor install' first%s", utils.ColorRed, path, utils.ColorReset)
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("%s[%s] is not a directory%s", utils.ColorRed, path, utils.ColorReset)
	}

	// Check config file exists
	configFile := filepath.Join(path, "config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("%sConfig file not found at [%s]. Please run 'contributoor install' first%s", utils.ColorRed, configFile, utils.ColorReset)
	}

	configService, err := service.NewConfigService(log, configPath)
	if err != nil {
		return err
	}

	log.WithField("version", configService.Get().Version).Info("Current version")

	github := service.NewGitHubService("ethpandaops", "contributoor-test")

	// Update version in config if specified
	if c.IsSet("version") {
		tag := c.String("version")
		log.WithField("version", tag).Info("Update version provided")

		if tag == configService.Get().Version {
			log.Infof(
				"%sContributoor is already running version %s%s",
				utils.ColorGreen,
				tag,
				utils.ColorReset,
			)
			return nil
		}

		exists, err := github.VersionExists(tag)
		if err != nil {
			return fmt.Errorf("failed to check version: %w", err)
		}

		if !exists {
			return fmt.Errorf(
				"%sVersion %s not found. Use 'contributoor update' without --version to get the latest version%s",
				utils.ColorRed,
				tag,
				utils.ColorReset,
			)
		}

		configService.Update(func(cfg *service.ContributoorConfig) {
			cfg.Version = tag
		})
	} else {
		tag, err := github.GetLatestVersion()
		if err != nil {
			return fmt.Errorf("failed to get latest version: %w", err)
		}

		log.WithField("version", tag).Info("Latest version detected")

		if tag == configService.Get().Version {
			log.Infof(
				"%sContributoor is up to date%s",
				utils.ColorGreen,
				utils.ColorReset,
			)
			return nil
		}

		if err := configService.Update(func(cfg *service.ContributoorConfig) {
			cfg.Version = tag
		}); err != nil {
			return fmt.Errorf("failed to update config version: %w", err)
		}
	}

	// Save the updated config
	if err := service.WriteConfig(configFile, configService.Get()); err != nil {
		log.Errorf("could not save updated config: %v", err)
		return err
	}

	switch configService.Get().RunMethod {
	case service.RunMethodDocker:
		dockerService, err := service.NewDockerService(log, configService)
		if err != nil {
			log.Errorf("could not create docker service: %v", err)
			return err
		}

		log.WithField("version", configService.Get().Version).Info("Updating Contributoor")

		if err := dockerService.Update(); err != nil {
			log.Errorf("could not update service: %v", err)
			return err
		}

		// Check if service is running
		running, err := dockerService.IsRunning()
		if err != nil {
			log.Errorf("could not check service status: %v", err)
			return err
		}

		if running {
			if utils.Confirm("Service is running. Would you like to restart it with the new version?") {
				if err := dockerService.Stop(); err != nil {
					return fmt.Errorf("failed to stop service: %w", err)
				}
				if err := dockerService.Start(); err != nil {
					return fmt.Errorf("failed to start service: %w", err)
				}
			} else {
				log.Info("Service will continue running with the previous version until next restart")
			}
		} else {
			if utils.Confirm("Service is not running. Would you like to start it?") {
				if err := dockerService.Start(); err != nil {
					return fmt.Errorf("failed to start service: %w", err)
				}
			}
		}

		log.Infof("%sContributoor updated successfully to version %s%s", utils.ColorGreen, configService.Get().Version, utils.ColorReset)
	case service.RunMethodBinary:
		binaryService := service.NewBinaryService(log, configService)
		if err := binaryService.Update(); err != nil {
			log.Errorf("could not update service: %v", err)
			return err
		}

		// Save the updated config back to file
		if err := service.WriteConfig(configFile, configService.Get()); err != nil {
			log.Errorf("could not save updated config: %v", err)
			return err
		}
	}

	return nil
}
