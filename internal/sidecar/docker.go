package sidecar

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/installer"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/sirupsen/logrus"
)

//go:generate mockgen -package mock -destination mock/docker.mock.go github.com/ethpandaops/contributoor-installer/internal/sidecar DockerSidecar

type DockerSidecar interface {
	SidecarRunner
	GetComposeEnv() []string
}

// dockerSidecar is a basic service for interacting with the docker container.
type dockerSidecar struct {
	logger             *logrus.Logger
	composePath        string
	composeMetricsPath string
	composeHealthPath  string
	composeNetworkPath string
	configPath         string
	sidecarCfg         ConfigManager
	installerCfg       *installer.Config
}

// NewDockerSidecar creates a new DockerSidecar.
func NewDockerSidecar(logger *logrus.Logger, sidecarCfg ConfigManager, installerCfg *installer.Config) (DockerSidecar, error) {
	var (
		composeFilename        = "docker-compose.yml"
		composeMetricsFilename = "docker-compose.metrics.yml"
		composeHealthFilename  = "docker-compose.health.yml"
		composeNetworkFilename = "docker-compose.network.yml"
	)

	composePath, err := findComposeFile(composeFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to find %s: %w", composeFilename, err)
	}

	composeMetricsPath, err := findComposeFile(composeMetricsFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to find %s: %w", composeMetricsFilename, err)
	}

	composeHealthPath, err := findComposeFile(composeHealthFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to find %s: %w", composeHealthFilename, err)
	}

	composeNetworkPath, err := findComposeFile(composeNetworkFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to find %s: %w", composeNetworkFilename, err)
	}

	if err := validateComposePath(composePath); err != nil {
		return nil, fmt.Errorf("invalid %s file: %w", composeFilename, err)
	}

	if err := validateComposePath(composeMetricsPath); err != nil {
		return nil, fmt.Errorf("invalid %s file: %w", composeMetricsPath, err)
	}

	if err := validateComposePath(composeHealthPath); err != nil {
		return nil, fmt.Errorf("invalid %s file: %w", composeHealthPath, err)
	}

	if err := validateComposePath(composeNetworkPath); err != nil {
		return nil, fmt.Errorf("invalid %s file: %w", composeNetworkFilename, err)
	}

	return &dockerSidecar{
		logger:             logger,
		composePath:        filepath.Clean(composePath),
		composeMetricsPath: filepath.Clean(composeMetricsPath),
		composeNetworkPath: filepath.Clean(composeNetworkPath),
		composeHealthPath:  filepath.Clean(composeHealthPath),
		configPath:         sidecarCfg.GetConfigPath(),
		sidecarCfg:         sidecarCfg,
		installerCfg:       installerCfg,
	}, nil
}

// Start starts the docker container using docker-compose.
func (s *dockerSidecar) Start() error {
	// Check if container exists and remove it first, if it does.
	cmd := exec.Command("docker", "ps", "-aq", "-f", "name=contributoor")

	output, err := cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		removeCmd := exec.Command("docker", "rm", "-f", "contributoor")
		if output, err := removeCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to remove existing container: %w\nOutput: %s", err, string(output))
		}
	}

	args := append(s.getComposeArgs(), "up", "-d")

	cmd = exec.Command("docker", args...)
	cmd.Env = s.GetComposeEnv()

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start containers: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("%sContributoor started successfully%s\n", tui.TerminalColorGreen, tui.TerminalColorReset)

	return nil
}

// Stop stops and removes the docker container using docker-compose.
func (s *dockerSidecar) Stop() error {
	// First try to stop via compose. If there has been any sort of configuration change
	// between versions, then this will not stop the container.
	args := append(s.getComposeArgs(), "down",
		"--remove-orphans",
		"-v",
		"--rmi", "local",
		"--timeout", "30")

	cmd := exec.Command("docker", args...)
	cmd.Env = s.GetComposeEnv()

	if output, err := cmd.CombinedOutput(); err != nil {
		// Don't return error here, try our fallback.
		s.logger.Debugf("failed to stop via compose: %v\noutput: %s", err, string(output))
	}

	// Fallback in the case of a configuration change between versions, attempt to remove
	// the container by name.
	cmd = exec.Command("docker", "rm", "-f", "contributoor")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop container: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("%sContributoor stopped successfully%s\n", tui.TerminalColorGreen, tui.TerminalColorReset)

	return nil
}

// Status returns the current state of the docker container.
func (s *dockerSidecar) Status() (string, error) {
	// First check if container exists.
	cmd := exec.Command("docker", "ps", "-a", "--filter", "name=contributoor", "--format", "{{.ID}}")

	output, err := cmd.Output()
	if err != nil || len(strings.TrimSpace(string(output))) == 0 {
		//nolint:nilerr // We don't care about the error here.
		return "not running", nil
	}

	// Container exists, get its status.
	cmd = exec.Command("docker", "inspect", "-f", "{{.State.Status}}", "contributoor")

	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get container status: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// IsRunning checks if the docker container is running.
func (s *dockerSidecar) IsRunning() (bool, error) {
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Status}}", "contributoor")

	output, err := cmd.Output()
	if err != nil {
		//nolint:nilerr // We don't care about the error here.
		return false, nil
	}

	return strings.TrimSpace(string(output)) == "running", nil
}

