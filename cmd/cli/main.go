package main

import (
	"fmt"
	"os"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/commands/install"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/commands/run"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal/display"
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
	install.RegisterCommands(app, "install", []string{"i"})
	run.RegisterCommands(app, "run", []string{"r"})

	// Run application
	fmt.Println("")
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
	fmt.Println("")

}
