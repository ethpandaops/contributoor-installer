package update

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/terminal"
	"github.com/ethpandaops/contributoor-installer/internal/service"
)

// RegisterCommands registers the update command.
func RegisterCommands(app *cli.App, opts *terminal.CommandOpts) {
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

// updateContributoor updates contributoor to the latest version (based on run method configured in config.yaml).
func updateContributoor(c *cli.Context, opts *terminal.CommandOpts) error {
	log := opts.Logger()

	configService, err := service.NewConfigService(log, c.GlobalString("config-path"))
	if err != nil {
		if _, ok := err.(*service.ConfigNotFoundError); ok {
			return fmt.Errorf("%s%v%s", terminal.ColorRed, err, terminal.ColorReset)
		}

		return fmt.Errorf("%sError loading config: %v%s", terminal.ColorRed, err, terminal.ColorReset)
	}

	var (
		updateSuccessful bool
		targetVersion    string
		currentVersion   = configService.Get().Version
		github           = service.NewGitHubService("ethpandaops", "contributoor")
	)

	log.WithField("version", currentVersion).Info("Current version")

	defer func() {
		if !updateSuccessful {
			if err := configService.Update(func(cfg *service.ContributoorConfig) {
				cfg.Version = currentVersion
			}); err != nil {
				log.Errorf("Failed to roll back version in config: %v", err)
				return
			}

			if err := configService.Save(); err != nil {
				log.Errorf("Failed to save config after version rollback: %v", err)
			}
		}
	}()

	// Determine target version
	if c.IsSet("version") {
		targetVersion = c.String("version")

		log.WithField("version", targetVersion).Info("Update version provided")

		exists, err := github.VersionExists(targetVersion)
		if err != nil {
			return fmt.Errorf("failed to check version: %w", err)
		}

		if !exists {
			return fmt.Errorf(
				"%sVersion %s not found. Use 'contributoor update' without --version to get the latest version%s",
				terminal.ColorRed,
				targetVersion,
				terminal.ColorReset,
			)
		}
	} else {
		var err error

		targetVersion, err = github.GetLatestVersion()
		if err != nil {
			return fmt.Errorf("failed to get latest version: %w", err)
		}

		log.WithField("version", targetVersion).Info("Latest version detected")
	}

	// Check if update is even needed.
	if targetVersion == configService.Get().Version {
		if c.IsSet("version") {
			log.Infof(
				"%sContributoor is already running version %s%s",
				terminal.ColorGreen,
				targetVersion,
				terminal.ColorReset,
			)
		} else {
			log.Infof(
				"%sContributoor is up to date at version %s%s",
				terminal.ColorGreen,
				targetVersion,
				terminal.ColorReset,
			)
		}

		return nil
	}

	// Update config version.
	if err := configService.Update(func(cfg *service.ContributoorConfig) {
		cfg.Version = targetVersion
	}); err != nil {
		return fmt.Errorf("failed to update config version: %w", err)
	}

	// Save the updated config.
	if err := configService.Save(); err != nil {
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

		if e := dockerService.Update(); e != nil {
			log.Errorf("could not update service: %v", e)
			return e
		}

		// Check if service is running
		running, err := dockerService.IsRunning()
		if err != nil {
			log.Errorf("could not check service status: %v", err)
			return err
		}

		if running {
			if terminal.Confirm("Service is running. Would you like to restart it with the new version?") {
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
			if terminal.Confirm("Service is not running. Would you like to start it?") {
				if err := dockerService.Start(); err != nil {
					return fmt.Errorf("failed to start service: %w", err)
				}
			}
		}

		log.Infof("%sContributoor updated successfully to version %s%s", terminal.ColorGreen, configService.Get().Version, terminal.ColorReset)
	case service.RunMethodBinary:
		binaryService := service.NewBinaryService(log, configService)

		log.WithField("version", configService.Get().Version).Info("Updating Contributoor")

		// Check if service is running
		running, err := binaryService.IsRunning()
		if err != nil {
			log.Errorf("could not check service status: %v", err)
			return err
		}

		if running {
			if terminal.Confirm("Service is running. In order to update, it must be stopped. Would you like to stop it?") {
				if err := binaryService.Stop(); err != nil {
					return fmt.Errorf("failed to stop service: %w", err)
				}
			} else {
				log.Error("Update process was cancelled")

				return nil
			}
		}

		if err := binaryService.Update(); err != nil {
			log.Errorf("could not update service: %v", err)

			return err
		}

		if err := configService.Save(); err != nil {
			log.Errorf("could not save updated config: %v", err)

			return err
		}

		// If it was running, start it again.
		if running {
			if err := binaryService.Start(); err != nil {
				return fmt.Errorf("failed to start service: %w", err)
			}
		}
	}

	updateSuccessful = true

	return nil
}
