package sidecar

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/ethpandaops/contributoor-installer/internal/installer"
	"github.com/ethpandaops/contributoor-installer/internal/tui"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
)

// updateInstaller updates the installer binary to the specified version.
func updateInstaller(cfg *config.Config, installerCfg *installer.Config) error {
	releaseDir := filepath.Join(cfg.ContributoorDirectory, "releases", fmt.Sprintf("installer-%s", cfg.Version))
	if err := os.MkdirAll(releaseDir, 0755); err != nil {
		return fmt.Errorf("failed to create release directory: %w", err)
	}

	// Download new version.
	downloadURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/v%s/contributoor-installer_%s_%s_%s.tar.gz",
		installerCfg.GithubOrg,
		installerCfg.GithubInstallerRepo,
		cfg.Version,
		cfg.Version,
		runtime.GOOS,
		runtime.GOARCH,
	)

	//nolint:gosec // controlled url.
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download installer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download installer: HTTP %d", resp.StatusCode)
	}

	// Extract directly to release directory.
	cmd := exec.Command("tar", "--no-same-owner", "-xzf", "-", "-C", releaseDir)
	cmd.Stdin = resp.Body

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract installer: %w", err)
	}

	// Update symlink.
	binPath := filepath.Join(cfg.ContributoorDirectory, "bin")
	newBinary := filepath.Join(releaseDir, "contributoor")
	symlink := filepath.Join(binPath, "contributoor")

	// Set permissions and create symlink.
	if err := os.Chmod(newBinary, 0755); err != nil {
		return fmt.Errorf("failed to set binary permissions: %w", err)
	}

	if err := os.Remove(symlink); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old symlink: %w", err)
	}

	if err := os.Symlink(newBinary, symlink); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	fmt.Printf("%sInstaller updated successfully%s\n", tui.TerminalColorGreen, tui.TerminalColorReset)

	return nil
}
