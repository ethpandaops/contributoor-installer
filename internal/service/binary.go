package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

type BinaryService struct {
	logger *logrus.Logger
	config *ContributoorConfig
}

func NewBinaryService(logger *logrus.Logger, configService *ConfigService) *BinaryService {
	return &BinaryService{
		logger: logger,
		config: configService.Get(),
	}
}

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

	s.logger.WithField("run_cmd", fmt.Sprintf("%s --config %s", binaryPath, configPath)).Info("Binary mode is still WIP, please execute run_cmd to start the service")

	return nil
}

func (s *BinaryService) Stop() error {
	return nil
}

func (s *BinaryService) Update() error {
	return nil
}
