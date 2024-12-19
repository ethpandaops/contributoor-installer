package update

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/installer"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

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
			log := opts.Logger()

			installerCfg := installer.NewConfig()

			sidecarCfg, err := sidecar.NewConfigService(log, c.GlobalString("config-path"))
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			dockerSidecar, err := sidecar.NewDockerSidecar(log, sidecarCfg, installerCfg)
			if err != nil {
				return fmt.Errorf("error creating docker sidecar service: %w", err)
			}

			binarySidecar, err := sidecar.NewBinarySidecar(log, sidecarCfg, installerCfg)
			if err != nil {
				return fmt.Errorf("error creating binary sidecar service: %w", err)
			}

			githubService, err := service.NewGitHubService(log, installerCfg)
			if err != nil {
				return fmt.Errorf("error creating github service: %w", err)
			}

			return updateContributoor(c, log, sidecarCfg, dockerSidecar, binarySidecar, githubService)
		},
	})
}

func updateContributoor(
	c *cli.Context,
	log *logrus.Logger,
	sidecarCfg sidecar.ConfigManager,
	docker sidecar.DockerSidecar,
	binary sidecar.BinarySidecar,
	github service.GitHubService,
) error {
	var (
		success        bool
		targetVersion  string
		cfg            = sidecarCfg.Get()
		currentVersion = cfg.Version
	)

	log.WithField("version", currentVersion).Info("Current version")

	defer func() {
		if !success {
			if err := rollbackVersion(log, sidecarCfg, currentVersion); err != nil {
				log.Error(err)
			}
		}
	}()

	// Determine target version.
	targetVersion, err := determineTargetVersion(c, github)
	if err != nil {
		// Flag as success, there's nothing to update on rollback if we fail to determine the target version.
		success = true

		return err
	}

	// Check if update is needed.
	if targetVersion == currentVersion {
		// Flag as success, there's nothing to update.
		success = true

		logUpdateStatus(log, c.IsSet("version"), targetVersion)

		return nil
	}

	// Update config version.
	if uerr := updateConfigVersion(sidecarCfg, targetVersion); uerr != nil {
		return uerr
	}

	// Refresh our config state, given it was updated above.
	cfg = sidecarCfg.Get()

	log.WithField("version", cfg.Version).Info("Updating Contributoor")

	// Update the sidecar.
	success, err = updateSidecar(log, cfg, docker, binary)
	if err != nil {
		return err
	}

	log.Infof(
		"%sContributoor updated successfully to version %s%s",
		tui.TerminalColorGreen,
		cfg.Version,
		tui.TerminalColorReset,
	)

	return nil
}

func updateSidecar(log *logrus.Logger, cfg *sidecar.Config, docker sidecar.DockerSidecar, binary sidecar.BinarySidecar) (bool, error) {
	switch cfg.RunMethod {
	case sidecar.RunMethodDocker:
		return updateDocker(log, cfg, docker)
	case sidecar.RunMethodBinary:
		return updateBinary(log, cfg, binary)
	default:
		return false, fmt.Errorf("invalid sidecar run method: %s", cfg.RunMethod)
	}
}

func updateBinary(log *logrus.Logger, cfg *sidecar.Config, binary sidecar.BinarySidecar) (bool, error) {
	// Check if sidecar is currently running.
	running, err := binary.IsRunning()
	if err != nil {
		log.Errorf("could not check sidecar status: %v", err)

		return false, err
	}

	// If the sidecar is running, we need to stop it before we can update the binary.
	if running {
		if tui.Confirm("Contributoor is running. In order to update, it must be stopped. Would you like to stop it?") {
			if err := binary.Stop(); err != nil {
				return false, fmt.Errorf("failed to stop sidecar: %w", err)
			}
		} else {
			log.Error("update process was cancelled")

			return false, nil
		}
	}

	if err := binary.Update(); err != nil {
		log.Errorf("could not update sidecar: %v", err)

		return false, err
	}

	// If it was running, start it again for them.
	if running {
		if err := binary.Start(); err != nil {
			return true, fmt.Errorf("failed to start sidecar: %w", err)
		}
	}

	return true, nil
}

func updateDocker(log *logrus.Logger, cfg *sidecar.Config, docker sidecar.DockerSidecar) (bool, error) {
	if err := docker.Update(); err != nil {
		log.Errorf("could not update service: %v", err)

		return false, err
	}

	// Check if service is currently running.
	running, err := docker.IsRunning()
	if err != nil {
		log.Errorf("could not check sidecar status: %v", err)

		return true, err
	}

	// If the service is running, we need to restart it with the new version.
	if running {
		if tui.Confirm("Contributoor is running. Would you like to restart it with the new version?") {
			if err := docker.Stop(); err != nil {
				return true, fmt.Errorf("failed to stop sidecar: %w", err)
			}

			if err := docker.Start(); err != nil {
				return true, fmt.Errorf("failed to start sidecar: %w", err)
			}
		} else {
			log.Info("service will continue running with the previous version until next restart")
		}
	} else {
		if tui.Confirm("Contributoor is not running. Would you like to start it?") {
			if err := docker.Start(); err != nil {
				return true, fmt.Errorf("failed to start service: %w", err)
			}
		}
	}

	return true, nil
}

func determineTargetVersion(c *cli.Context, github service.GitHubService) (string, error) {
	if c.IsSet("version") {
		version := c.String("version")

		exists, err := github.VersionExists(version)
		if err != nil {
			return "", fmt.Errorf("failed to check version: %w", err)
		}

		if !exists {
			return "", fmt.Errorf(
				"%sversion %s not found. Use 'contributoor update' without --version to get the latest version%s",
				tui.TerminalColorRed,
				version,
				tui.TerminalColorReset,
			)
		}

		return version, nil
	}

	version, err := github.GetLatestVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get latest version: %w", err)
	}

	return version, nil
}

func updateConfigVersion(config sidecar.ConfigManager, version string) error {
	if err := config.Update(func(cfg *sidecar.Config) {
		cfg.Version = version
	}); err != nil {
		return fmt.Errorf("failed to update sidecar config version: %w", err)
	}

	if err := config.Save(); err != nil {
		return fmt.Errorf("could not save updated sidecar config: %w", err)
	}

	return nil
}

func rollbackVersion(log *logrus.Logger, config sidecar.ConfigManager, version string) error {
	if err := config.Update(func(cfg *sidecar.Config) {
		cfg.Version = version
	}); err != nil {
		return fmt.Errorf("failed to roll back version in sidecar config: %w", err)
	}

	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to save sidecar config after version rollback: %w", err)
	}

	return nil
}

func logUpdateStatus(log *logrus.Logger, isVersionSet bool, version string) {
	if isVersionSet {
		log.Infof(
			"%scontributoor is already running version %s%s",
			tui.TerminalColorGreen,
			version,
			tui.TerminalColorReset,
		)
	} else {
		log.Infof(
			"%scontributoor is up to date at version %s%s",
			tui.TerminalColorGreen,
			version,
			tui.TerminalColorReset,
		)
	}
}
