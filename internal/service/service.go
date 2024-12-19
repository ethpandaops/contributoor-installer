package service

// RunMethods defines the possible ways to run the contributoor service.
const (
	RunMethodDocker = "docker"
	RunMethodBinary = "binary"
)

// ServiceRunner handles operations for the various run methods.
type ServiceRunner interface {
	// Start starts the service.
	Start() error

	// Stop stops the service.
	Stop() error

	// Update updates the service.
	Update() error

	// IsRunning checks if the service is running.
	IsRunning() (bool, error)
}

// ConfigManager defines the interface for configuration management.
type ConfigManager interface {
	// Update modifies the configuration using the provided update function.
	Update(updates func(*ContributoorConfig)) error

	// Get returns the current configuration.
	Get() *ContributoorConfig

	// GetConfigDir returns the directory containing the config file.
	GetConfigDir() string

	// GetConfigPath returns the path of the file config.
	GetConfigPath() string

	// Save persists the current configuration to disk.
	Save() error
}
