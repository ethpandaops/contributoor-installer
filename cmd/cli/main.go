package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/commands/install"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/commands/start"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/commands/stop"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/commands/update"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal/display"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/utils"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// Run
func main() {
	// Add logo to application help template
	cli.AppHelpTemplate = fmt.Sprintf(`%s
Authored by the ethPandaOps team

%s`, display.Logo, cli.AppHelpTemplate)
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Handle cleanup on exit
	go func() {
		<-sigChan
		logger.Info("Received exit signal")
		cleanup(logger)
		os.Exit(0)
	}()

	app := cli.NewApp()

	// Set application info
	app.Name = "contributoor"
	app.Usage = "Xatu Contributoor CLI"
	app.Copyright = "(c) 2024 ethPandaOps"

	// Initialize app metadata
	app.Metadata = make(map[string]interface{})
	app.Metadata["logger"] = logger // Attach the logger

	// Set application flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config-path, c",
			Usage: "Contributoor config asset `path`",
			Value: "~/.contributoor",
		},
	}

	// Register commands
	install.RegisterCommands(app, utils.NewCommandOpts(
		utils.WithName("install"),
		utils.WithLogger(logger),
		utils.WithAliases([]string{"i"}),
	))

	start.RegisterCommands(app, utils.NewCommandOpts(
		utils.WithName("start"),
		utils.WithLogger(logger),
	))

	stop.RegisterCommands(app, utils.NewCommandOpts(
		utils.WithName("stop"),
		utils.WithLogger(logger),
	))

	update.RegisterCommands(app, utils.NewCommandOpts(
		utils.WithName("update"),
		utils.WithLogger(logger),
	))

	// Handle normal exit
	app.After = func(c *cli.Context) error {
		cleanup(logger)
		return nil
	}

	// Run application
	fmt.Println("")
	if err := app.Run(os.Args); err != nil {
		logger.Error(err)
	}
	fmt.Println("")

}

func cleanup(log *logrus.Logger) {
	// Add any cleanup tasks here
	// - Save state
	// - Close connections
	// - Remove temp files
	// etc.
}
