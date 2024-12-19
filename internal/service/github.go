package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ethpandaops/contributoor-installer/internal/installer"
	"github.com/sirupsen/logrus"
)

var (
	githubAPIHost     = "api.github.com"
	validateGitHubURL = func(owner, repo string) (*url.URL, error) {
		if owner == "" || repo == "" {
			return nil, errors.New("owner and repo cannot be empty")
		}

		if strings.ContainsAny(owner+repo, "/?#[]@!$&'()*+,;=") {
			return nil, errors.New("invalid owner or repo name")
		}

		urlStr := fmt.Sprintf("https://%s/repos/%s/%s/releases", githubAPIHost, owner, repo)

		u, err := url.Parse(urlStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}

		if u.Host != githubAPIHost {
			return nil, fmt.Errorf("invalid GitHub API host: %s", u.Host)
		}

		return u, nil
	}
)

//go:generate mockgen -package mock -destination mock/github.mock.go github.com/ethpandaops/contributoor-installer/internal/service GitHubService

// GitHubService defines the interface for GitHub operations.
type GitHubService interface {
	// GetLatestVersion returns the latest version tag (e.g., "0.0.1") from GitHub releases.
	GetLatestVersion() (string, error)

	// VersionExists checks if a specific version exists in the GitHub releases.
	VersionExists(version string) (bool, error)
}

// GitHubRelease is a struct that represents a GitHub release.
type GitHubRelease struct {
	TagName string `json:"tag_name"` //nolint:tagliatelle // Upstream response doesnt camelCase.
}

// githubService is a basic service for interacting with the GitHub API.
type githubService struct {
	log          *logrus.Logger
	client       *http.Client
	githubURL    *url.URL
	installerCfg *installer.Config
}

// NewGitHubService creates a new GitHubService.
func NewGitHubService(log *logrus.Logger, installerCfg *installer.Config) (GitHubService, error) {
	githubURL, err := validateGitHubURL(installerCfg.GithubOrg, installerCfg.GithubRepo)
	if err != nil {
		return nil, fmt.Errorf("invalid github url: %w", err)
	}

	return &githubService{
		log:          log,
		installerCfg: installerCfg,
		githubURL:    githubURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// GetLatestVersion returns the latest version tag (e.g., "0.0.1") from GitHub releases.
func (s *githubService) GetLatestVersion() (string, error) {
	resp, err := s.client.Get(s.githubURL.String())
	if err != nil {
		return "", fmt.Errorf("failed to fetch releases: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", fmt.Errorf("failed to parse releases response: %w", err)
	}

	// Find highest version tag
	var (
		latestVersion string
		latestParts   []int
	)

	// Iterate over releases and find the highest semver version.
	for _, release := range releases {
		if release.TagName == "" {
			continue
		}

		parts := strings.Split(strings.TrimPrefix(release.TagName, "v"), ".")
		if len(parts) != 3 {
			continue
		}

		// Convert versions from GH to ints for comparison.
		var versionParts []int

		for _, part := range parts {
			num, err := strconv.Atoi(part)
			if err != nil {
				continue
			}

			versionParts = append(versionParts, num)
		}

		if len(versionParts) == 3 {
			// If we don't have a latest version, set it.
			if latestVersion == "" {
				latestVersion = release.TagName
				latestParts = versionParts

				continue
			}

			// Now we compare the version parts to find the highest version.
			for i := 0; i < 3; i++ {
				if versionParts[i] > latestParts[i] {
					latestVersion = release.TagName
					latestParts = versionParts

					break
				} else if versionParts[i] < latestParts[i] {
					break
				}
			}
		}
	}

	// Something's cooked if we don't have a latest version.
	if latestVersion == "" {
		return "", fmt.Errorf("no valid version tags found")
	}

	// Return the version without the 'v' prefix.
	return strings.TrimPrefix(latestVersion, "v"), nil
}

// VersionExists checks if a specific version exists in the GitHub releases.
func (s *githubService) VersionExists(version string) (bool, error) {
	resp, err := s.client.Get(s.githubURL.String())
	if err != nil {
		return false, fmt.Errorf("failed to fetch releases: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return false, fmt.Errorf("failed to parse releases response: %w", err)
	}

	// Add 'v' prefix if not present
	searchVersion := version
	if !strings.HasPrefix(searchVersion, "v") {
		searchVersion = "v" + searchVersion
	}

	// Look for exact match
	for _, release := range releases {
		if release.TagName == searchVersion {
			return true, nil
		}
	}

	return false, nil
}
