package sidecar

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/installer"
	"github.com/sirupsen/logrus"
)

//go:generate mockgen -package mock -destination mock/systemd.mock.go github.com/ethpandaops/contributoor-installer/internal/sidecar SystemdSidecar

type SystemdSidecar interface {
	SidecarRunner
}

// systemdSidecar is a service for managing the contributoor systemd service.
type systemdSidecar struct {
	logger       *logrus.Logger
	sidecarCfg   ConfigManager
	installerCfg *installer.Config
}

// NewSystemdSidecar creates a new SystemdSidecar.
func NewSystemdSidecar(logger *logrus.Logger, sidecarCfg ConfigManager, installerCfg *installer.Config) (SystemdSidecar, error) {
	return &systemdSidecar{
		logger:       logger,
		sidecarCfg:   sidecarCfg,
		installerCfg: installerCfg,
	}, nil
}

// Start starts the systemd service.
func (s *systemdSidecar) Start() error {
	// Check if service exists
	if err := s.checkServiceExists(); err != nil {
		return fmt.Errorf("service not found: %w", err)
	}

	// Start the service
	cmd := exec.Command("sudo", "systemctl", "start", "contributoor.service")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start service: %s: %w", string(output), err)
	}

	// Verify service is running
	if running, err := s.IsRunning(); err != nil {
		return fmt.Errorf("failed to verify service status: %w", err)
	} else if !running {
		return fmt.Errorf("service failed to start")
	}

	return nil
}

// Stop stops the systemd service.
func (s *systemdSidecar) Stop() error {
	// Check if service exists
	if err := s.checkServiceExists(); err != nil {
		return fmt.Errorf("service not found: %w", err)
	}

	// Stop the service
	cmd := exec.Command("sudo", "systemctl", "stop", "contributoor.service")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop service: %s: %w", string(output), err)
	}

	// Verify service is stopped
	if running, err := s.IsRunning(); err != nil {
		return fmt.Errorf("failed to verify service status: %w", err)
	} else if running {
		return fmt.Errorf("service failed to stop")
	}

	return nil
}

// IsRunning checks if the systemd service is running.
func (s *systemdSidecar) IsRunning() (bool, error) {
	// Check if service exists first
	if err := s.checkServiceExists(); err != nil {
		return false, nil
	}

	// Get service status
	cmd := exec.Command("sudo", "systemctl", "is-active", "contributoor.service")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// systemctl is-active returns non-zero if service is not active
		return false, nil
	}

	return strings.TrimSpace(string(output)) == "active", nil
}

// Update updates the systemd service.
func (s *systemdSidecar) Update() error {
	// Check if service exists
	if err := s.checkServiceExists(); err != nil {
		return fmt.Errorf("service not found: %w", err)
	}

	// Stop service if running
	if running, _ := s.IsRunning(); running {
		if err := s.Stop(); err != nil {
			return fmt.Errorf("failed to stop service for update: %w", err)
		}
	}

	// Update binary
	binarySidecar, err := NewBinarySidecar(s.logger, s.sidecarCfg, s.installerCfg)
	if err != nil {
		return fmt.Errorf("failed to create binary sidecar: %w", err)
	}

	if err := binarySidecar.Update(); err != nil {
		return fmt.Errorf("failed to update binary: %w", err)
	}

	// Reload systemd daemon
	cmd := exec.Command("sudo", "systemctl", "daemon-reload")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reload systemd: %s: %w", string(output), err)
	}

	return nil
}

// checkServiceExists verifies the systemd service exists.
func (s *systemdSidecar) checkServiceExists() error {
	cmd := exec.Command("sudo", "systemctl", "list-unit-files", "contributoor.service")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list service: %s: %w", string(output), err)
	}

	if !strings.Contains(string(output), "contributoor.service") {
		return fmt.Errorf("service not installed")
	}

	return nil
}
