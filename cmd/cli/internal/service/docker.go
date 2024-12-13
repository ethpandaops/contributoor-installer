package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

type DockerService struct {
	logger      *logrus.Logger
	config      *internal.ContributoorConfig
	composePath string
	configPath  string
}

func NewDockerService(logger *logrus.Logger, config *internal.ContributoorConfig) (*DockerService, error) {
	composePath, err := findComposeFile()
	if err != nil {
		return nil, fmt.Errorf("failed to find docker-compose.yml: %w", err)
	}

	configPath, err := expandConfigPath(config.ContributoorDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to expand config path: %w", err)
	}

	return &DockerService{
		logger:      logger,
		config:      config,
		composePath: composePath,
		configPath:  configPath,
	}, nil
}

// Start starts the docker container using docker-compose
func (s *DockerService) Start() error {
	cmd := exec.Command("docker", "compose", "-f", s.composePath, "up", "-d", "--pull", "always")
	cmd.Env = s.getComposeEnv()

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start containers: %w\nOutput: %s", err, string(output))
	}

	s.logger.Info("Service started successfully")
	return nil
}

// Stop stops and removes the docker container using docker-compose
func (s *DockerService) Stop() error {
	// Stop and remove containers, volumes, and networks
	cmd := exec.Command("docker", "compose", "-f", s.composePath, "down",
		"--remove-orphans", // Remove containers not defined in compose
		"-v",               // Remove volumes
		"--rmi", "local",   // Remove local images
		"--timeout", "30") // Wait up to 30 seconds before force killing
	cmd.Env = s.getComposeEnv()

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop containers: %w\nOutput: %s", err, string(output))
	}

	s.logger.Info("Service stopped and cleaned up successfully")
	return nil
}

// IsRunning checks if the docker container is running
func (s *DockerService) IsRunning() (bool, error) {
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

// Update pulls the latest image and restarts the container
func (s *DockerService) Update() error {
	// Pull image
	cmd := exec.Command("docker", "pull", fmt.Sprintf("ethpandaops/contributoor-test:%s", s.config.Version))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to pull image: %w\nOutput: %s", err, string(output))
	}

	s.logger.WithField("version", s.config.Version).Info("Image updated successfully")
	return nil
}

func (s *DockerService) getComposeEnv() []string {
	return append(os.Environ(),
		fmt.Sprintf("CONTRIBUTOOR_CONFIG_PATH=%s", s.configPath),
		fmt.Sprintf("CONTRIBUTOOR_VERSION=%s", s.config.Version),
	)
}

func findComposeFile() (string, error) {
	// Get binary directory
	ex, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not get executable path: %w", err)
	}
	binDir := filepath.Dir(ex)

	// Check release mode (next to binary)
	composePath := filepath.Join(binDir, "docker-compose.yml")
	if _, err := os.Stat(composePath); err == nil {
		return composePath, nil
	}

	// Check dev mode paths
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get working directory: %w", err)
	}

	// Try current directory
	if _, err := os.Stat(filepath.Join(cwd, "docker-compose.yml")); err == nil {
		return filepath.Join(cwd, "docker-compose.yml"), nil
	}

	// Try repo root
	if _, err := os.Stat(filepath.Join(cwd, "..", "..", "docker-compose.yml")); err == nil {
		return filepath.Join(cwd, "..", "..", "docker-compose.yml"), nil
	}

	return "", fmt.Errorf("docker-compose.yml not found")
}

func expandConfigPath(path string) (string, error) {
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Expand home directory
	expandedPath, err := homedir.Expand(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to expand home directory: %w", err)
	}

	return expandedPath, nil
}
