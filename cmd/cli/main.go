package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/commands/config"
	"github.com/ethpandaops/contributoor-installer/cmd/cli/commands/install"
	"github.com/ethpandaops/contributoor-installer/cmd/cli/commands/logs"
	"github.com/ethpandaops/contributoor-installer/cmd/cli/commands/restart"
	"github.com/ethpandaops/contributoor-installer/cmd/cli/commands/start"
	"github.com/ethpandaops/contributoor-installer/cmd/cli/commands/status"
	"github.com/ethpandaops/contributoor-installer/cmd/cli/commands/stop"
	"github.com/ethpandaops/contributoor-installer/cmd/cli/commands/update"
	"github.com/ethpandaops/contributoor-installer/cmd/cli/options"
	"github.com/ethpandaops/contributoor-installer/internal/installer"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	installerCfg := installer.NewConfig()

	logLevel, err := logrus.ParseLevel(installerCfg.LogLevel)
	if err != nil {
		logLevel = logrus.InfoLevel
	}

	log := logrus.New()
	log.SetLevel(logLevel)
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		DisableColors: false,
	})

	// Set up log rotation for CLI logs.
	// TODO(@matty): Move this to install.sh?
	logDir := filepath.Join(os.Getenv("HOME"), ".contributoor", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Handle cleanup on exit.
	go func() {
		<-sigChan
		log.Info("Received exit signal")
		os.Exit(0)
	}()

	cli.AppHelpTemplate = tui.AppHelpTemplate
	app := cli.NewApp()
	app.Name = "contributoor"
	app.Usage = "Xatu Contributoor CLI"
	app.Copyright = "(c) 2024 ethPandaOps"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "config-path, c",
			Usage: "Contributoor config asset `path`",
			Value: "~/.contributoor",
		},
		&cli.BoolFlag{
			Name:  "non-interactive",
			Usage: "Skip all interactive prompts and use default values",
		},
		&cli.BoolFlag{
			Name:  "release, r",
			Usage: "Print release and exit",
		},
	}

	app.Before = func(c *cli.Context) error {
		if c.Bool("release") {
			fmt.Printf("%s\n", installer.Release)
			os.Exit(0)
		}

		return nil
	}

	install.RegisterCommands(app, options.NewCommandOpts(
		options.WithName("install"),
		options.WithLogger(log),
	))

	start.RegisterCommands(app, options.NewCommandOpts(
		options.WithName("start"),
		options.WithLogger(log),
		options.WithInstallerConfig(installerCfg),
	))

	stop.RegisterCommands(app, options.NewCommandOpts(
		options.WithName("stop"),
		options.WithLogger(log),
		options.WithInstallerConfig(installerCfg),
	))

	restart.RegisterCommands(app, options.NewCommandOpts(
		options.WithName("restart"),
		options.WithLogger(log),
		options.WithInstallerConfig(installerCfg),
	))

	status.RegisterCommands(app, options.NewCommandOpts(
		options.WithName("status"),
		options.WithLogger(log),
		options.WithInstallerConfig(installerCfg),
	))

	update.RegisterCommands(app, options.NewCommandOpts(
		options.WithName("update"),
		options.WithLogger(log),
		options.WithInstallerConfig(installerCfg),
	))

	config.RegisterCommands(app, options.NewCommandOpts(
		options.WithName("config"),
		options.WithLogger(log),
	))

	logs.RegisterCommands(app, options.NewCommandOpts(
		options.WithName("logs"),
		options.WithLogger(log),
	))

	// Handle normal exit.
	app.After = func(c *cli.Context) error {
		return nil
	}

	fmt.Println("")

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
	}

	fmt.Println("")
}
