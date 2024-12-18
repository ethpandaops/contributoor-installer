package service

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

// BinaryService is a basic service for interacting with the contributoor binary.
type BinaryService struct {
	logger *logrus.Logger
	config *ContributoorConfig
	stdout *os.File
	stderr *os.File
}

// NewBinaryService creates a new BinaryService.
func NewBinaryService(logger *logrus.Logger, configService *ConfigService) *BinaryService {
	expandedDir, err := homedir.Expand(configService.Get().ContributoorDirectory)
	if err != nil {
		logger.Errorf("Failed to expand config path: %v", err)
		return &BinaryService{
			logger: logger,
			config: configService.Get(),
		}
	}

	logsDir := filepath.Join(expandedDir, "logs")

	// Open log files
	stdout, err := os.OpenFile(filepath.Join(logsDir, "debug.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logger.Errorf("Failed to open stdout log file: %v", err)
		return &BinaryService{
			logger: logger,
			config: configService.Get(),
		}
	}

	stderr, err := os.OpenFile(filepath.Join(logsDir, "service.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		stdout.Close()
		logger.Errorf("Failed to open stderr log file: %v", err)
		return &BinaryService{
			logger: logger,
			config: configService.Get(),
		}
	}

	return &BinaryService{
		logger: logger,
		config: configService.Get(),
		stdout: stdout,
		stderr: stderr,
	}
}

// Start starts the binary service.
func (s *BinaryService) Start() error {
	binaryPath := filepath.Join(s.config.ContributoorDirectory, "bin", "sentry")
	if _, err := os.Stat(binaryPath); err != nil {
		return fmt.Errorf("binary not found at %s - please reinstall", binaryPath)
	}

	expandedDir, err := homedir.Expand(s.config.ContributoorDirectory)
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

	pidFile := filepath.Join(s.config.ContributoorDirectory, "contributoor.pid")
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644); err != nil {
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

	s.logger.Info("Service started successfully")

	return nil
}

// Stop stops the binary service.
func (s *BinaryService) Stop() error {
	pidFile := filepath.Join(s.config.ContributoorDirectory, "contributoor.pid")
	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("failed to read pid file: %w", err)
	}

	pid := string(pidBytes)
	cmd := exec.Command("kill", pid)
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

	s.logger.Info("Service stopped and cleaned up successfully")

	return nil
}

// IsRunning checks if the binary service is running.
func (s *BinaryService) IsRunning() (bool, error) {
	pidFile := filepath.Join(s.config.ContributoorDirectory, "contributoor.pid")
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		return false, nil
	}

	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		return false, err
	}

	// kill -0 just checks if process exists. It doesn't actually send a
	// signal that affects the process.
	cmd := exec.Command("kill", "-0", string(pidBytes))
	if err := cmd.Run(); err != nil {
		os.Remove(pidFile)

		return false, nil
	}

	return true, nil
}

// Update updates the binary service.
func (s *BinaryService) Update() error {
	expandedDir, err := homedir.Expand(s.config.ContributoorDirectory)
	if err != nil {
		return fmt.Errorf("failed to expand config path: %w", err)
	}

	binaryPath := filepath.Join(expandedDir, "bin", "sentry")
	binaryDir := filepath.Dir(binaryPath)

	// Download and verify checksums
	checksumURL := fmt.Sprintf("https://github.com/ethpandaops/contributoor/releases/download/v%s/contributoor_%s_checksums.txt",
		s.config.Version, s.config.Version)

	resp, err := http.Get(checksumURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download checksums: %w", err)
	}
	defer resp.Body.Close()

	// Determine platform and arch
	platform := runtime.GOOS
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	}

	// Download binary
	binaryURL := fmt.Sprintf("https://github.com/ethpandaops/contributoor/releases/download/v%s/contributoor_%s_%s_%s.tar.gz",
		s.config.Version, s.config.Version, platform, arch)

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
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
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

	cmd := exec.Command("tar", "--no-same-owner", "-xzf", tmpFile.Name(), "-C", binaryDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}

	// Set permissions
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return fmt.Errorf("failed to set binary permissions: %w", err)
	}

	s.logger.WithField("version", s.config.Version).Infof(
		"%sBinary updated successfully%s",
		display.TerminalColorGreen,
		display.TerminalColorReset,
	)

	// Restart if it was running
	if running {
		if err := s.Start(); err != nil {
			return fmt.Errorf("failed to restart service: %w", err)
		}
	}

	return nil
}
