package sidecar

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ethpandaops/contributoor-installer/internal/installer"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateInstaller(t *testing.T) {
	// Setup a mock tar.gz file.
	mockTarGz := createMockTarGz(t, "contributoor", []byte("mock-installer-binary"))

	tests := []struct {
		name         string
		cfg          *config.Config
		installerCfg *installer.Config
		serverFunc   func() *httptest.Server
		wantErr      bool
		errContains  string
	}{
		{
			name: "successful update",
			cfg: &config.Config{
				ContributoorDirectory: "",
				Version:               "0.0.1",
			},
			installerCfg: &installer.Config{
				GithubOrg:           "ethpandaops",
				GithubInstallerRepo: "contributoor-installer",
			},
			serverFunc: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if _, err := w.Write(mockTarGz); err != nil {
						t.Fatalf("failed to write mock tar.gz: %v", err)
					}
				}))
			},
			wantErr: false,
		},
		{
			name: "404 error",
			cfg: &config.Config{
				ContributoorDirectory: "",
				Version:               "0.0.1",
			},
			installerCfg: &installer.Config{
				GithubOrg:           "ethpandaops",
				GithubInstallerRepo: "contributoor-installer",
			},
			serverFunc: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			wantErr:     true,
			errContains: "HTTP 404",
		},
		{
			name: "invalid directory",
			cfg: &config.Config{
				ContributoorDirectory: "/nonexistent/directory",
				Version:               "0.0.1",
			},
			installerCfg: &installer.Config{
				GithubOrg:           "ethpandaops",
				GithubInstallerRepo: "contributoor-installer",
			},
			serverFunc: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if _, err := w.Write(mockTarGz); err != nil {
						t.Fatalf("failed to write mock tar.gz: %v", err)
					}
				}))
			},
			wantErr:     true,
			errContains: "failed to create release directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new temporary directory for each test.
			if tt.cfg.ContributoorDirectory != "/nonexistent/directory" {
				tempDir, err := os.MkdirTemp("", "installer-test-*")
				require.NoError(t, err)

				defer os.RemoveAll(tempDir)

				// Create bin directory.
				require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "bin"), 0755))

				// Use the temp directory as the Contributoor directory for the test.
				tt.cfg.ContributoorDirectory = tempDir
			}

			// Start test server.
			server := tt.serverFunc()
			defer server.Close()

			// Override the download URL to use our test server.
			origURL := fmt.Sprintf(
				"https://github.com/%s/%s/releases/download/v%s/contributoor-installer_%s_%s_%s.tar.gz",
				tt.installerCfg.GithubOrg,
				tt.installerCfg.GithubInstallerRepo,
				tt.cfg.Version,
				tt.cfg.Version,
				runtime.GOOS,
				runtime.GOARCH,
			)

			// Patch http.Get to redirect GitHub URLs to our test server.
			oldClient := http.DefaultClient
			http.DefaultClient = &http.Client{
				Transport: &mockTransport{
					origURL:    origURL,
					mockURL:    server.URL,
					origClient: oldClient.Transport,
				},
			}

			defer func() { http.DefaultClient = oldClient }()

			err := updateInstaller(tt.cfg, tt.installerCfg)

			if tt.wantErr {
				require.Error(t, err)

				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)

				// Verify the binary was created and is executable.
				binPath := filepath.Join(tt.cfg.ContributoorDirectory, "bin", "contributoor")
				_, err := os.Stat(binPath)
				assert.NoError(t, err)

				// Verify it's a symlink.
				fi, err := os.Lstat(binPath)
				assert.NoError(t, err)
				assert.True(t, fi.Mode()&os.ModeSymlink != 0)

				// Verify the symlink points to the correct release.
				target, err := os.Readlink(binPath)
				assert.NoError(t, err)
				assert.Contains(t, target, fmt.Sprintf("installer-%s", tt.cfg.Version))
			}
		})
	}
}

// mockTransport redirects GitHub URLs to our test server.
type mockTransport struct {
	origURL    string
	mockURL    string
	origClient http.RoundTripper
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.String() == t.origURL {
		newReq := *req

		var err error

		newReq.URL, err = req.URL.Parse(t.mockURL)
		if err != nil {
			return nil, err
		}

		if t.origClient == nil {
			t.origClient = http.DefaultTransport
		}

		return t.origClient.RoundTrip(&newReq)
	}

	if t.origClient == nil {
		t.origClient = http.DefaultTransport
	}

	return t.origClient.RoundTrip(req)
}

// createMockTarGz creates a mock tar.gz file for testing.
func createMockTarGz(t *testing.T, filename string, data []byte) []byte {
	t.Helper()

	var (
		buf bytes.Buffer
		gw  = gzip.NewWriter(&buf)
		tw  = tar.NewWriter(gw)
		hdr = &tar.Header{
			Name: filename,
			Mode: 0755,
			Size: int64(len(data)),
		}
	)

	err := tw.WriteHeader(hdr)
	require.NoError(t, err)

	_, err = tw.Write(data)
	require.NoError(t, err)

	err = tw.Close()
	require.NoError(t, err)

	err = gw.Close()
	require.NoError(t, err)

	return buf.Bytes()
}
