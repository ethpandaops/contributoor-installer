/*
	Package service provides the configuration management system for Contributoor.

	Config Service Overview:
	----------------------
	The ConfigService handles loading, saving, and migrating user configurations
	while preserving user settings during updates. This is necessary to handle
	scenarios like:

	- Adding new config fields in newer versions
	- Changing the format/structure of existing fields
	- Preserving user customizations during updates
	- Ensuring safe atomic writes of config files

How It Works:
-------------
1. Loading Config:

  - Service loads user's existing config file
  - Creates new config with latest defaults
  - Merges user's settings into new config
  - Detects version differences

2. Version Migration (e.g., 0.0.1 -> 0.0.2):
  - Runs version-specific migrations
  - Transforms old values if needed
  - Adds new fields with defaults
  - Preserves user's custom settings

3. Saving Config:
  - Writes to temporary file first
  - Uses atomic rename for safety
  - Prevents corruption during writes

Example Migration:
----------------
Starting with user's v0.0.1 config:

	version: "0.0.1"
	network:
	  name: "mainnet"

When v0.0.2 adds a new field:

	version: "0.0.2"
	network:
	  name: "mainnet"
	logLevel: "info"  # New field

The service will:
1. Load user's 0.0.1 config
2. Create new 0.0.2 config with defaults
3. Merge: Keep user's network name, add logLevel default
4. Run migrations if needed
5. Save updated config atomically
*/
package service

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/sirupsen/logrus"
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

type ConfigService struct {
	logger     *logrus.Logger
	configPath string
	configDir  string
	config     *ContributoorConfig
}

func NewConfigService(logger *logrus.Logger, configPath string) (*ConfigService, error) {
	// If configPath is a directory, append config.yaml
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat config path: %w", err)
	}

	fullConfigPath := configPath
	if fileInfo.IsDir() {
		fullConfigPath = filepath.Join(configPath, "config.yaml")
	}

	// Load existing config
	data, err := os.ReadFile(fullConfigPath)
	if err != nil {
		return nil, err
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
		if err := migrateConfig(logger, newConfig, oldConfig); err != nil {
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

func (s *ConfigService) GetConfigDir() string {
	return s.configDir
}

func (s *ConfigService) Get() *ContributoorConfig {
	return s.config
}

func WriteConfig(path string, cfg *ContributoorConfig) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

func newDefaultConfig() *ContributoorConfig {
	return &ContributoorConfig{
		Version:   "latest",
		RunMethod: RunMethodDocker,
		Network: &NetworkConfig{
			Name:              "mainnet",
			BeaconNodeAddress: "http://localhost:5052",
		},
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
	return nil
}

// mergeConfig merges old config values into new config
func mergeConfig(new, old *ContributoorConfig) error {
	// Use reflection to copy non-zero values from old to new
	newVal := reflect.ValueOf(new).Elem()
	oldVal := reflect.ValueOf(old).Elem()

	for i := 0; i < newVal.NumField(); i++ {
		newField := newVal.Field(i)
		oldField := oldVal.FieldByName(newVal.Type().Field(i).Name)

		if oldField.IsValid() && !oldField.IsZero() {
			newField.Set(oldField)
		}
	}

	return nil
}

// migrateConfig handles version-specific migrations
func migrateConfig(logger *logrus.Logger, new, old *ContributoorConfig) error {
	switch old.Version {
	case "0.0.1":
		// For example, 0.0.1 -> 0.0.2, we might want to migrate the network name.
		// oldName := old.Network.Name
		// new.Network.Name = old.Network.Name + "_migrated"
		//
		// Or perhaps we want to populate a new default value for a new field.
		// new.Foo = "bar"
	}
	return nil
}
