package sidecar

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/installer"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/sirupsen/logrus"
)

//go:generate mockgen -package mock -destination mock/systemd.mock.go github.com/ethpandaops/contributoor-installer/internal/sidecar SystemdSidecar

type SystemdSidecar interface {
	SidecarRunner
}

// systemdSidecar is a service for managing the contributoor service (systemd on Linux, launchd on macOS).
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

// Start starts the service.
func (s *systemdSidecar) Start() error {
	if err := s.checkBinaryExists(); err != nil {
		return wrapNotInstalledError(err, "systemd")
	}

	if runtime.GOOS == ArchDarwin {
		return s.startLaunchd()
	}

	return s.startSystemd()
}

// Stop stops the service.
func (s *systemdSidecar) Stop() error {
	if err := s.checkDaemonExists(); err != nil {
		return err
	}

	if runtime.GOOS == ArchDarwin {
		return s.stopLaunchd()
	}

	return s.stopSystemd()
}

// IsRunning checks if the service is running.
func (s *systemdSidecar) IsRunning() (bool, error) {
	if runtime.GOOS == ArchDarwin {
		return s.isRunningLaunchd()
	}

	return s.isRunningSystemd()
}

// Update updates the service.
func (s *systemdSidecar) Update() error {
	// Stop service if running
	if running, _ := s.IsRunning(); running {
		if err := s.Stop(); err != nil {
			return fmt.Errorf("failed to stop service for update: %w", err)
		}
	}

	// systemd + launchd are underpinned by the binary sidecar.
	binarySidecar, err := NewBinarySidecar(s.logger, s.sidecarCfg, s.installerCfg)
	if err != nil {
		return fmt.Errorf("failed to create binary sidecar: %w", err)
	}

	if err := binarySidecar.Update(); err != nil {
		return fmt.Errorf("failed to update binary: %w", err)
	}

	// Reload service manager
	if runtime.GOOS == ArchDarwin {
		return s.reloadLaunchd()
	}

	return s.reloadSystemd()
}

// Logs shows the logs from the service.
func (s *systemdSidecar) Logs(tailLines int, follow bool) error {
	// For macOS, use binary logs.
	if runtime.GOOS == ArchDarwin {
		binarySidecar, err := NewBinarySidecar(s.logger, s.sidecarCfg, s.installerCfg)
		if err != nil {
			return fmt.Errorf("failed to create binary sidecar for logs: %w", err)
		}

		return binarySidecar.Logs(tailLines, follow)
	}

	// For Linux/systemd, use journalctl.
	args := []string{"-u", "contributoor.service"}

	if follow {
		args = append(args, "-f")
	}

	if tailLines > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", tailLines))
	}

	//nolint:gosec // controlled input.
	cmd := exec.Command("sudo", append([]string{"journalctl"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (s *systemdSidecar) startSystemd() error {
	if err := s.checkDaemonExists(); err != nil {
		return wrapNotInstalledError(err, "systemd")
	}

	cmd := exec.Command("sudo", "systemctl", "start", "contributoor.service")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start service: %s: %w", string(output), err)
	}

	fmt.Printf("%sContributoor started successfully%s\n", tui.TerminalColorGreen, tui.TerminalColorReset)

	return nil
}

func (s *systemdSidecar) stopSystemd() error {
	if err := s.checkDaemonExists(); err != nil {
		return wrapNotInstalledError(err, "systemd")
	}

	cmd := exec.Command("sudo", "systemctl", "stop", "contributoor.service")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop service: %s: %w", string(output), err)
	}

	fmt.Printf("%sContributoor stopped successfully%s\n", tui.TerminalColorGreen, tui.TerminalColorReset)

	return nil
}

func (s *systemdSidecar) isRunningSystemd() (bool, error) {
	if err := s.checkDaemonExists(); err != nil {
		//nolint:nilerr // We want to return false if the service doesn't exist.
		return false, nil
	}

	cmd := exec.Command("sudo", "systemctl", "is-active", "contributoor.service")

	output, err := cmd.CombinedOutput()
	if err != nil {
		//nolint:nilerr // We want to return false if the service doesn't exist.
		return false, nil
	}

	return strings.TrimSpace(string(output)) == "active", nil
}

func (s *systemdSidecar) reloadSystemd() error {
	cmd := exec.Command("sudo", "systemctl", "daemon-reload")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reload systemd: %s: %w", string(output), err)
	}

	return nil
}

