package update

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func RegisterCommands(app *cli.App, opts *options.CommandOpts) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:      opts.Name(),
		Aliases:   opts.Aliases(),
		Usage:     "Update Contributoor to the latest version",
		UsageText: "contributoor update [options]",
		Action: func(c *cli.Context) error {
			var (
				log          = opts.Logger()
				installerCfg = opts.InstallerConfig()
			)

			sidecarCfg, err := sidecar.NewConfigService(log, c.String("config-path"))
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
		success bool
		cfg     = sidecarCfg.Get()
		err     error
		runner  sidecar.SidecarRunner
	)

	fmt.Printf("%sUpdating Contributoor Version%s\n", tui.TerminalColorLightBlue, tui.TerminalColorReset)

	switch cfg.RunMethod {
	case config.RunMethod_RUN_METHOD_DOCKER:
		runner = docker
	case config.RunMethod_RUN_METHOD_SYSTEMD:
		runner = systemd
	case config.RunMethod_RUN_METHOD_BINARY:
		runner = binary
	default:
		return fmt.Errorf("invalid sidecar run method: %s", cfg.RunMethod)
	}

	current, latest, needsUpdate, err := sidecar.CheckVersion(runner, github, cfg.Version)
	if err != nil {
		return err
	}

	defer func() {
		if !success {
			if rollbackErr := rollbackVersion(sidecarCfg, current); rollbackErr != nil {
				log.Error(rollbackErr)
			}
		}
	}()

	fmt.Printf("%-20s: %s\n", "Current Version", current)
	fmt.Printf("%-20s: %s\n", "Latest Version", latest)

	// Check if update is needed.
	if !needsUpdate {
		success = true

		fmt.Printf(
			"%sContributoor is up to date at version %s%s\n",
			tui.TerminalColorGreen,
			latest,
			tui.TerminalColorReset,
		)

		return nil
	}

	// Update config version.
	if configErr := updateConfigVersion(sidecarCfg, latest); configErr != nil {
		return configErr
	}

	// Refresh our config state, given it was updated above.
	cfg = sidecarCfg.Get()

	// Update the sidecar.
	success, err = updateSidecar(c, log, cfg, docker, systemd, binary)
	if err != nil {
		return err
	}

	return nil
}

func updateSidecar(
	c *cli.Context,
	log *logrus.Logger,
	cfg *config.Config,
	docker sidecar.DockerSidecar,
	systemd sidecar.SystemdSidecar,
	binary sidecar.BinarySidecar,
) (bool, error) {
	switch cfg.RunMethod {
	case config.RunMethod_RUN_METHOD_DOCKER:
		return updateDocker(c, log, cfg, docker)
	case config.RunMethod_RUN_METHOD_SYSTEMD:
		return updateSystemd(c, log, cfg, systemd)
	case config.RunMethod_RUN_METHOD_BINARY:
		return updateBinary(c, log, cfg, binary)
	default:
		return false, fmt.Errorf("invalid sidecar run method: %s", cfg.RunMethod)
	}
}

func updateSystemd(c *cli.Context, log *logrus.Logger, cfg *config.Config, systemd sidecar.SystemdSidecar) (bool, error) {
	// Check if sidecar is currently running.
	running, err := systemd.IsRunning()
	if err != nil {
		return false, err
	}

	// If the sidecar is running, we need to stop it before we can update the binary.
	if running {
		if err := systemd.Stop(); err != nil {
			return false, fmt.Errorf("failed to stop sidecar: %w", err)
		}
	}

	if err := systemd.Update(); err != nil {
		return false, err
	}

	fmt.Printf("%sContributoor updated successfully to version %s%s\n", tui.TerminalColorGreen, cfg.Version, tui.TerminalColorReset)

	// If it was running, start it again for them.
	if running {
		if c.Bool("non-interactive") || tui.Confirm("Would you like to restart Contributoor with the new version?") {
			if err := systemd.Start(); err != nil {
				return true, fmt.Errorf("failed to start sidecar: %w", err)
			}
		} else {
			fmt.Printf("%sContributoor will remain stopped until manually started%s\n", tui.TerminalColorYellow, tui.TerminalColorReset)
		}
	}

	return true, nil
}

func updateBinary(c *cli.Context, log *logrus.Logger, cfg *config.Config, binary sidecar.BinarySidecar) (bool, error) {
	// Check if sidecar is currently running.
	running, err := binary.IsRunning()
	if err != nil {
		log.Errorf("could not check sidecar status: %v", err)

		return false, err
	}

	// If the sidecar is running, we need to stop it before we can update the binary.
	if running {
		if c.Bool("non-interactive") || tui.Confirm("Contributoor is running. In order to update, it must be stopped. Would you like to stop it?") {
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

func updateDocker(c *cli.Context, log *logrus.Logger, cfg *config.Config, docker sidecar.DockerSidecar) (bool, error) {
	if err := docker.Update(); err != nil {
		log.Errorf("could not update service: %v", err)

		return false, err
	}

	fmt.Printf("%sContributoor updated successfully to version %s%s", tui.TerminalColorGreen, cfg.Version, tui.TerminalColorReset)

	// Check if service is currently running.
	running, err := docker.IsRunning()
	if err != nil {
		log.Errorf("could not check sidecar status: %v", err)

		return true, err
	}

	fmt.Printf("\n")

	// If the service is running, we need to restart it with the new version.
	if running {
		if c.Bool("non-interactive") || tui.Confirm("Contributoor is running. Would you like to restart it with the new version?") {
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
		if c.Bool("non-interactive") || tui.Confirm("Contributoor is not running. Would you like to start it?") {
			if err := docker.Start(); err != nil {
				return true, fmt.Errorf("failed to start service: %w", err)
			}
		}
	}

	return true, nil
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
