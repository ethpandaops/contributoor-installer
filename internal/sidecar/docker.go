package sidecar

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/installer"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/sirupsen/logrus"
)

//go:generate mockgen -package mock -destination mock/docker.mock.go github.com/ethpandaops/contributoor-installer/internal/sidecar DockerSidecar

type DockerSidecar interface {
	SidecarRunner
}

// dockerSidecar is a basic service for interacting with the docker container.
type dockerSidecar struct {
	logger           *logrus.Logger
	composePath      string
	composePortsPath string
	configPath       string
	sidecarCfg       ConfigManager
	installerCfg     *installer.Config
}

// NewDockerSidecar creates a new DockerSidecar.
func NewDockerSidecar(logger *logrus.Logger, sidecarCfg ConfigManager, installerCfg *installer.Config) (DockerSidecar, error) {
	composePath, err := findComposeFile()
	if err != nil {
		return nil, fmt.Errorf("failed to find docker-compose.yml: %w", err)
	}

	composePortsPath, err := findComposePortsFile()
	if err != nil {
		return nil, fmt.Errorf("failed to find docker-compose.ports.yml: %w", err)
	}

	if err := validateComposePath(composePath); err != nil {
		return nil, fmt.Errorf("invalid docker-compose file: %w", err)
	}

	if err := validateComposePath(composePortsPath); err != nil {
		return nil, fmt.Errorf("invalid docker-compose.ports file: %w", err)
	}

	return &dockerSidecar{
		logger:           logger,
		composePath:      filepath.Clean(composePath),
		composePortsPath: filepath.Clean(composePortsPath),
		configPath:       sidecarCfg.GetConfigPath(),
		sidecarCfg:       sidecarCfg,
		installerCfg:     installerCfg,
	}, nil
}

// Start starts the docker container using docker-compose.
func (s *dockerSidecar) Start() error {
	// If metrics are enabled, append our ports.yml as an additional -f arg.
	var additionalArgs []string
	if metricsHost, _ := s.sidecarCfg.Get().GetMetricsHostPort(); metricsHost != "" {
		additionalArgs = append(additionalArgs, "-f", s.composePortsPath)
	}

	args := append([]string{"compose", "-f", s.composePath}, additionalArgs...)
	args = append(args, "up", "-d", "--pull", "always")

	cmd := exec.Command("docker", args...)
	cmd.Env = s.getComposeEnv()

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start containers: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("%sContributoor started successfully%s\n", tui.TerminalColorGreen, tui.TerminalColorReset)

	return nil
}

// Stop stops and removes the docker container using docker-compose.
func (s *dockerSidecar) Stop() error {
	// Stop and remove containers, volumes, and networks
	//nolint:gosec // validateComposePath() and filepath.Clean() in-use.
	cmd := exec.Command("docker", "compose", "-f", s.composePath, "down",
		"--remove-orphans",
		"-v",
		"--rmi", "local",
		"--timeout", "30")
	cmd.Env = s.getComposeEnv()

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop containers: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("%sContributoor stopped successfully%s\n", tui.TerminalColorGreen, tui.TerminalColorReset)

	return nil
}

// IsRunning checks if the docker container is running.
func (s *dockerSidecar) IsRunning() (bool, error) {
	//nolint:gosec // validateComposePath() and filepath.Clean() in-use.
	cmd := exec.Command("docker", "compose", "-f", s.composePath, "ps", "--format", "{{.State}}")
	cmd.Env = s.getComposeEnv()

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check container status: %w", err)
	}

	states := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, state := range states {
		if strings.Contains(strings.ToLower(state), "running") {
			return true, nil
		}
	}

	return false, nil
}

// Update pulls the latest image and restarts the container.
func (s *dockerSidecar) Update() error {
	cfg := s.sidecarCfg.Get()

	// Update installer first.
	if err := updateInstaller(cfg, s.installerCfg); err != nil {
		s.logger.Warnf("Failed to update installer: %v", err)
	}

	// Update sidecar.
	if err := s.updateSidecar(); err != nil {
		return fmt.Errorf("failed to update sidecar: %w", err)
	}

	return nil
}

// updateSidecar updates the docker image to the specified version.
func (s *dockerSidecar) updateSidecar() error {
	cfg := s.sidecarCfg.Get()

	image := fmt.Sprintf("%s:%s", s.installerCfg.DockerImage, cfg.Version)

	cmd := exec.Command("docker", "pull", image)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to pull image %s: %w\nOutput: %s", image, err, string(output))
	}

	fmt.Printf(
		"%sContributoor image %s updated successfully%s\n",
		tui.TerminalColorGreen,
		image,
		tui.TerminalColorReset,
	)

	return nil
}

