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
	"strings"

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
	if err := s.checkBinaryExists(); err != nil {
		return wrapNotInstalledError(err, "binary")
	}

	if err := s.checkBinaryVersion(); err != nil {
		return fmt.Errorf("version check failed: %w", err)
	}

	cfg := s.sidecarCfg.Get()

	// Use symlink path instead of direct binary path
	binaryPath := filepath.Join(cfg.ContributoorDirectory, "bin", "sentry")
	if _, err := os.Stat(binaryPath); err != nil {
		return fmt.Errorf("binary symlink not found at %s - please reinstall", binaryPath)
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
	if err := s.checkBinaryExists(); err != nil {
		return wrapNotInstalledError(err, "binary")
	}

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
	if err := s.checkBinaryExists(); err != nil {
		return wrapNotInstalledError(err, "binary")
	}

	cfg := s.sidecarCfg.Get()

	// Update installer first
	if err := updateInstaller(cfg, s.installerCfg); err != nil {
		return fmt.Errorf("failed to update installer: %w", err)
	}

	// Update sidecar
	if err := s.updateSidecar(); err != nil {
		return fmt.Errorf("failed to update sidecar: %w", err)
	}

	return nil
}

// Logs shows the logs from the binary sidecar.
func (s *binarySidecar) Logs(tailLines int, follow bool) error {
	if err := s.checkBinaryExists(); err != nil {
		return wrapNotInstalledError(err, "binary")
	}

	cfg := s.sidecarCfg.Get()

	expandedDir, err := homedir.Expand(cfg.ContributoorDirectory)
	if err != nil {
		return fmt.Errorf("failed to expand config path: %w", err)
	}

	logFile := filepath.Join(expandedDir, "logs", "debug.log")

	args := []string{}

	if follow {
		args = append(args, "-f")
	}

	if tailLines > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", tailLines))
	}

	args = append(args, logFile)

	cmd := exec.Command("tail", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// updateSidecar updates the sidecar binary to the specified version.
func (s *binarySidecar) updateSidecar() error {
	cfg := s.sidecarCfg.Get()

	expandedDir, err := homedir.Expand(cfg.ContributoorDirectory)
	if err != nil {
		return fmt.Errorf("failed to expand config path: %w", err)
	}

	// Define both symlink and release paths
	symlinkPath := filepath.Join(expandedDir, "bin", "sentry")
	releaseDir := filepath.Join(expandedDir, "releases", fmt.Sprintf("contributoor-%s", cfg.Version))
	releaseBinaryPath := filepath.Join(releaseDir, "sentry")

	// Download and verify checksums.
	checksumURL := fmt.Sprintf(
		"https://github.com/%s/%s/releases/download/v%s/contributoor_%s_checksums.txt",
		s.installerCfg.GithubOrg,
		s.installerCfg.GithubContributoorRepo,
		cfg.Version,
		cfg.Version,
	)

	//nolint:gosec // controlled url.
	resp, err := http.Get(checksumURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download checksums: %w", err)
	}

	defer resp.Body.Close()

	// Determine platform and arch.
	var (
		platform = runtime.GOOS
		arch     = runtime.GOARCH
	)

	binaryURL := fmt.Sprintf(
		"https://github.com/%s/%s/releases/download/v%s/contributoor_%s_%s_%s.tar.gz",
		s.installerCfg.GithubOrg,
		s.installerCfg.GithubContributoorRepo,
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

	// Copy download to temp file.
	if _, ioerr := io.Copy(tmpFile, resp.Body); ioerr != nil {
		return fmt.Errorf("failed to write binary to temp file: %w", err)
	}

	// Stop service if running.
	running, err := s.IsRunning()
	if err != nil {
		return fmt.Errorf("failed to check if service is running: %w", err)
	}

	if running {
		if err := s.Stop(); err != nil {
			return fmt.Errorf("failed to stop service: %w", err)
		}
	}

	// Create release directory.
	if err := os.MkdirAll(releaseDir, 0755); err != nil {
		return fmt.Errorf("failed to create release directory: %w", err)
	}

	// Extract binary to release directory.
	cmd := exec.Command("tar", "--no-same-owner", "-xzf", tmpFile.Name(), "-C", releaseDir) //nolint:gosec // controlled extraction.
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}

	// Set permissions on release binary.
	if err := os.Chmod(releaseBinaryPath, 0755); err != nil {
		return fmt.Errorf("failed to set binary permissions: %w", err)
	}

	// Update symlink.
	if err := os.Remove(symlinkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old symlink: %w", err)
	}

	if err := os.Symlink(releaseBinaryPath, symlinkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Restart if it was running.
	if running {
		if err := s.Start(); err != nil {
			return fmt.Errorf("failed to restart service: %w", err)
		}
	}

	return nil
}

// checkBinaryExists checks if the binary exists.
func (s *binarySidecar) checkBinaryExists() error {
	cfg := s.sidecarCfg.Get()

	expandedDir, err := homedir.Expand(cfg.ContributoorDirectory)
	if err != nil {
		return fmt.Errorf("failed to expand config path: %w", err)
	}

	binaryPath := filepath.Join(expandedDir, "bin", "sentry")

	if _, err := os.Stat(binaryPath); err != nil {
		return fmt.Errorf("binary not found: %w", err)
	}

	return nil
}

// checkBinaryVersion checks if the binary version matches the config version.
func (s *binarySidecar) checkBinaryVersion() error {
	version, err := s.getBinaryVersion()
	if err != nil {
		return fmt.Errorf("failed to check binary version: %w", err)
	}

	cfg := s.sidecarCfg.Get()
	if version != cfg.Version {
		fmt.Printf(
			"%sVersion mismatch detected: binary is %s but config expects %s. Auto-updating...%s\n",
			tui.TerminalColorYellow,
			version,
			cfg.Version,
			tui.TerminalColorReset,
		)

		if err := s.Update(); err != nil {
			return fmt.Errorf("failed to auto-update binary: %w", err)
		}
	}

	return nil
}

// getBinaryVersion gets the version of the binary by running it with --release flag.
func (s *binarySidecar) getBinaryVersion() (string, error) {
	var (
		cfg        = s.sidecarCfg.Get()
		binaryPath = filepath.Join(cfg.ContributoorDirectory, "bin", "sentry")
		cmd        = exec.Command(binaryPath, "--release")
	)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get binary version: %w", err)
	}

	// Output will be in format vx.y.z, we want to strip the v prefix.
	return strings.TrimPrefix(strings.TrimSpace(string(output)), "v"), nil
}