func (s *systemdSidecar) startLaunchd() error {
	if err := s.checkDaemonExists(); err != nil {
		return wrapNotInstalledError(err, "launchd")
	}

	// Mac's launchd is a bit different from systemd. We need to load the service first.
	cmd := exec.Command("sudo", "launchctl", "load", "-w", "/Library/LaunchDaemons/io.ethpandaops.contributoor.plist")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to load service: %s: %w", string(output), err)
	}

	// Then we can start it.
	cmd = exec.Command("sudo", "launchctl", "start", "io.ethpandaops.contributoor")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start service: %s: %w", string(output), err)
	}

	fmt.Printf("%sContributoor started successfully%s\n", tui.TerminalColorGreen, tui.TerminalColorReset)

	return nil
}

func (s *systemdSidecar) stopLaunchd() error {
	if err := s.checkDaemonExists(); err != nil {
		return wrapNotInstalledError(err, "launchd")
	}

	// First stop the service.
	cmd := exec.Command("sudo", "launchctl", "stop", "io.ethpandaops.contributoor")
	_ = cmd.Run()

	// Then (similar to what we do with Start()), mac requires us to unload the service, otherwise it never stops.
	cmd = exec.Command("sudo", "launchctl", "unload", "/Library/LaunchDaemons/io.ethpandaops.contributoor.plist")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to unload service: %s: %w", string(output), err)
	}

	fmt.Printf("%sContributoor stopped successfully%s\n", tui.TerminalColorGreen, tui.TerminalColorReset)

	return nil
}

func (s *systemdSidecar) isRunningLaunchd() (bool, error) {
	if err := s.checkDaemonExists(); err != nil {
		//nolint:nilerr // We want to return false if the service doesn't exist.
		return false, nil
	}

	cmd := exec.Command("sudo", "launchctl", "list", "io.ethpandaops.contributoor")

	output, err := cmd.CombinedOutput()
	if err != nil {
		//nolint:nilerr // We want to return false if the service doesn't exist.
		return false, nil
	}

	// If service is running, output will contain a PID
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		fields := strings.Fields(lines[0])

		return len(fields) > 0 && fields[0] != "-", nil
	}

	return false, nil
}

func (s *systemdSidecar) reloadLaunchd() error {
	cmd := exec.Command("sudo", "launchctl", "unload", "/Library/LaunchDaemons/io.ethpandaops.contributoor.plist")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to unload service: %s: %w", string(output), err)
	}

	cmd = exec.Command("sudo", "launchctl", "load", "-w", "/Library/LaunchDaemons/io.ethpandaops.contributoor.plist")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reload service: %s: %w", string(output), err)
	}

	return nil
}

// checkDaemonExists checks if the daemon exists.
func (s *systemdSidecar) checkDaemonExists() error {
	if runtime.GOOS == ArchDarwin {
		// Check if plist file exists
		cmd := exec.Command("sudo", "test", "-f", "/Library/LaunchDaemons/io.ethpandaops.contributoor.plist")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("service not installed")
		}

		return nil
	}

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

// checkBinaryExists checks if the binary exists and has the correct version.
func (s *systemdSidecar) checkBinaryExists() error {
	// Create a binary sidecar to check version
	bs, err := NewBinarySidecar(s.logger, s.sidecarCfg, s.installerCfg)
	if err != nil {
		return fmt.Errorf("failed to create binary sidecar: %w", err)
	}

	// Check binary version
	if impl, ok := bs.(*binarySidecar); ok {
		if err := impl.checkBinaryVersion(); err != nil {
			return fmt.Errorf("version check failed: %w", err)
		}
	}

	return nil
}
