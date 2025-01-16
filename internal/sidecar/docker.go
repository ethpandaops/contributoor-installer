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
}

// dockerSidecar is a basic service for interacting with the docker container.
type dockerSidecar struct {
	logger             *logrus.Logger
	composePath        string
	composePortsPath   string
	composeNetworkPath string
	configPath         string
	sidecarCfg         ConfigManager
	installerCfg       *installer.Config
}

// NewDockerSidecar creates a new DockerSidecar.
func NewDockerSidecar(logger *logrus.Logger, sidecarCfg ConfigManager, installerCfg *installer.Config) (DockerSidecar, error) {
	var (
		composeFilename        = "docker-compose.yml"
		composePortsFilename   = "docker-compose.ports.yml"
		composeNetworkFilename = "docker-compose.network.yml"
	)

	composePath, err := findComposeFile(composeFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to find %s: %w", composeFilename, err)
	}

	composePortsPath, err := findComposeFile(composePortsFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to find %s: %w", composePortsFilename, err)
	}

	composeNetworkPath, err := findComposeFile(composeNetworkFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to find %s: %w", composeNetworkFilename, err)
	}

	if err := validateComposePath(composePath); err != nil {
		return nil, fmt.Errorf("invalid %s file: %w", composeFilename, err)
	}

	if err := validateComposePath(composePortsPath); err != nil {
		return nil, fmt.Errorf("invalid %s file: %w", composePortsFilename, err)
	}

	if err := validateComposePath(composeNetworkPath); err != nil {
		return nil, fmt.Errorf("invalid %s file: %w", composeNetworkFilename, err)
	}

	return &dockerSidecar{
		logger:             logger,
		composePath:        filepath.Clean(composePath),
		composePortsPath:   filepath.Clean(composePortsPath),
		composeNetworkPath: filepath.Clean(composeNetworkPath),
		configPath:         sidecarCfg.GetConfigPath(),
		sidecarCfg:         sidecarCfg,
		installerCfg:       installerCfg,
	}, nil
}

// Start starts the docker container using docker-compose.
func (s *dockerSidecar) Start() error {
	args := append(s.getComposeArgs(), "up", "-d", "--pull", "always")

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
	// First try to stop via compose. If there has been any sort of configuration change
	// between versions, then this will not stop the container.
	args := append(s.getComposeArgs(), "down",
		"--remove-orphans",
		"-v",
		"--rmi", "local",
		"--timeout", "30")

	cmd := exec.Command("docker", args...)
	cmd.Env = s.getComposeEnv()

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

// IsRunning checks if the docker container is running.
func (s *dockerSidecar) IsRunning() (bool, error) {
	// Check via compose first. If there has been any sort of configuration change between
	// versions, then this will return a non running state.
	args := append(s.getComposeArgs(), "ps", "--format", "{{.State}}")
	cmd := exec.Command("docker", args...)
	cmd.Env = s.getComposeEnv()

	output, err := cmd.Output()
	if err == nil {
		states := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, state := range states {
			if strings.Contains(strings.ToLower(state), "running") {
				return true, nil
			}
		}
	}

	// In that case, we will fallback to checking for any container with the name 'contributoor'.
	cmd = exec.Command("docker", "ps", "-q", "-f", "name=contributoor")

	output, err = cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check container status: %w", err)
	}

	return len(strings.TrimSpace(string(output))) > 0, nil
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
	cmd.Env = s.getComposeEnv()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(s.composePath)

	return cmd.Run()
}

// getComposeArgs returns the consistent set of compose arguments including file paths.
func (s *dockerSidecar) getComposeArgs() []string {
	var additionalArgs []string

	if metricsHost, _ := s.sidecarCfg.Get().GetMetricsHostPort(); metricsHost != "" {
		additionalArgs = append(additionalArgs, "-f", s.composePortsPath)
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
