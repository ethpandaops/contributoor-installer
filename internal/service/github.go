package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type GitHubService struct {
	owner string
	repo  string
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

func NewGitHubService(owner, repo string) *GitHubService {
	return &GitHubService{
		owner: owner,
		repo:  repo,
	}
}

// GetLatestVersion returns the latest version tag (e.g., "0.0.1") from GitHub releases
func (s *GitHubService) GetLatestVersion() (string, error) {
	// Fetch releases from GitHub API
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", s.owner, s.repo)
	resp, err := http.Get(url)
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
	var latestVersion string
	var latestParts []int

	for _, release := range releases {
		// Skip empty tags
		if release.TagName == "" {
			continue
		}

		// Parse version parts
		parts := strings.Split(strings.TrimPrefix(release.TagName, "v"), ".")
		if len(parts) != 3 {
			continue
		}

		// Convert parts to integers
		var versionParts []int
		for _, part := range parts {
			num, err := strconv.Atoi(part)
			if err != nil {
				continue
			}
			versionParts = append(versionParts, num)
		}

		// Compare versions
		if len(versionParts) == 3 {
			if latestVersion == "" {
				latestVersion = release.TagName
				latestParts = versionParts
				continue
			}

			// Compare version parts
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

	if latestVersion == "" {
		return "", fmt.Errorf("no valid version tags found")
	}

	return strings.TrimPrefix(latestVersion, "v"), nil
}

// VersionExists checks if a specific version exists in the GitHub releases
func (s *GitHubService) VersionExists(version string) (bool, error) {
	// Fetch releases from GitHub API
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", s.owner, s.repo)
	resp, err := http.Get(url)
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