// findComposeFile finds the docker-compose file based on the OS.
func findComposeFile() (string, error) {
	// Get binary directory.
	ex, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not get executable path: %w", err)
	}

	binDir := filepath.Dir(ex)

	// Get the actual binary path (resolve symlink).
	actualBin, err := filepath.EvalSymlinks(ex)
	if err != nil {
		return "", fmt.Errorf("could not resolve symlink: %w", err)
	}

	releaseDir := filepath.Dir(actualBin)

	// First check release directory (next to actual binary).
	composePath := filepath.Join(releaseDir, "docker-compose.yml")
	if _, e := os.Stat(composePath); e == nil {
		return composePath, nil
	}

	// Fallback to bin directory for backward compatibility.
	if _, statErr := os.Stat(filepath.Join(binDir, "docker-compose.yml")); statErr == nil {
		return filepath.Join(binDir, "docker-compose.yml"), nil
	}

	// Try current directory.
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get working directory: %w", err)
	}

	if _, err := os.Stat(filepath.Join(cwd, "docker-compose.yml")); err == nil {
		return filepath.Join(cwd, "docker-compose.yml"), nil
	}

	// Try repo root
	if _, err := os.Stat(filepath.Join(cwd, "..", "..", "docker-compose.yml")); err == nil {
		return filepath.Join(cwd, "..", "..", "docker-compose.yml"), nil
	}

	return "", fmt.Errorf("docker-compose.yml not found")
}

// findComposeFile finds the docker-compose file based on the OS.
func findComposePortsFile() (string, error) {
	// Get binary directory.
	ex, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not get executable path: %w", err)
	}

	binDir := filepath.Dir(ex)

	// Get the actual binary path (resolve symlink).
	actualBin, err := filepath.EvalSymlinks(ex)
	if err != nil {
		return "", fmt.Errorf("could not resolve symlink: %w", err)
	}

	releaseDir := filepath.Dir(actualBin)

	// First check release directory (next to actual binary).
	composePath := filepath.Join(releaseDir, "docker-compose.ports.yml")
	if _, e := os.Stat(composePath); e == nil {
		return composePath, nil
	}

	// Fallback to bin directory for backward compatibility.
	if _, statErr := os.Stat(filepath.Join(binDir, "docker-compose.ports.yml")); statErr == nil {
		return filepath.Join(binDir, "docker-compose.ports.yml"), nil
	}

	// Try current directory.
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get working directory: %w", err)
	}

	if _, err := os.Stat(filepath.Join(cwd, "docker-compose.ports.yml")); err == nil {
		return filepath.Join(cwd, "docker-compose.ports.yml"), nil
	}

	// Try repo root
	if _, err := os.Stat(filepath.Join(cwd, "..", "..", "docker-compose.ports.yml")); err == nil {
		return filepath.Join(cwd, "..", "..", "docker-compose.ports.yml"), nil
	}

	return "", fmt.Errorf("docker-compose.ports.yml not found")
}

func validateComposePath(path string) error {
	// Check if path exists and is a regular file
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("invalid compose file path: %w", err)
	}

	if fi.IsDir() {
		return fmt.Errorf("compose path is a directory, not a file")
	}

	// Ensure absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify file extension
	if !strings.HasSuffix(strings.ToLower(absPath), ".yml") &&
		!strings.HasSuffix(strings.ToLower(absPath), ".yaml") {
		return fmt.Errorf("compose file must have .yml or .yaml extension")
	}

	return nil
}

func (s *dockerSidecar) getComposeEnv() []string {
	cfg := s.sidecarCfg.Get()

	env := append(
		os.Environ(),
		fmt.Sprintf("CONTRIBUTOOR_CONFIG_PATH=%s", filepath.Dir(s.configPath)),
		fmt.Sprintf("CONTRIBUTOOR_VERSION=%s", cfg.Version),
	)

	// Handle metrics address (only added if set).
	if metricsHost, metricsPort := cfg.GetMetricsHostPort(); metricsHost != "" {
		env = append(
			env,
			fmt.Sprintf("CONTRIBUTOOR_METRICS_ADDRESS=%s", metricsHost),
			fmt.Sprintf("CONTRIBUTOOR_METRICS_PORT=%s", metricsPort),
		)
	}

	// Handle pprof address (only added if set).
	if pprofHost, pprofPort := cfg.GetPprofHostPort(); pprofHost != "" {
		env = append(
			env,
			fmt.Sprintf("CONTRIBUTOOR_PPROF_ADDRESS=%s", pprofHost),
			fmt.Sprintf("CONTRIBUTOOR_PPROF_PORT=%s", pprofPort),
		)
	}

	return env
}

// Logs shows the logs from the docker container.
func (s *dockerSidecar) Logs(tailLines int, follow bool) error {
	args := []string{"compose", "-f", s.composePath, "logs"}

	if tailLines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tailLines))
	}

	if follow {
		args = append(args, "-f")
	}

	cmd := exec.Command("docker", args...)
	cmd.Env = s.getComposeEnv()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
