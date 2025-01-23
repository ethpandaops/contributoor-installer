package sidecar

// RunMethods defines the possible ways to run the contributoor service.
const (
	RunMethodDocker  = "docker"
	RunMethodSystemd = "systemd"
	RunMethodBinary  = "binary"
)

const (
	ArchDarwin = "darwin"
	ArchLinux  = "linux"
)

// SidecarRunner handles operations for the various run methods.
type SidecarRunner interface {
	// Start starts the service.
	Start() error

	// Stop stops the service.
	Stop() error

	// Update updates the service.
	Update() error

	// Status returns the status of the service.
	Status() (string, error)

	// IsRunning checks if the service is running.
	IsRunning() (bool, error)

	// Logs returns the logs from the service.
	Logs(tailLines int, follow bool) error

	// Version returns the current version the underlying sidecar is running.
	Version() (string, error)
}
