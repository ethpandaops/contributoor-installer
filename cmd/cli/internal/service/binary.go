package service

import (
	"archive/tar"
	"compress/gzip"
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
	"github.com/mitchellh/go-homedir"
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
	binaryPath := filepath.Join(s.config.ContributoorDirectory, "bin", "sentry")
	if _, err := os.Stat(binaryPath); err != nil {
		return fmt.Errorf("binary not found at %s - please reinstall", binaryPath)
	}

	return nil
}

func (s *BinaryService) verifyChecksum(binaryPath string) error {
	checksumsURL := fmt.Sprintf("https://github.com/ethpandaops/contributoor-test/releases/download/v%s/contributoor-test_%s_checksums.txt",
		s.config.Version, s.config.Version)
	s.logger.WithField("url", checksumsURL).Info("Verifying checksum")

	resp, err := http.Get(checksumsURL)
	if err != nil {
		return fmt.Errorf("failed to download checksums file: %w", err)
	}
	defer resp.Body.Close()

	checksumsData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read checksums file: %w", err)
	}

	checksums := strings.Split(string(checksumsData), "\n")
	binaryName := s.getBinaryName() + ".tar.gz"
	expectedChecksum := ""

	for _, line := range checksums {
		if strings.Contains(line, binaryName) {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				expectedChecksum = parts[0]
				break
			}
		}
	}

	if expectedChecksum == "" {
		return fmt.Errorf("checksum for %s not found", binaryName)
	}

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

	s.logger.WithField("url", url).Info("Downloading binary")

	// Create bin directory
	binDir := filepath.Join(s.config.ContributoorDirectory, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Download binary to temp file first
	binaryPath := filepath.Join(binDir, s.getBinaryName()+".tar.gz")
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
	if err := s.verifyChecksum(tempPath); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	// Move to final location
	if err := os.Rename(tempPath, binaryPath); err != nil {
		return fmt.Errorf("failed to move binary to final location: %w", err)
	}

	// Untar the downloaded tar.gz
	if err := s.untar(binaryPath, binDir); err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}

	// Remove the tar.gz file after extraction
	if err := os.Remove(binaryPath); err != nil {
		s.logger.Warnf("Failed to remove tar.gz file: %v", err)
	}

	// Expand the home directory if necessary. Takes care of paths provided with `~`.
	expandedDir, err := homedir.Expand(s.config.ContributoorDirectory)
	if err != nil {
		return fmt.Errorf("failed to expand config path: %w", err)
	}
	configPath := filepath.Join(expandedDir, "contributoor.yaml")

	s.logger.WithField("path", binDir).Info("Binary installed successfully")
	s.logger.WithField("run_cmd", fmt.Sprintf("%s/sentry --config %s", binDir, configPath)).Info("Binary mode is still WIP, please execute run_cmd to start the service")

	return nil
}

func (s *BinaryService) getBinaryName() string {
	return fmt.Sprintf("contributoor-test_%s_%s_%s", s.config.Version, runtime.GOOS, runtime.GOARCH)
}

func (s *BinaryService) getBinaryURL() string {
	arch := runtime.GOARCH
	os := runtime.GOOS
	version := s.config.Version

	return fmt.Sprintf("https://github.com/ethpandaops/contributoor-test/releases/download/v%s/contributoor-test_%s_%s_%s.tar.gz",
		version, version, os, arch)
}

func (s *BinaryService) checkForUpdates() (bool, string, error) {
	currentVersion := s.config.Version

	resp, err := http.Get("https://api.github.com/repos/ethpandaops/contributoor-test/releases/latest")
	if err != nil {
		return false, "", fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	release := &GithubRelease{}
	if err := json.NewDecoder(resp.Body).Decode(release); err != nil {
		return false, "", fmt.Errorf("failed to parse release info: %w", err)
	}

	tag := strings.ReplaceAll(release.TagName, "v", "")

	return tag != currentVersion, tag, nil
}

// untar extracts a tar.gz file to the specified destination directory
func (s *BinaryService) untar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tarReader := tar.NewReader(gzr)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar archive: %w", err)
		}

		targetPath := filepath.Join(dest, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()
		default:
			s.logger.Warnf("Unknown file type in tar archive: %v", header.Typeflag)
		}
	}
	return nil
}
