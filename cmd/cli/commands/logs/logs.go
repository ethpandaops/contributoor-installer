package logs

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func RegisterCommands(app *cli.App, opts *options.CommandOpts) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:      "logs",
		Usage:     "View Contributoor logs",
		UsageText: "contributoor logs [options]",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "tail",
				Usage: "Number of lines to show from the end of logs",
				Value: 100,
			},
			&cli.BoolFlag{
				Name:  "follow, f",
				Usage: "Follow log output",
			},
		},
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

			return showLogs(c, log, sidecarCfg, dockerSidecar, systemdSidecar, binarySidecar)
		},
	})
}

func showLogs(
	c *cli.Context,
	log *logrus.Logger,
	sidecarCfg sidecar.ConfigManager,
	docker sidecar.DockerSidecar,
	systemd sidecar.SystemdSidecar,
	binary sidecar.BinarySidecar,
) error {
	var (
		runner sidecar.SidecarRunner
		cfg    = sidecarCfg.Get()
	)

	// Get the appropriate runner based on run method.
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

	return runner.Logs(c.Int("tail"), c.Bool("follow"))
}
