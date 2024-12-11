package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ethpandaops/contributoor-installer-test/cmd/cli/internal"
	"github.com/sirupsen/logrus"
)

type BinaryService struct {
	logger *logrus.Logger
	config *internal.ContributoorConfig
}

type GithubRelease struct {
	TagName string `json:"tag_name"`
}

func NewBinaryService(logger *logrus.Logger, cfg *internal.ContributoorConfig) *BinaryService {
	return &BinaryService{
		logger: logger,
		config: cfg,
	}
}

func (s *BinaryService) Start() error {
	hasUpdate, latestVersion, err := s.CheckForUpdates()
	if err != nil {
		s.logger.Warnf("Failed to check for updates: %v", err)
	} else if hasUpdate {
		s.logger.Infof("New version %s available", latestVersion)
	}

	binaryPath := filepath.Join(s.config.ContributoorDirectory, "bin", getBinaryName())
	if _, err := os.Stat(binaryPath); err == nil {
		s.logger.Info("Binary already exists, skipping download")
		return nil
	}

	return s.downloadAndInstallBinary()
}

func (s *BinaryService) verifyChecksum(binaryPath string, expectedChecksum string) error {
	file, err := os.Open(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to open binary for verification: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	actualChecksum := hex.EncodeToString(hash.Sum(nil))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

func (s *BinaryService) downloadAndInstallBinary() error {
	url := s.getBinaryURL()
	checksumURL := url + ".sha256"
	s.logger.WithField("url", url).Info("Downloading binary")

	// Download checksum first
	checksumResp, err := http.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("failed to download checksum: %w", err)
	}
	defer checksumResp.Body.Close()

	checksumBytes, err := io.ReadAll(checksumResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read checksum: %w", err)
	}
	expectedChecksum := strings.TrimSpace(string(checksumBytes))

	// Create bin directory
	binDir := filepath.Join(s.config.ContributoorDirectory, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Download binary to temp file first
	binaryPath := filepath.Join(binDir, getBinaryName())
	tempPath := binaryPath + ".tmp"
	out, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempPath) // Clean up temp file

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer resp.Body.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write binary: %w", err)
	}
	out.Close()

	// Verify checksum
	if err := s.verifyChecksum(tempPath, expectedChecksum); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	// Move to final location
	if err := os.Rename(tempPath, binaryPath); err != nil {
		return fmt.Errorf("failed to move binary to final location: %w", err)
	}

	s.logger.WithField("path", binaryPath).Info("Binary installed successfully")
	return nil
}

func getBinaryName() string {
	return "contributoor"
}

func (s *BinaryService) getBinaryURL() string {
	arch := runtime.GOARCH
	os := runtime.GOOS
	version := s.config.Version

	return fmt.Sprintf("https://github.com/ethpandaops/contributoor-installer-test/releases/download/%s/contributoor-%s-%s",
		version, os, arch)
}

func (s *BinaryService) CheckForUpdates() (bool, string, error) {
	currentVersion := "v0.1.0" // TODO: Get from build info

	resp, err := http.Get("https://api.github.com/repos/ethpandaops/contributoor/releases/latest")
	if err != nil {
		return false, "", fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return false, "", fmt.Errorf("failed to parse release info: %w", err)
	}

	return release.TagName != currentVersion, release.TagName, nil
}
