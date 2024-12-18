package update

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
)

// RegisterCommands registers the update command.
func RegisterCommands(app *cli.App, opts *options.CommandOpts) {
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
func updateContributoor(c *cli.Context, opts *options.CommandOpts) error {
	log := opts.Logger()

	configService, err := service.NewConfigService(log, c.GlobalString("config-path"))
	if err != nil {
		return fmt.Errorf("%sError loading config: %v%s", tui.TerminalColorRed, err, tui.TerminalColorReset)
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

	// Determine target version. If we were passed a version, use that.
	// If not, get the latest version from GitHub.
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
				tui.TerminalColorRed,
				targetVersion,
				tui.TerminalColorReset,
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

	// We don't need to update if the target version is the same as the current version.
	if targetVersion == configService.Get().Version {
		if c.IsSet("version") {
			log.Infof(
				"%sContributoor is already running version %s%s",
				tui.TerminalColorGreen,
				targetVersion,
				tui.TerminalColorReset,
			)
		} else {
			log.Infof(
				"%sContributoor is up to date at version %s%s",
				tui.TerminalColorGreen,
				targetVersion,
				tui.TerminalColorReset,
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

	// Update the service via whatever method the user has configured (docker or binary).
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

		// Check if service is currently running.
		running, err := dockerService.IsRunning()
		if err != nil {
			log.Errorf("could not check service status: %v", err)

			return err
		}

		// If the service is running, we need to restart it with the new version.
		// Given its docker, we can ask the user if they want to restart it. Otherwise,
		// we'll just let it run with the previous version until next restart.
		if running {
			if tui.Confirm("Service is running. Would you like to restart it with the new version?") {
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
			if tui.Confirm("Service is not running. Would you like to start it?") {
				if err := dockerService.Start(); err != nil {
					return fmt.Errorf("failed to start service: %w", err)
				}
			}
		}

		log.Infof("%sContributoor updated successfully to version %s%s", tui.TerminalColorGreen, configService.Get().Version, tui.TerminalColorReset)
	case service.RunMethodBinary:
		binaryService := service.NewBinaryService(log, configService)

		log.WithField("version", configService.Get().Version).Info("Updating Contributoor")

		// Check if service is currently running.
		running, err := binaryService.IsRunning()
		if err != nil {
			log.Errorf("could not check service status: %v", err)

			return err
		}

		// If the service is running, we need to stop it before we can update the binary.
		if running {
			if tui.Confirm("Service is running. In order to update, it must be stopped. Would you like to stop it?") {
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

		// If it was running, start it again for them.
		if running {
			if err := binaryService.Start(); err != nil {
				return fmt.Errorf("failed to start service: %w", err)
			}
		}
	}

	updateSuccessful = true

	return nil
}
