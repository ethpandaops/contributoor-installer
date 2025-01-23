package sidecar

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer/internal/service"
)

// CheckVersion checks if the current running version needs an update.
// - For "latest" tag, it compares the actual running version with latest available.
// - For specific versions, it compares the config version with latest available.
func CheckVersion(
	runner SidecarRunner,
	github service.GitHubService,
	configVersion string,
) (currentVersion, latestVersion string, needsUpdate bool, err error) {
	latestVersion, err = github.GetLatestVersion()
	if err != nil {
		err = fmt.Errorf("failed to get latest version: %w", err)

		return currentVersion, latestVersion, false, err
	}

	if configVersion == "latest" {
		// For "latest" tag, compare the actual running version with latest available.
		currentVersion, err = runner.Version()
		if err != nil {
			err = fmt.Errorf("failed to get running version: %w", err)

			return currentVersion, latestVersion, false, err
		}

		needsUpdate = currentVersion != latestVersion

		return currentVersion, latestVersion, needsUpdate, nil
	}

	// For specific versions, compare config version with latest
	currentVersion = configVersion
	needsUpdate = currentVersion != latestVersion

	return currentVersion, latestVersion, needsUpdate, nil
}
