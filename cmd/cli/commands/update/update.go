package update

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
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
			var (
				log          = opts.Logger()
				installerCfg = opts.InstallerConfig()
			)

			sidecarCfg, err := sidecar.NewConfigService(log, c.GlobalString("config-path"))
			if err != nil {
				return fmt.Errorf("%s%v%s", tui.TerminalColorRed, err, tui.TerminalColorReset)
			}

			dockerSidecar, err := sidecar.NewDockerSidecar(log, sidecarCfg, installerCfg)
			if err != nil {
				return fmt.Errorf("error creating docker sidecar service: %w", err)
			}

			systemdSidecar, err := sidecar.NewSystemdSidecar(log, sidecarCfg, installerCfg)
			if err != nil {
				return fmt.Errorf("error creating systemd sidecar service: %w", err)
			}

			binarySidecar, err := sidecar.NewBinarySidecar(log, sidecarCfg, installerCfg)
			if err != nil {
				return fmt.Errorf("error creating binary sidecar service: %w", err)
			}

			githubService, err := service.NewGitHubService(log, installerCfg)
			if err != nil {
				return fmt.Errorf("error creating github service: %w", err)
			}

			return updateContributoor(c, log, sidecarCfg, dockerSidecar, systemdSidecar, binarySidecar, githubService)
		},
	})
}

func updateContributoor(
	c *cli.Context,
	log *logrus.Logger,
	sidecarCfg sidecar.ConfigManager,
	docker sidecar.DockerSidecar,
	systemd sidecar.SystemdSidecar,
	binary sidecar.BinarySidecar,
	github service.GitHubService,
) error {
	var (
		success        bool
		targetVersion  string
		cfg            = sidecarCfg.Get()
		currentVersion = cfg.Version
	)

	fmt.Printf("%sUpdating Contributoor Version%s\n", tui.TerminalColorLightBlue, tui.TerminalColorReset)
	fmt.Printf("%-20s: %s\n", "Current Version", cfg.Version)

	defer func() {
		if !success {
			if err := rollbackVersion(sidecarCfg, currentVersion); err != nil {
				log.Error(err)
			}
		}
	}()

	// Determine target version.
	targetVersion, err := determineTargetVersion(c, github)
	if err != nil || targetVersion == "" {
		// Flag as success, there's nothing to update on rollback if we fail to determine the target version.
		success = true

		return err
	}

	fmt.Printf("%-20s: %s\n", "Latest Version", targetVersion)

	// Check if update is needed.
	if targetVersion == currentVersion {
		// Flag as success, there's nothing to update.
		success = true

		printUpdateStatus(c.IsSet("version"), targetVersion)

		return nil
	}

	// Update config version.
	if uerr := updateConfigVersion(sidecarCfg, targetVersion); uerr != nil {
		return uerr
	}

	// Refresh our config state, given it was updated above.
	cfg = sidecarCfg.Get()

	// Update the sidecar.
	success, err = updateSidecar(log, cfg, docker, systemd, binary)
	if err != nil {
		return err
	}

	return nil
}

func updateSidecar(log *logrus.Logger, cfg *config.Config, docker sidecar.DockerSidecar, systemd sidecar.SystemdSidecar, binary sidecar.BinarySidecar) (bool, error) {
	switch cfg.RunMethod {
	case config.RunMethod_RUN_METHOD_DOCKER:
		return updateDocker(log, cfg, docker)
	case config.RunMethod_RUN_METHOD_SYSTEMD:
		return updateSystemd(log, cfg, systemd)
	case config.RunMethod_RUN_METHOD_BINARY:
		return updateBinary(log, cfg, binary)
	default:
		return false, fmt.Errorf("invalid sidecar run method: %s", cfg.RunMethod)
	}
}

