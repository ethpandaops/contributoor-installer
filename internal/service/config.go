/*
Package service provides the configuration management system for Contributoor.
It handles loading, saving, and migrating user configurations while preserving
user settings during updates. This is necessary to handle scenarios like:

- Adding new config fields in newer versions
- Changing the format/structure of existing fields
- Preserving user customizations during updates
- Ensuring safe atomic writes of config files
*/
package service

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// RunMethods defines the possible ways to run the contributoor service.
const (
	RunMethodDocker = "docker"
	RunMethodBinary = "binary"
)

// ContributoorConfig is the configuration for the contributoor service.
type ContributoorConfig struct {
	Version               string              `yaml:"version"`
	ContributoorDirectory string              `yaml:"contributoorDirectory"`
	RunMethod             string              `yaml:"runMethod"`
	NetworkName           string              `yaml:"networkName"`
	BeaconNodeAddress     string              `yaml:"beaconNodeAddress"`
	OutputServer          *OutputServerConfig `yaml:"outputServer"`
}

// OutputServerConfig is the configuration for the output server.
type OutputServerConfig struct {
	Address     string `yaml:"address"`
	Credentials string `yaml:"credentials,omitempty"`
}

// ConfigService is a basic service for interacting with file configuration.
type ConfigService struct {
	logger     *logrus.Logger
	configPath string
	configDir  string
	config     *ContributoorConfig
}

// ConfigNotFoundError is an error that occurs when the file config is not found.
type ConfigNotFoundError struct {
	Path string
}

// Error returns the error message.
func (e *ConfigNotFoundError) Error() string {
	return fmt.Sprintf("Config file not found at [%s]. Please run 'contributoor install' first", e.Path)
}

// NewConfigService creates a new ConfigService.
func NewConfigService(logger *logrus.Logger, configPath string) (*ConfigService, error) {
	// Expand home directory
	path, err := homedir.Expand(configPath)
	if err != nil {
		return nil, fmt.Errorf("error expanding config path [%s]: %w", configPath, err)
	}

	// Check directory exists
	dirInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("directory [%s] does not exist", path)
	}

	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("[%s] is not a directory", path)
	}

	// Determine full config path
	fullConfigPath := filepath.Join(path, "config.yaml")

	// Check if config exists
	if _, serr := os.Stat(fullConfigPath); os.IsNotExist(serr) {
		return nil, &ConfigNotFoundError{Path: fullConfigPath}
	}

	// Load existing config
	data, err := os.ReadFile(fullConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	oldConfig := &ContributoorConfig{}
	if err := yaml.Unmarshal(data, oldConfig); err != nil {
		return nil, err
	}

	// Get default config with latest schema
	newConfig := newDefaultConfig()

	// Merge old config into new config
	if err := mergeConfig(newConfig, oldConfig); err != nil {
		return nil, fmt.Errorf("failed to migrate config: %w", err)
	}

	// Check if config needs migration by comparing versions
	if oldConfig.Version != newConfig.Version {
		// Perform version-specific migrations
		if err := migrateConfig(newConfig, oldConfig); err != nil {
			return nil, fmt.Errorf("failed to migrate config: %w", err)
		}

		// Save migrated config
		if err := WriteConfig(fullConfigPath, newConfig); err != nil {
			return nil, fmt.Errorf("failed to save migrated config: %w", err)
		}
	}

	return &ConfigService{
		logger:     logger,
		configPath: fullConfigPath,
		configDir:  filepath.Dir(fullConfigPath),
		config:     newConfig,
	}, nil
}

// Update updates the file config with the given updates.
func (s *ConfigService) Update(updates func(*ContributoorConfig)) error {
	// Apply updates to a copy
	updatedConfig := *s.config
	updates(&updatedConfig)

	// Validate the updated config
	if err := s.validate(&updatedConfig); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Write to temporary file first
	tmpPath := s.configPath + ".tmp"
	if err := WriteConfig(tmpPath, &updatedConfig); err != nil {
		os.Remove(tmpPath)

		return err
	}

	// Atomic rename
	if err := os.Rename(tmpPath, s.configPath); err != nil {
		os.Remove(tmpPath)

		return fmt.Errorf("failed to save config: %w", err)
	}

	// Update internal state
	s.config = &updatedConfig

	return nil
}

// Get returns the current file config.
func (s *ConfigService) Get() *ContributoorConfig {
	return s.config
}

// GetConfigDir returns the directory of the file config.
func (s *ConfigService) GetConfigDir() string {
	return s.configDir
}

// WriteConfig writes the file config to the given path.
func WriteConfig(path string, cfg *ContributoorConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

func newDefaultConfig() *ContributoorConfig {
	return &ContributoorConfig{
		Version:           "latest",
		RunMethod:         RunMethodDocker,
		NetworkName:       "mainnet",
		BeaconNodeAddress: "http://localhost:5052",
	}
}

func (s *ConfigService) validate(cfg *ContributoorConfig) error {
	if cfg.Version == "" {
		return fmt.Errorf("version is required")
	}

	if cfg.ContributoorDirectory == "" {
		return fmt.Errorf("contributoorDirectory is required")
	}

	if cfg.RunMethod != RunMethodDocker && cfg.RunMethod != RunMethodBinary {
		return fmt.Errorf("invalid runMethod: %s", cfg.RunMethod)
	}

	if cfg.NetworkName == "" {
		return fmt.Errorf("networkName is required")
	}

	if cfg.BeaconNodeAddress == "" {
		return fmt.Errorf("beaconNodeAddress is required")
	}

	return nil
}

// mergeConfig merges old config values into new config.
func mergeConfig(target, source *ContributoorConfig) error {
	// Use reflection to copy non-zero values from old to new
	newVal := reflect.ValueOf(target).Elem()
	oldVal := reflect.ValueOf(source).Elem()

	for i := 0; i < newVal.NumField(); i++ {
		newField := newVal.Field(i)
		oldField := oldVal.FieldByName(newVal.Type().Field(i).Name)

		if oldField.IsValid() && !oldField.IsZero() {
			newField.Set(oldField)
		}
	}

	return nil
}

// migrateConfig handles version-specific migrations.

func migrateConfig(target, source *ContributoorConfig) error {
	/*
		switch source.Version {
			case "0.0.1":
				For example, 0.0.1 -> 0.0.2, we might want to migrate the network name.
				oldName := old.Network.Name
				new.Network.Name = old.Network.Name + "_migrated"

			Or perhaps we want to populate a new default value for a new field.
			new.Foo = "bar"
		}
	*/
	return nil
}

func (s *ConfigService) Save() error {
	return WriteConfig(s.configPath, s.config)
}
