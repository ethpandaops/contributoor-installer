package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/ethpandaops/contributoor-installer/cmd/cli/commands/config"
	"github.com/ethpandaops/contributoor-installer/cmd/cli/commands/install"
	"github.com/ethpandaops/contributoor-installer/cmd/cli/commands/start"
	"github.com/ethpandaops/contributoor-installer/cmd/cli/commands/stop"
	"github.com/ethpandaops/contributoor-installer/cmd/cli/commands/update"
	"github.com/ethpandaops/contributoor-installer/cmd/cli/terminal"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	// Enable color output
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		DisableColors: false,
	})

	// Set up log rotation for CLI logs
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

	cli.AppHelpTemplate = terminal.AppHelpTemplate
	app := cli.NewApp()
	app.Name = "contributoor"
	app.Usage = "Xatu Contributoor CLI"
	app.Copyright = "(c) 2024 ethPandaOps"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config-path, c",
			Usage: "Contributoor config asset `path`",
			Value: "~/.contributoor",
		},
	}

	install.RegisterCommands(app, terminal.NewCommandOpts(
		terminal.WithName("install"),
		terminal.WithLogger(log),
		terminal.WithAliases([]string{"i"}),
	))

	start.RegisterCommands(app, terminal.NewCommandOpts(
		terminal.WithName("start"),
		terminal.WithLogger(log),
	))

	stop.RegisterCommands(app, terminal.NewCommandOpts(
		terminal.WithName("stop"),
		terminal.WithLogger(log),
	))

	update.RegisterCommands(app, terminal.NewCommandOpts(
		terminal.WithName("update"),
		terminal.WithLogger(log),
	))

	config.RegisterCommands(app, terminal.NewCommandOpts(
		terminal.WithName("config"),
		terminal.WithLogger(log),
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
