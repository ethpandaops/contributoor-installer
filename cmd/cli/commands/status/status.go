package status

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/service"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func RegisterCommands(app *cli.App, opts *options.CommandOpts) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:      opts.Name(),
		Aliases:   opts.Aliases(),
		Usage:     "Show Contributoor status",
		UsageText: "contributoor status [options]",
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

	// Check version and show upgrade warning if needed.
	current, latest, needsUpdate, err := sidecar.CheckVersion(runner, github, cfg.Version)
	if err == nil && needsUpdate {
		tui.UpgradeWarning(current, latest)
	}

	// Check if running.
	running, err := runner.IsRunning()
	if err != nil {
		return fmt.Errorf("failed to check status: %w", err)
	}

	// Get the underlying status from the sidecar.
	status, err := runner.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	// Print status information.
	fmt.Printf("%sContributoor Status%s\n", tui.TerminalColorLightBlue, tui.TerminalColorReset)
	fmt.Printf("%-20s: %s\n", "Version", current)
	fmt.Printf("%-20s: %s\n", "Run Method", cfg.RunMethod)
	fmt.Printf("%-20s: %s\n", "Beacon Node", cfg.BeaconNodeAddress)
	fmt.Printf("%-20s: %s\n", "Config Path", sidecarCfg.GetConfigPath())

	if cfg.OutputServer != nil {
		fmt.Printf("%-20s: %s\n", "Output Server", cfg.OutputServer.Address)
	}

	fmt.Printf(
		"%-20s: %v\n", "Opt-in Attestations",
		cfg.AttestationSubnetCheck != nil && cfg.AttestationSubnetCheck.Enabled,
	)

	// Print running status with color.
	statusColor := tui.TerminalColorRed
	statusText := cases.Title(language.English).String(status)

	if running {
		statusColor = tui.TerminalColorGreen
	}

	fmt.Printf("%-20s: %s%s%s\n", "Status", statusColor, statusText, tui.TerminalColorReset)

	// Print beacon node information if configured.
	if cfg.BeaconNodeAddress != "" {
		printBeaconNodeInfo(c.Context, log, cfg.BeaconNodeAddress)
	}

	return nil
}

func printBeaconNodeInfo(ctx context.Context, log *logrus.Logger, addresses string) {
	nodes := strings.Split(addresses, ",")

	for i, address := range nodes {
		address = strings.TrimSpace(address)
		if address == "" {
			continue
		}

		// Skip status check for non-localhost addresses (e.g., Docker network hostnames).
		// These are not reachable from the host machine where the CLI runs.
		if !isLocalhostAddress(address) {
			continue
		}

		beaconSvc := service.NewBeaconService(log, address)
		info := beaconSvc.GetBeaconInfo(ctx)

		fmt.Println()

		// Show header with node number if multiple nodes.
		if len(nodes) > 1 {
			fmt.Printf("%sBeacon Node %d Status%s (%s)\n",
				tui.TerminalColorLightBlue, i+1, tui.TerminalColorReset, address)
		} else {
			fmt.Printf("%sBeacon Node Status%s\n", tui.TerminalColorLightBlue, tui.TerminalColorReset)
		}

		// Show error if we couldn't connect.
		if info.Error != nil {
			fmt.Printf("%-20s: %s%s%s\n", "Status", tui.TerminalColorRed, "Unreachable", tui.TerminalColorReset)
			fmt.Printf("%-20s: %s\n", "Error", info.Error.Error())

			continue
		}

		// Network.
		if info.Network != "" {
			fmt.Printf("%-20s: %s\n", "Network", info.Network)
		}

		// Health status.
		if info.Health != nil {
			healthColor := tui.TerminalColorGreen
			healthText := "Healthy"

			if info.Health.IsSyncing {
				healthColor = tui.TerminalColorYellow
				healthText = "Syncing"
			} else if !info.Health.IsHealthy {
				healthColor = tui.TerminalColorRed
				healthText = "Unhealthy"
			}

			fmt.Printf("%-20s: %s%s%s\n", "Health", healthColor, healthText, tui.TerminalColorReset)
		}

		// Sync status.
		if info.Sync != nil {
			syncColor := tui.TerminalColorGreen
			syncText := "Synced"

			if info.Sync.IsSyncing {
				syncColor = tui.TerminalColorYellow
				syncText = fmt.Sprintf("Syncing (head: %s, distance: %s)", info.Sync.HeadSlot, info.Sync.SyncDistance)
			}

			if info.Sync.ELOffline {
				syncColor = tui.TerminalColorRed
				syncText = "Execution Layer Offline"
			}

			fmt.Printf("%-20s: %s%s%s\n", "Sync Status", syncColor, syncText, tui.TerminalColorReset)
		}

		// Peer ID (truncated for readability).
		if info.Identity != nil && info.Identity.PeerID != "" {
			peerID := info.Identity.PeerID
			if len(peerID) > 20 {
				peerID = peerID[:10] + "..." + peerID[len(peerID)-10:]
			}

			fmt.Printf("%-20s: %s\n", "Peer ID", peerID)
		}
	}
}

// isLocalhostAddress checks if the address points to localhost.
// Non-localhost addresses (e.g., Docker network hostnames) are not reachable from the host.
func isLocalhostAddress(address string) bool {
	host := strings.TrimPrefix(strings.TrimPrefix(address, "http://"), "https://")
	host = strings.Split(host, ":")[0]

	return strings.HasPrefix(host, "127.0.0.1") || strings.HasPrefix(host, "localhost")
}
