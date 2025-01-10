package sidecar

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

//go:generate mockgen -package mock -destination mock/config.mock.go github.com/ethpandaops/contributoor-installer/internal/sidecar ConfigManager

// ConfigManager provides the configuration management system for the Contributoor sidecar.
// It handles loading, saving, and migrating user configurations while preserving
// user settings during updates. This is necessary to handle scenarios like:
//
// - Adding new config fields in newer versions.
// - Changing the format/structure of existing fields.
// - Preserving user customizations during updates.
// - Ensuring safe atomic writes of config files.
type ConfigManager interface {
	// Save persists the current configuration to disk.
	Save() error

	// Update modifies the configuration using the provided update function.
	Update(updates func(*config.Config)) error

	// Get returns the current configuration.
	Get() *config.Config

	// GetConfigPath returns the path of the file config.
	GetConfigPath() string
}

// configService is a basic service for interacting with file configuration.
type configService struct {
	logger     *logrus.Logger
	configPath string
	config     *config.Config
}

// NewConfigService creates a new ConfigManager.
func NewConfigService(logger *logrus.Logger, configPath string) (ConfigManager, error) {
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
		return nil, wrapMissingConfigError(fmt.Errorf("config file not found at [%s]", fullConfigPath))
	}

	// Load existing config
	data, err := os.ReadFile(fullConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// First unmarshal YAML into a map
	var yamlMap map[string]interface{}
	if yerr := yaml.Unmarshal(data, &yamlMap); yerr != nil {
		return nil, wrapInvalidConfigError(yerr)
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(yamlMap)
	if err != nil {
		return nil, wrapInvalidConfigError(err)
	}

	oldConfig := &config.Config{}
	if err := protojson.Unmarshal(jsonBytes, oldConfig); err != nil {
		return nil, wrapInvalidConfigError(err)
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
		if err := writeConfig(fullConfigPath, newConfig); err != nil {
			return nil, fmt.Errorf("failed to save migrated config: %w", err)
		}
	}

	return &configService{
		logger:     logger,
		configPath: fullConfigPath,
		config:     newConfig,
	}, nil
}

func newDefaultConfig() *config.Config {
	return &config.Config{
		LogLevel:          logrus.InfoLevel.String(),
		Version:           "latest",
		RunMethod:         config.RunMethod_RUN_METHOD_DOCKER,
		NetworkName:       config.NetworkName_NETWORK_NAME_MAINNET,
		BeaconNodeAddress: "",
		OutputServer: &config.OutputServer{
			Address: tui.OutputServerProduction,
			Tls:     true,
		},
	}
}

// Update updates the file config with the given updates.
func (s *configService) Update(updates func(*config.Config)) error {
	// Clone the config.
	updatedConfig, ok := proto.Clone(s.config).(*config.Config)
	if !ok {
		return fmt.Errorf("failed to clone config")
	}

	updates(updatedConfig)

	// Validate the updated config
	if err := s.validate(updatedConfig); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Write to temporary file first
	tmpPath := fmt.Sprintf("%s.tmp", s.configPath)
	if err := writeConfig(tmpPath, updatedConfig); err != nil {
		os.Remove(tmpPath)

		return err
	}

	// Atomic rename
	if err := os.Rename(tmpPath, s.configPath); err != nil {
		os.Remove(tmpPath)

		return fmt.Errorf("failed to save config: %w", err)
	}

	// Update internal state
	s.config = updatedConfig

	return nil
}

// Get returns the current file config.
func (s *configService) Get() *config.Config {
	return s.config
}

// GetConfigPath returns the path of the file config.
func (s *configService) GetConfigPath() string {
	return s.configPath
}

// Save persists the current configuration to disk.
func (s *configService) Save() error {
	return writeConfig(s.configPath, s.config)
}

// writeConfig writes the file config to the given path.
func writeConfig(path string, cfg *config.Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// We wanna keep hold of the camelCase output in yaml.
	jsonData, err := protojson.MarshalOptions{
		UseProtoNames:   false, // This ensures we use camelCase.
		EmitUnpopulated: false,
	}.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error marshaling config to json: %w", err)
	}

	// Now marshal to map for YAML.
	var jsonMap map[string]interface{}
	if jerr := json.Unmarshal(jsonData, &jsonMap); jerr != nil {
		return fmt.Errorf("error unmarshaling json: %w", jerr)
	}

	data, err := yaml.Marshal(jsonMap)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// validate validates the config.
func (s *configService) validate(cfg *config.Config) error {
	if cfg.Version == "" {
		return fmt.Errorf("version is required")
	}

	if cfg.ContributoorDirectory == "" {
		return fmt.Errorf("contributoorDirectory is required")
	}

	if cfg.RunMethod == config.RunMethod_RUN_METHOD_UNSPECIFIED {
		return fmt.Errorf("invalid runMethod: %s", cfg.RunMethod)
	}

	if cfg.NetworkName == config.NetworkName_NETWORK_NAME_UNSPECIFIED {
		return fmt.Errorf("networkName is required")
	}

	return nil
}

// mergeConfig merges old config values into new config.
func mergeConfig(target, source *config.Config) error {
	if source != nil {
		proto.Merge(target, source)
	}

	return nil
}

// migrateConfig handles version-specific migrations.
func migrateConfig(target, source *config.Config) error {
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

// wrapInvalidConfigError wraps an error with a user-friendly message.
func wrapInvalidConfigError(err error) error {
	return fmt.Errorf("configuration error:\n\n"+
		"Your config.yaml file appears to be invalid. Please check:\n"+
		"1. All fields are correctly spelled\n"+
		"2. No unknown fields are present\n"+
		"3. All required fields are set\n\n"+
		"If the problem persists, try removing your config.yaml and re-running install.sh\n\n"+
		"For detailed configuration help, visit: https://github.com/ethpandaops/contributoor#configuration\n\n"+
		"Debug details: %w",
		err)
}

// wrapMissingConfigError wraps an error with a user-friendly message.
func wrapMissingConfigError(err error) error {
	return fmt.Errorf("configuration error:\n\n"+
		"Your config.yaml file does not exist. Please run 'contributoor install' first.\n\n"+
		"Debug details: %w",
		err)
}
