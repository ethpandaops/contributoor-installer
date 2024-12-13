package stop

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/terminal"
	"github.com/ethpandaops/contributoor-installer-test/internal/service"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
)

func RegisterCommands(app *cli.App, opts *terminal.CommandOpts) {
	app.Commands = append(app.Commands, cli.Command{
		Name:      opts.Name(),
		Aliases:   opts.Aliases(),
		Usage:     "Stop Contributoor",
		UsageText: "contributoor stop [options]",
		Action: func(c *cli.Context) error {
			return stopContributoor(c, opts)
		},
	})
}

func stopContributoor(c *cli.Context, opts *terminal.CommandOpts) error {
	log := opts.Logger()
	configPath := c.GlobalString("config-path")
	path, err := homedir.Expand(configPath)
	if err != nil {
		return fmt.Errorf("error expanding config path [%s]: %w", configPath, err)
	}

	// Check directory exists
	dirInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("%sYour configured contributoor directory [%s] does not exist. Please run 'contributoor install' first%s", terminal.ColorRed, path, terminal.ColorReset)
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("%s[%s] is not a directory%s", terminal.ColorRed, path, terminal.ColorReset)
	}

	// Check config file exists
	configFile := filepath.Join(path, "config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("%sConfig file not found at [%s]. Please run 'contributoor install' first%s", terminal.ColorRed, configFile, terminal.ColorReset)
	}

	configService, err := service.NewConfigService(log, configFile)
	if err != nil {
		return err
	}

	switch configService.Get().RunMethod {
	case service.RunMethodDocker:
		log.WithField("version", configService.Get().Version).Info("Stopping Contributoor")

		dockerService, err := service.NewDockerService(log, configService)
		if err != nil {
			log.Errorf("could not create docker service: %v", err)
			return err
		}

		// Check if running before attempting to stop
		running, err := dockerService.IsRunning()
		if err != nil {
			log.Errorf("could not check service status: %v", err)
			return err
		}
		if !running {
			return fmt.Errorf("%sContributoor is not running. Use 'contributoor start' to start it%s", terminal.ColorRed, terminal.ColorReset)
		}

		if err := dockerService.Stop(); err != nil {
			log.Errorf("could not stop service: %v", err)
			return err
		}
	case service.RunMethodBinary:
		binaryService := service.NewBinaryService(log, configService)
		if err := binaryService.Stop(); err != nil {
			log.Errorf("could not stop service: %v", err)
			return err
		}
	}

	return nil
}
