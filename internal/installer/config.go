package installer

import "github.com/sirupsen/logrus"

// Config holds installer-specific configuration that isn't exposed to the sidecar.
type Config struct {
	// LogLevel is the log level to use for the installer.
	LogLevel string
	// DockerImage is the image name of the sidecar.
	DockerImage string
	// DockerTag is the tag of the sidecar.
	DockerTag string
	// GithubOrg is the organization name housing the sidecar repository.
	GithubOrg string
	// GithubContributoorRepo is the repository name of the sidecar repository.
	GithubContributoorRepo string
	// GithubInstallerRepo is the repository name of the installer repository.
	GithubInstallerRepo string
}

// NewConfig returns the default installer configuration.
func NewConfig() *Config {
	return &Config{
		LogLevel:               logrus.InfoLevel.String(),
		DockerImage:            "ethpandaops/contributoor",
		GithubOrg:              "ethpandaops",
		GithubContributoorRepo: "contributoor",
		GithubInstallerRepo:    "contributoor-installer",
	}
}
