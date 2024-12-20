package status

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
		Usage:     "Show Contributoor status",
		UsageText: "contributoor status [options]",
		Action: func(c *cli.Context) error {
			var (
				log          = opts.Logger()
				installerCfg = opts.InstallerConfig()
			)

			sidecarCfg, err := sidecar.NewConfigService(log, c.GlobalString("config-path"))
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
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

			return showStatus(c, log, sidecarCfg, dockerSidecar, systemdSidecar, binarySidecar, githubService)
		},
	})
}

func showStatus(
	c *cli.Context,
	log *logrus.Logger,
	sidecarCfg sidecar.ConfigManager,
	docker sidecar.DockerSidecar,
	systemd sidecar.SystemdSidecar,
	binary sidecar.BinarySidecar,
	github service.GitHubService,
) error {
	var (
		runner sidecar.SidecarRunner
		cfg    = sidecarCfg.Get()
	)

	// Determine which runner to use.
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

	// Check if running.
	running, err := runner.IsRunning()
	if err != nil {
		return fmt.Errorf("failed to check status: %w", err)
	}

	// Check if there's a newer version available.
	var latestVersionLine string

	latestVersion, err := github.GetLatestVersion()
	if err == nil && cfg.Version != latestVersion {
		latestVersionLine = fmt.Sprintf("%-20s: %s%s%s", "Latest Version", tui.TerminalColorYellow, latestVersion, tui.TerminalColorReset)
	}

	// Print status information.
	fmt.Printf("%sContributoor Status%s\n", tui.TerminalColorLightBlue, tui.TerminalColorReset)
	fmt.Printf("%-20s: %s\n", "Version", cfg.Version)

	if latestVersionLine != "" {
		fmt.Printf("%s\n", latestVersionLine)
	}

	fmt.Printf("%-20s: %s\n", "Run Method", cfg.RunMethod)
	fmt.Printf("%-20s: %s\n", "Network", cfg.NetworkName)
	fmt.Printf("%-20s: %s\n", "Beacon Node", cfg.BeaconNodeAddress)
	fmt.Printf("%-20s: %s\n", "Config Path", sidecarCfg.GetConfigPath())

	if cfg.OutputServer != nil {
		fmt.Printf("%-20s: %s\n", "Output Server", cfg.OutputServer.Address)
	}

	// Print running status with color
	statusColor := tui.TerminalColorRed
	statusText := "Stopped"

	if running {
		statusColor = tui.TerminalColorGreen
		statusText = "Running"
	}

	fmt.Printf("%-20s: %s%s%s\n", "Status", statusColor, statusText, tui.TerminalColorReset)

	return nil
}
