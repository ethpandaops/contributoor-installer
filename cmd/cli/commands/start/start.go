package start

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal/service"
	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/utils"
)

func RegisterCommands(app *cli.App, opts *utils.CommandOpts) {
	app.Commands = append(app.Commands, cli.Command{
		Name:      opts.Name(),
		Aliases:   opts.Aliases(),
		Usage:     "Start Contributoor",
		UsageText: "contributoor start [options]",
		Action: func(c *cli.Context) error {
			return startContributoor(c, opts)
		},
	})
}

func startContributoor(c *cli.Context, opts *utils.CommandOpts) error {
	log := opts.Logger()
	configPath := c.GlobalString("config-path")
	path, err := homedir.Expand(configPath)
	if err != nil {
		return fmt.Errorf("%sFailed to expand config path: %w%s", utils.ColorRed, err, utils.ColorReset)
	}

	// Check directory exists
	dirInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("%sYour configured contributoor directory [%s] does not exist. Please run 'contributoor install' first%s", utils.ColorRed, path, utils.ColorReset)
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("%s[%s] is not a directory%s", utils.ColorRed, path, utils.ColorReset)
	}

	// Check config file exists
	configFile := filepath.Join(path, "config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("%sConfig file not found at [%s]. Please run 'contributoor install' first%s", utils.ColorRed, configFile, utils.ColorReset)
	}

	configService, err := service.NewConfigService(log, configFile)
	if err != nil {
		return err
	}

	switch configService.Get().RunMethod {
	case service.RunMethodDocker:
		log.WithField("version", configService.Get().Version).Info("Starting Contributoor")
		dockerService, err := service.NewDockerService(log, configService)
		if err != nil {
			log.Errorf("could not create docker service: %v", err)
			return err
		}

		// Check if already running
		running, err := dockerService.IsRunning()
		if err != nil {
			log.Errorf("could not check service status: %v", err)
			return err
		}
		if running {
			return fmt.Errorf("%sContributoor is already running. Use 'contributoor stop' first if you want to restart it%s", utils.ColorRed, utils.ColorReset)
		}

		if err := dockerService.Start(); err != nil {
			log.Errorf("could not start service: %v", err)
			return err
		}
	case service.RunMethodBinary:
		binaryService := service.NewBinaryService(log, configService)
		if err := binaryService.Start(); err != nil {
			log.Errorf("could not start service: %v", err)
			return err
		}
	}

	return nil
}
