package sidecar

// RunMethods defines the possible ways to run the contributoor service.
const (
	RunMethodDocker = "docker"
	RunMethodBinary = "binary"
)

// SidecarRunner handles operations for the various run methods.
type SidecarRunner interface {
	// Start starts the service.
	Start() error

	// Stop stops the service.
	Stop() error

	// Update updates the service.
	Update() error

	// IsRunning checks if the service is running.
	IsRunning() (bool, error)
}
