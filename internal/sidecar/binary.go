package sidecar

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/ethpandaops/contributoor-installer/internal/installer"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

//go:generate mockgen -package mock -destination mock/binary.mock.go github.com/ethpandaops/contributoor-installer/internal/sidecar BinarySidecar

type BinarySidecar interface {
	SidecarRunner
}

// binarySidecar is a basic service for interacting with the contributoor binary.
type binarySidecar struct {
	logger       *logrus.Logger
	sidecarCfg   ConfigManager
	installerCfg *installer.Config
	stdout       *os.File
	stderr       *os.File
}

// NewBinarySidecar creates a new BinarySidecar.
func NewBinarySidecar(logger *logrus.Logger, sidecarCfg ConfigManager, installerCfg *installer.Config) (BinarySidecar, error) {
	expandedDir, err := homedir.Expand(sidecarCfg.Get().ContributoorDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to expand config path: %w", err)
	}

	logsDir := filepath.Join(expandedDir, "logs")

	// Open log files
	stdout, err := os.OpenFile(filepath.Join(logsDir, "debug.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open stdout log file: %w", err)
	}

	stderr, err := os.OpenFile(filepath.Join(logsDir, "service.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		stdout.Close()

		return nil, fmt.Errorf("failed to open stderr log file: %w", err)
	}

	return &binarySidecar{
		logger:       logger,
		stdout:       stdout,
		stderr:       stderr,
		sidecarCfg:   sidecarCfg,
		installerCfg: installerCfg,
	}, nil
}

// Start starts the binary service.
func (s *binarySidecar) Start() error {
	cfg := s.sidecarCfg.Get()

	binaryPath := filepath.Join(cfg.ContributoorDirectory, "bin", "sentry")
	if _, err := os.Stat(binaryPath); err != nil {
		return fmt.Errorf("binary not found at %s - please reinstall", binaryPath)
	}

	expandedDir, err := homedir.Expand(cfg.ContributoorDirectory)
	if err != nil {
		return fmt.Errorf("failed to expand config path: %w", err)
	}

	configPath := filepath.Join(expandedDir, "config.yaml")
	cmd := exec.Command(binaryPath, "--config", configPath)
	cmd.Stdout = s.stderr
	cmd.Stderr = s.stdout

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start binary: %w", err)
	}

	pidFile := filepath.Join(cfg.ContributoorDirectory, "contributoor.pid")
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0600); err != nil {
		return fmt.Errorf("failed to write pid file: %w", err)
	}

	// Start a goroutine to wait for the process and clean up
	go func() {
		defer s.stdout.Close()
		defer s.stderr.Close()

		if err := cmd.Wait(); err != nil {
			s.logger.Errorf("Process exited with error: %v", err)
		}

		// Clean up pid file
		if err := os.Remove(pidFile); err != nil {
			s.logger.Errorf("Failed to remove pid file: %v", err)
		}
	}()

	fmt.Printf("%sContributoor started successfully%s\n", tui.TerminalColorGreen, tui.TerminalColorReset)

	return nil
}

// Stop stops the binary service.
func (s *binarySidecar) Stop() error {
	cfg := s.sidecarCfg.Get()

	pidFile := filepath.Join(cfg.ContributoorDirectory, "contributoor.pid")

	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("failed to read pid file: %w", err)
	}

	pidStr := string(pidBytes)
	if !regexp.MustCompile(`^\d+$`).MatchString(pidStr) {
		return fmt.Errorf("invalid PID format")
	}

	//nolint:gosec // sanitized.
	cmd := exec.Command("kill", string(pidBytes))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop process: %w", err)
	}

	os.Remove(pidFile)

	// Close log files if they exist
	if s.stdout != nil {
		s.stdout.Close()
		s.stdout = nil
	}

	if s.stderr != nil {
		s.stderr.Close()
		s.stderr = nil
	}

	fmt.Printf("%sContributoor stopped successfully%s\n", tui.TerminalColorGreen, tui.TerminalColorReset)

	return nil
}

// IsRunning checks if the binary service is running.
func (s *binarySidecar) IsRunning() (bool, error) {
	cfg := s.sidecarCfg.Get()

	pidFile := filepath.Join(cfg.ContributoorDirectory, "contributoor.pid")
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		return false, nil
	}

	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		return false, err
	}

	pidStr := string(pidBytes)
	if !regexp.MustCompile(`^\d+$`).MatchString(pidStr) {
		return false, fmt.Errorf("invalid PID format")
	}

	// kill -0 just checks if process exists. It doesn't actually send a
	// signal that affects the process.
	cmd := exec.Command("kill", "-0", pidStr)
	if err := cmd.Run(); err != nil {
		os.Remove(pidFile)

		//nolint:nilerr // We don't care about the error here.
		return false, nil
	}

	return true, nil
}

// Update updates the binary service.
func (s *binarySidecar) Update() error {
	cfg := s.sidecarCfg.Get()

	expandedDir, err := homedir.Expand(cfg.ContributoorDirectory)
	if err != nil {
		return fmt.Errorf("failed to expand config path: %w", err)
	}

	binaryPath := filepath.Join(expandedDir, "bin", "sentry")
	binaryDir := filepath.Dir(binaryPath)

	// Download and verify checksums.
	checksumURL := fmt.Sprintf(
		"https://github.com/%s/%s/releases/download/v%s/contributoor_%s_checksums.txt",
		s.installerCfg.GithubOrg,
		s.installerCfg.GithubRepo,
		cfg.Version,
		cfg.Version,
	)

	//nolint:gosec // controlled url.
	resp, err := http.Get(checksumURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download checksums: %w", err)
	}

	defer resp.Body.Close()

	// Determine platform and arch
	var (
		platform = runtime.GOOS
		arch     = runtime.GOARCH
	)

	binaryURL := fmt.Sprintf(
		"https://github.com/%s/%s/releases/download/v%s/contributoor_%s_%s_%s.tar.gz",
		s.installerCfg.GithubOrg,
		s.installerCfg.GithubRepo,
		cfg.Version,
		cfg.Version,
		platform,
		arch,
	)

	//nolint:gosec // controlled url.
	resp, err = http.Get(binaryURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download binary: %w", err)
	}

	defer resp.Body.Close()

	// Create temp file for download
	tmpFile, err := os.CreateTemp("", "contributoor-*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Copy download to temp file
	if _, ioerr := io.Copy(tmpFile, resp.Body); ioerr != nil {
		return fmt.Errorf("failed to write binary to temp file: %w", err)
	}

	// Stop service if running
	running, err := s.IsRunning()
	if err != nil {
		return fmt.Errorf("failed to check if service is running: %w", err)
	}

	if running {
		if err := s.Stop(); err != nil {
			return fmt.Errorf("failed to stop service: %w", err)
		}
	}

	// Extract binary
	if err := os.MkdirAll(binaryDir, 0755); err != nil {
		return fmt.Errorf("failed to create binary directory: %w", err)
	}

	//nolint:gosec // binaryPath is controlled by us.
	cmd := exec.Command("tar", "--no-same-owner", "-xzf", tmpFile.Name(), "-C", binaryDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}

	// Set permissions
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return fmt.Errorf("failed to set binary permissions: %w", err)
	}

	fmt.Printf("%sBinary updated successfully%s\n", tui.TerminalColorGreen, tui.TerminalColorReset)

	// Restart if it was running
	if running {
		if err := s.Start(); err != nil {
			return fmt.Errorf("failed to restart service: %w", err)
		}
	}

	return nil
}
