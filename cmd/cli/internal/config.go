package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	ContributorVersion = "e29c4ae125d4787e4b48c5787f0f4998c8db15c5"
	RunMethodDocker    = "docker"
	RunMethodBinary    = "binary"
)

type ContributoorConfig struct {
	Title                 string         `yaml:"title"`
	Version               string         `yaml:"version"`
	ContributoorDirectory string         `yaml:"contributoor_directory"`
	Network               *NetworkConfig `yaml:"network"`
	RunMethod             string         `yaml:"run_method"`
}

type NetworkConfig struct {
	Name              Parameter `yaml:"name"`
	BeaconNodeAddress Parameter `yaml:"beaconNodeAddress"`
}

// A parameter that can be configured by the user
type Parameter struct {
	ID                 string      `yaml:"id,omitempty"`
	Name               string      `yaml:"name,omitempty"`
	Description        string      `yaml:"description,omitempty"`
	MaxLength          int         `yaml:"maxLength,omitempty"`
	CanBeBlank         bool        `yaml:"canBeBlank,omitempty"`
	OverwriteOnUpgrade bool        `yaml:"overwriteOnUpgrade,omitempty"`
	Value              interface{} `yaml:"value,omitempty"`
}

func NewContributoorConfig(dir string) *ContributoorConfig {
	return &ContributoorConfig{
		Title:                 "Contributoor",
		Version:               "latest",
		ContributoorDirectory: dir,
		Network: &NetworkConfig{
			Name: Parameter{
				ID:          "networkName",
				Name:        "Network name",
				Value:       "",
				Description: "The name of the network",
			},
			BeaconNodeAddress: Parameter{
				ID:          "beaconNodeAddress",
				Name:        "Beacon node address",
				Value:       "",
				Description: "The address of the beacon node to attach to",
			},
		},
		RunMethod: RunMethodDocker,
	}
}

func (cfg *NetworkConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&cfg.Name,
		&cfg.BeaconNodeAddress,
	}
}

func LoadConfig(path string) (*ContributoorConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := NewContributoorConfig(path)
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
