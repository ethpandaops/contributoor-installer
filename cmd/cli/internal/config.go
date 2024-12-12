package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	RunMethodDocker = "docker"
	RunMethodBinary = "binary"
)

type ContributoorConfig struct {
	Version               string         `yaml:"version"`
	ContributoorDirectory string         `yaml:"contributoorDirectory"`
	RunMethod             string         `yaml:"runMethod"`
	Network               *NetworkConfig `yaml:"network"`
}

type NetworkConfig struct {
	Name              string `yaml:"name"`
	BeaconNodeAddress string `yaml:"beaconNodeAddress"`
}

func LoadConfig(path string) (*ContributoorConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &ContributoorConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *ContributoorConfig) WriteToFile(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}