func updateSystemd(log *logrus.Logger, cfg *config.Config, systemd sidecar.SystemdSidecar) (bool, error) {
	// Check if sidecar is currently running.
	running, err := systemd.IsRunning()
	if err != nil {
		log.Errorf("could not check sidecar status: %v", err)

		return false, err
	}

	// If the sidecar is running, we need to stop it before we can update the binary.
	if running {
		if err := systemd.Stop(); err != nil {
			return false, fmt.Errorf("failed to stop sidecar: %w", err)
		}
	}

	if err := systemd.Update(); err != nil {
		log.Errorf("could not update sidecar: %v", err)

		return false, err
	}

	fmt.Printf("%sContributoor updated successfully to version %s%s\n", tui.TerminalColorGreen, cfg.Version, tui.TerminalColorReset)

	// If it was running, start it again for them.
	if running {
		if err := systemd.Start(); err != nil {
			return true, fmt.Errorf("failed to start sidecar: %w", err)
		}
	}

	return true, nil
}

func updateBinary(log *logrus.Logger, cfg *config.Config, binary sidecar.BinarySidecar) (bool, error) {
	// Check if sidecar is currently running.
	running, err := binary.IsRunning()
	if err != nil {
		log.Errorf("could not check sidecar status: %v", err)

		return false, err
	}

	// If the sidecar is running, we need to stop it before we can update the binary.
	if running {
		fmt.Printf("\n")

		if tui.Confirm("Contributoor is running. In order to update, it must be stopped. Would you like to stop it?") {
			if err := binary.Stop(); err != nil {
				return false, fmt.Errorf("failed to stop sidecar: %w", err)
			}
		} else {
			fmt.Printf("%sUpdate process was cancelled%s\n", tui.TerminalColorRed, tui.TerminalColorReset)

			return false, nil
		}
	}

	if err := binary.Update(); err != nil {
		log.Errorf("could not update sidecar: %v", err)

		return false, err
	}

	fmt.Printf("%sContributoor updated successfully to version %s%s\n", tui.TerminalColorGreen, cfg.Version, tui.TerminalColorReset)

	// If it was running, start it again for them.
	if running {
		if err := binary.Start(); err != nil {
			return true, fmt.Errorf("failed to start sidecar: %w", err)
		}
	}

	return true, nil
}

func updateDocker(log *logrus.Logger, cfg *config.Config, docker sidecar.DockerSidecar) (bool, error) {
	if err := docker.Update(); err != nil {
		log.Errorf("could not update service: %v", err)

		return false, err
	}

	fmt.Printf("%sContributoor updated successfully to version %s%s\n", tui.TerminalColorGreen, cfg.Version, tui.TerminalColorReset)

	// Check if service is currently running.
	running, err := docker.IsRunning()
	if err != nil {
		log.Errorf("could not check sidecar status: %v", err)

		return true, err
	}

	fmt.Printf("\n")

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
			fmt.Printf("%sContributoor will continue running with the previous version until next restart%s\n", tui.TerminalColorYellow, tui.TerminalColorReset)
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
			fmt.Printf(
				"%sVersion %s not found. Use 'contributoor update' without --version to get the latest version%s\n",
				tui.TerminalColorRed,
				version,
				tui.TerminalColorReset,
			)

			return "", nil
		}

		return version, nil
	}

	version, err := github.GetLatestVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get latest version: %w", err)
	}

	return version, nil
}

func updateConfigVersion(sidecarCfg sidecar.ConfigManager, version string) error {
	if err := sidecarCfg.Update(func(cfg *config.Config) {
		cfg.Version = version
	}); err != nil {
		return fmt.Errorf("failed to update sidecar config version: %w", err)
	}

	if err := sidecarCfg.Save(); err != nil {
		return fmt.Errorf("could not save updated sidecar config: %w", err)
	}

	return nil
}

func rollbackVersion(sidecarCfg sidecar.ConfigManager, version string) error {
	if err := sidecarCfg.Update(func(cfg *config.Config) {
		cfg.Version = version
	}); err != nil {
		return fmt.Errorf("failed to roll back version in sidecar config: %w", err)
	}

	if err := sidecarCfg.Save(); err != nil {
		return fmt.Errorf("failed to save sidecar config after version rollback: %w", err)
	}

	return nil
}

func printUpdateStatus(isVersionSet bool, version string) {
	if isVersionSet {
		fmt.Printf(
			"%sContributoor is already running version %s%s\n",
			tui.TerminalColorGreen,
			version,
			tui.TerminalColorReset,
		)
	} else {
		fmt.Printf(
			"%sContributoor is up to date at version %s%s\n",
			tui.TerminalColorGreen,
			version,
			tui.TerminalColorReset,
		)
	}
}
