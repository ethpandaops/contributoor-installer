package stop

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func RegisterCommands(app *cli.App, opts *options.CommandOpts) {
	app.Commands = append(app.Commands, cli.Command{
		Name:      opts.Name(),
		Aliases:   opts.Aliases(),
		Usage:     "Stop Contributoor",
		UsageText: "contributoor stop [options]",
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

			binarySidecar, err := sidecar.NewBinarySidecar(log, sidecarCfg, installerCfg)
			if err != nil {
				return fmt.Errorf("error creating binary sidecar service: %w", err)
			}

			systemdSidecar, err := sidecar.NewSystemdSidecar(log, sidecarCfg, installerCfg)
			if err != nil {
				return fmt.Errorf("error creating systemd sidecar service: %w", err)
			}

			return stopContributoor(c, log, sidecarCfg, dockerSidecar, systemdSidecar, binarySidecar)
		},
	})
}

func stopContributoor(
	c *cli.Context,
	log *logrus.Logger,
	config sidecar.ConfigManager,
	docker sidecar.DockerSidecar,
	systemd sidecar.SystemdSidecar,
	binary sidecar.BinarySidecar,
) error {
	var (
		runner sidecar.SidecarRunner
		cfg    = config.Get()
	)

	fmt.Printf("%sStopping Contributoor%s\n", tui.TerminalColorLightBlue, tui.TerminalColorReset)

	// Stop the sidecar via whatever method the user has configured (docker or binary).
	switch cfg.RunMethod {
	case sidecar.RunMethodDocker:
		runner = docker
	case sidecar.RunMethodSystemd:
		runner = systemd
	case sidecar.RunMethodBinary:
		runner = binary
	default:
		return fmt.Errorf("invalid sidecar run method: %s", cfg.RunMethod)
	}

	// Check if running before attempting to stop.
	running, err := runner.IsRunning()
	if err != nil {
		log.Errorf("could not check sidecar status: %v", err)

		return err
	}

	// If the service is not running, we can just return.
	if !running {
		fmt.Printf("%sContributoor is not running. Use 'contributoor start' to start it%s\n", tui.TerminalColorYellow, tui.TerminalColorReset)

		return nil
	}

	if err := runner.Stop(); err != nil {
		return err
	}

	return nil
}
