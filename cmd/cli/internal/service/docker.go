package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

type DockerService struct {
	logger *logrus.Logger
	config *internal.ContributoorConfig
}

type DockerImageInfo struct {
	ID      string
	RepoTag string
}

func NewDockerService(logger *logrus.Logger, config *internal.ContributoorConfig) *DockerService {
	return &DockerService{
		logger: logger,
		config: config,
	}
}

func (s *DockerService) Stop() error {
	expandedPath, err := s.getConfigPath()
	if err != nil {
		return err
	}

	dockerComposeFile, err := s.getComposeFilePath()
	if err != nil {
		return err
	}

	logCtx := s.logger.WithFields(logrus.Fields{
		"config_path":  expandedPath,
		"compose_file": dockerComposeFile,
	})
	logCtx.Info("Stopping docker service")

	// Stop existing containers
	if err := s.stopContainers(dockerComposeFile, expandedPath); err != nil {
		return err
	}

	logCtx.Info("Service stopped successfully")
	return nil
}

func (s *DockerService) Start() error {
	if err := s.checkForUpdates(); err != nil {
		s.logger.Warnf("Failed to check for updates: %v", err)
	}

	expandedPath, err := s.getConfigPath()
	if err != nil {
		return err
	}

	dockerComposeFile, err := s.getComposeFilePath()
	if err != nil {
		return err
	}

	logCtx := s.logger.WithFields(logrus.Fields{
		"config_path":  expandedPath,
		"compose_file": dockerComposeFile,
	})
	logCtx.Info("Starting docker service")

	// Start containers
	if err := s.startContainers(dockerComposeFile, expandedPath); err != nil {
		return err
	}

	logCtx.Info("Service started successfully")
	return nil
}

func (s *DockerService) Restart() error {
	if err := s.checkForUpdates(); err != nil {
		s.logger.Warnf("Failed to check for updates: %v", err)
	}

	expandedPath, err := s.getConfigPath()
	if err != nil {
		return err
	}

	dockerComposeFile, err := s.getComposeFilePath()
	if err != nil {
		return err
	}

	logCtx := s.logger.WithFields(logrus.Fields{
		"config_path":  expandedPath,
		"compose_file": dockerComposeFile,
	})

	hasContainer, err := s.hasExistingContainer()
	if err != nil {
		s.logger.Warnf("Failed to check container status: %v", err)
	}

	if hasContainer {
		logCtx.Info("Found existing container, stopping first")
		if err := s.stopContainers(dockerComposeFile, expandedPath); err != nil {
			return err
		}
	} else {
		logCtx.Info("No existing container found, starting fresh")
	}

	// Start containers
	if err := s.startContainers(dockerComposeFile, expandedPath); err != nil {
		return err
	}

	logCtx.Info("Service restarted successfully")
	return nil
}

func (s *DockerService) stopContainers(dockerComposeFile string, cfgPath string) error {
	cmd := exec.Command("docker", "compose", "-f", dockerComposeFile, "down", "-v", "--remove-orphans")
	cmd.Env = append(os.Environ(), fmt.Sprintf("CONTRIBUTOOR_CONFIG_PATH=%s", cfgPath))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker compose down failed: %w, output: %s", err, string(output))
	}
	return nil
}

func (s *DockerService) startContainers(dockerComposeFile string, cfgPath string) error {
	cmd := exec.Command("docker", "compose", "-f", dockerComposeFile, "up", "-d", "--pull", "always")
	cmd.Env = append(os.Environ(), fmt.Sprintf("CONTRIBUTOOR_CONFIG_PATH=%s", cfgPath))
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker compose up failed: %w", err)
	}
	return nil
}

func (s *DockerService) getComposeFilePath() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("could not get current file path")
	}

	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))))
	return filepath.Join(repoRoot, "docker-compose.yml"), nil
}

func (s *DockerService) getConfigPath() (string, error) {
	// Ensure absolute path
	absPath, err := filepath.Abs(s.config.ContributoorDirectory)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Expand home directory if needed
	return homedir.Expand(absPath)
}

func (s *DockerService) checkForUpdates() error {
	var hasUpdate bool

	// Get current image info before pull
	before, err := s.getImageInfo("ethpandaops/contributoor-test:latest")
	if err != nil {
		return err
	}

	// Pull latest image
	cmd := exec.Command("docker", "pull", "ethpandaops/contributoor-test:latest")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull latest image: %w", err)
	}

	// Get image info after pull
	after, err := s.getImageInfo("ethpandaops/contributoor-test:latest")
	if err != nil {
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"before": before.ID,
		"after":  after.ID,
	}).Info("Checking for docker image updates")

	// Compare IDs to detect changes
	hasUpdate = before.ID != after.ID

	if hasUpdate {
		s.logger.Info("New Docker image version available")
	} else {
		s.logger.Info("Docker image is up to date")
	}

	return nil
}

func (s *DockerService) getImageInfo(image string) (*DockerImageInfo, error) {
	cmd := exec.Command("docker", "image", "inspect", "--format", "{{.ID}}", image)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// Image doesn't exist
			return &DockerImageInfo{}, nil
		}
		return nil, fmt.Errorf("failed to inspect image: %w", err)
	}

	return &DockerImageInfo{
		ID: strings.TrimSpace(string(output)),
	}, nil
}

func (s *DockerService) hasExistingContainer() (bool, error) {
	dockerComposeFile, err := s.getComposeFilePath()
	if err != nil {
		return false, fmt.Errorf("failed to get compose file path: %w", err)
	}

	expandedPath, err := s.getConfigPath()
	if err != nil {
		return false, fmt.Errorf("failed to get config path: %w", err)
	}

	cmd := exec.Command("docker", "compose", "-f", dockerComposeFile, "ps", "-a", "--format", "{{.Name}}")
	cmd.Env = append(os.Environ(), fmt.Sprintf("CONTRIBUTOOR_CONFIG_PATH=%s", expandedPath))

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// No containers found
			return false, nil
		}
		return false, fmt.Errorf("failed to check container status: %w", err)
	}

	// Check if any container exists
	return len(strings.TrimSpace(string(output))) > 0, nil
}