// Update pulls the latest image and restarts the container.
func (s *dockerSidecar) Update() error {
	cfg := s.sidecarCfg.Get()

	// Update installer first.
	if err := updateInstaller(cfg, s.installerCfg); err != nil {
		return fmt.Errorf("failed to update installer: %w", err)
	}

	// Update sidecar.
	if err := s.updateSidecar(); err != nil {
		return fmt.Errorf("failed to update sidecar: %w", err)
	}

	return nil
}

// updateSidecar updates the docker image to the specified version.
func (s *dockerSidecar) updateSidecar() error {
	var (
		cfg   = s.sidecarCfg.Get()
		image = fmt.Sprintf("%s:%s", s.installerCfg.DockerImage, cfg.Version)
	)

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

// GetComposeEnv returns the environment variables for docker-compose.
func (s *dockerSidecar) GetComposeEnv() []string {
	cfg := s.sidecarCfg.Get()

	env := append(
		os.Environ(),
		fmt.Sprintf("CONTRIBUTOOR_CONFIG_PATH=%s", filepath.Dir(s.configPath)),
		fmt.Sprintf("CONTRIBUTOOR_VERSION=%s", cfg.Version),
	)

	// Add docker network if using docker
	if cfg.RunMethod == config.RunMethod_RUN_METHOD_DOCKER && cfg.DockerNetwork != "" {
		env = append(env, fmt.Sprintf("CONTRIBUTOOR_DOCKER_NETWORK=%s", cfg.DockerNetwork))
	}

	// Handle metrics address (only added if set).
	if metricsHost, metricsPort := cfg.GetMetricsHostPort(); metricsHost != "" {
		env = append(
			env,
			fmt.Sprintf("CONTRIBUTOOR_METRICS_ADDRESS=%s", metricsHost),
			fmt.Sprintf("CONTRIBUTOOR_METRICS_PORT=%s", metricsPort),
		)
	}

	// Handle health address (only added if set).
	if healthHost, healthPort := cfg.GetHealthCheckHostPort(); healthHost != "" {
		env = append(
			env,
			fmt.Sprintf("CONTRIBUTOOR_HEALTH_ADDRESS=%s", healthHost),
			fmt.Sprintf("CONTRIBUTOOR_HEALTH_PORT=%s", healthPort),
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
	args := append(s.getComposeArgs(), "logs")

	if tailLines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tailLines))
	}

	if follow {
		args = append(args, "-f")
	}

	cmd := exec.Command("docker", args...)
	cmd.Env = s.GetComposeEnv()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(s.composePath)

	return cmd.Run()
}

// Version returns the version of the currently running container or local image.
func (s *dockerSidecar) Version() (string, error) {
	// First try to get version from running container.
	cmd := exec.Command("docker", "inspect", "-f", "{{index .Config.Labels \"org.opencontainers.image.version\"}}", "contributoor")

	output, err := cmd.Output()
	if err == nil {
		return strings.TrimPrefix(strings.TrimSpace(string(output)), "v"), nil
	}

	// If container not running, check local image version.
	var (
		cfg   = s.sidecarCfg.Get()
		image = fmt.Sprintf("%s:%s", s.installerCfg.DockerImage, cfg.Version)
	)

	cmd = exec.Command("docker", "inspect", "-f", "{{index .Config.Labels \"org.opencontainers.image.version\"}}", image)

	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get image version: %w", err)
	}

	return strings.TrimPrefix(strings.TrimSpace(string(output)), "v"), nil
}

// getComposeArgs returns the consistent set of compose arguments including file paths.
func (s *dockerSidecar) getComposeArgs() []string {
	var additionalArgs []string

	if metricsHost, _ := s.sidecarCfg.Get().GetMetricsHostPort(); metricsHost != "" {
		additionalArgs = append(additionalArgs, "-f", s.composeMetricsPath)
	}

	if healthHost, _ := s.sidecarCfg.Get().GetHealthCheckHostPort(); healthHost != "" {
		additionalArgs = append(additionalArgs, "-f", s.composeHealthPath)
	}

	if s.sidecarCfg.Get().RunMethod == config.RunMethod_RUN_METHOD_DOCKER && s.sidecarCfg.Get().DockerNetwork != "" {
		additionalArgs = append(additionalArgs, "-f", s.composeNetworkPath)
	}

	return append([]string{"compose", "-f", s.composePath}, additionalArgs...)
}

// findComposeFile finds the docker-compose file based on the OS.
func findComposeFile(filename string) (string, error) {
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
	composePath := filepath.Join(releaseDir, filename)
	if _, e := os.Stat(composePath); e == nil {
		return composePath, nil
	}

	// Fallback to bin directory for backward compatibility.
	if _, statErr := os.Stat(filepath.Join(binDir, filename)); statErr == nil {
		return filepath.Join(binDir, filename), nil
	}

	// Try current directory.
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get working directory: %w", err)
	}

	if _, err := os.Stat(filepath.Join(cwd, filename)); err == nil {
		return filepath.Join(cwd, filename), nil
	}

	// Try repo root
	if _, err := os.Stat(filepath.Join(cwd, "..", "..", filename)); err == nil {
		return filepath.Join(cwd, "..", "..", filename), nil
	}

	return "", fmt.Errorf("%s not found", filename)
}
