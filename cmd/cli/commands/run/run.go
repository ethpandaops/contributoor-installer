package run

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal/service"
)

const (
	colorReset     string = "\033[0m"
	colorBold      string = "\033[1m"
	colorRed       string = "\033[31m"
	colorYellow    string = "\033[33m"
	colorGreen     string = "\033[32m"
	colorLightBlue string = "\033[36m"
	clearLine      string = "\033[2K"
)

func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, cli.Command{
		Name:      name,
		Aliases:   aliases,
		Usage:     "Run Contributoor",
		UsageText: "contributoor run [options]",
		Action: func(c *cli.Context) error {
			return runContributoor(c)
		},
	})
}

func runContributoor(c *cli.Context) error {
	configPath := c.GlobalString("config-path")
	path, err := homedir.Expand(configPath)
	if err != nil {
		return fmt.Errorf("error expanding config path [%s]: %w", configPath, err)
	}

	// Check directory exists
	dirInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("%sYour configured contributoor directory [%s] does not exist. Please run 'contributoor install' first%s", colorRed, path, colorReset)
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("%s[%s] is not a directory%s", colorRed, path, colorReset)
	}

	// Check config file exists
	configFile := filepath.Join(path, "contributoor.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("%sConfig file not found at [%s]. Please run 'contributoor install' first%s", colorRed, configFile, colorReset)
	}

	logger := c.App.Metadata["logger"].(*logrus.Logger)

	cfg, err := internal.LoadConfig(configFile)
	if err != nil {
		return err
	}

	switch cfg.RunMethod {
	case internal.RunMethodDocker:
		dockerService := service.NewDockerService(logger, cfg)
		if err := dockerService.Start(); err != nil {
			logger.Errorf("could not start service: %v", err)
			return err
		}
	case internal.RunMethodBinary:
		binaryService := service.NewBinaryService(logger, cfg)
		if err := binaryService.Start(); err != nil {
			logger.Errorf("could not start service: %v", err)
			return err
		}
	}

	return nil
}
