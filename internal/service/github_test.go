package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ethpandaops/contributoor-installer/internal/installer"
	"github.com/sirupsen/logrus"
)

func TestGitHubService_GetLatestVersion(t *testing.T) {
	tests := []struct {
		name       string
		releases   string
		wantErr    bool
		wantResult string
	}{
		{
			name: "valid releases",
			releases: `[
				{"tag_name": "v0.0.1"},
				{"tag_name": "v1.2.3"},
				{"tag_name": "v0.0.2"},
				{"tag_name": "v1.0.0"}
			]`,
			wantErr:    false,
			wantResult: "1.2.3",
		},
		{
			name: "invalid version format",
			releases: `[
				{"tag_name": "invalid"},
				{"tag_name": "v0.0.1"}
			]`,
			wantErr:    false,
			wantResult: "0.0.1",
		},
		{
			name:       "empty releases",
			releases:   `[]`,
			wantErr:    true,
			wantResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup a test server to intercept the GitHub API requests. Override the
			// githubAPIHost and validateGitHubURL function for this test.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(tt.releases)); err != nil {
					t.Errorf("failed to write response: %v", err)
				}
			}))
			defer server.Close()

			var (
				host     = githubAPIHost
				validate = validateGitHubURL
			)

			githubAPIHost = strings.TrimPrefix(server.URL, "http://")
			defer func() { githubAPIHost = host }()

			validateGitHubURL = func(owner, repo string) (*url.URL, error) {
				return url.Parse(fmt.Sprintf("%s/repos/%s/%s/releases", server.URL, owner, repo))
			}
			defer func() { validateGitHubURL = validate }()

			svc, err := NewGitHubService(logrus.New(), installer.NewConfig())
			if err != nil {
				t.Errorf("NewGitHubService() error = %v", err)

				return
			}

			got, err := svc.GetLatestVersion()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestVersion() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.wantResult {
				t.Errorf("GetLatestVersion() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func TestGitHubService_VersionExists(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		releases   string
		wantErr    bool
		wantExists bool
	}{
		{
			name:    "version exists",
			version: "1.0.0",
			releases: `[
				{"tag_name": "v0.0.1"},
				{"tag_name": "v1.0.0"}
			]`,
			wantErr:    false,
			wantExists: true,
		},
		{
			name:    "version does not exist",
			version: "2.0.0",
			releases: `[
				{"tag_name": "v0.0.1"},
				{"tag_name": "v1.0.0"}
			]`,
			wantErr:    false,
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup a test server to intercept the GitHub API requests. Override the
			// githubAPIHost and validateGitHubURL function for this test.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(tt.releases)); err != nil {
					t.Errorf("failed to write response: %v", err)
				}
			}))
			defer server.Close()

			var (
				host     = githubAPIHost
				validate = validateGitHubURL
			)

			githubAPIHost = strings.TrimPrefix(server.URL, "http://")
			defer func() { githubAPIHost = host }()

			validateGitHubURL = func(owner, repo string) (*url.URL, error) {
				return url.Parse(fmt.Sprintf("%s/repos/%s/%s/releases", server.URL, owner, repo))
			}
			defer func() { validateGitHubURL = validate }()

			svc, err := NewGitHubService(logrus.New(), installer.NewConfig())
			if err != nil {
				t.Errorf("NewGitHubService() error = %v", err)

				return
			}

			exists, err := svc.VersionExists(tt.version)

			if (err != nil) != tt.wantErr {
				t.Errorf("VersionExists() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if exists != tt.wantExists {
				t.Errorf("VersionExists() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestValidateGitHubURL(t *testing.T) {
	tests := []struct {
		name    string
		owner   string
		repo    string
		wantErr bool
	}{
		{
			name:    "valid owner and repo",
			owner:   "ethpandaops",
			repo:    "contributoor",
			wantErr: false,
		},
		{
			name:    "empty owner",
			owner:   "",
			repo:    "contributoor",
			wantErr: true,
		},
		{
			name:    "invalid characters 1",
			owner:   "test/invalid",
			repo:    "contributoor",
			wantErr: true,
		},
		{
			name:    "invalid characters 2",
			owner:   "test$!invalid",
			repo:    "contributoor",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateGitHubURL(tt.owner, tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGitHubURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
