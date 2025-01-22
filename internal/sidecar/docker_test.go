package sidecar_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/ethpandaops/contributoor-installer/internal/installer"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar/mock"
	"github.com/ethpandaops/contributoor/pkg/config/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/mock/gomock"
)

const composeFile = `
services:
  sentry:
    container_name: contributoor
    image: busybox
    command: ["sh", "-c", "while true; do echo 'Container is running'; sleep 0.1; done"]
    healthcheck:
      test: ["CMD-SHELL", "ps aux | grep -v grep | grep 'sleep' || exit 1"]
      interval: 100ms
      timeout: 100ms
      retries: 2
      start_period: 100ms
`

const composePortsFile = `
services:
  test:
    ports:
      - "9090:9090"
`

const composeNetworkFile = `
services:
  sentry:
    networks:
      - contributoor

networks:
  contributoor:
    name: ${CONTRIBUTOOR_DOCKER_NETWORK}
    external: true
`

// TestDockerService_Integration tests the docker sidecar.
// We use test-containers to boot an instance of docker-in-docker.
// We can then use this to test our docker service in isolation.
// The test uses docker-in-docker to run the tests in a real container targeting busybox.
func TestDockerService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup our test environment.
	var (
		ctx    = context.Background()
		port   = 2375
		tmpDir = t.TempDir()
		logger = logrus.New()
		cfg    = &config.Config{
			Version:               "",
			ContributoorDirectory: tmpDir,
			RunMethod:             config.RunMethod_RUN_METHOD_DOCKER,
		}
	)

	// Create mock config manager.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInstallerConfig := installer.NewConfig()
	mockSidecarConfig := mock.NewMockConfigManager(ctrl)
	mockSidecarConfig.EXPECT().Get().Return(cfg).AnyTimes()
	mockSidecarConfig.EXPECT().GetConfigPath().Return(filepath.Join(tmpDir, "config.yaml")).AnyTimes()

	// Write out compose files first.
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte(composeFile), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docker-compose.metrics.yml"), []byte(composePortsFile), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docker-compose.health.yml"), []byte(composePortsFile), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docker-compose.network.yml"), []byte(composeNetworkFile), 0644))

	// Change working directory to our test directory before creating DockerSidecar.
	require.NoError(t, os.Chdir(tmpDir))

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "docker:dind",
			ExposedPorts: []string{fmt.Sprintf("%d/tcp", port)},
			Privileged:   true,
			WaitingFor: wait.ForAll(
				wait.ForLog("Daemon has completed initialization").WithStartupTimeout(30*time.Second),
				wait.ForListeningPort(nat.Port(fmt.Sprintf("%d/tcp", port))).WithStartupTimeout(30*time.Second),
			),
			Env: map[string]string{
				"DOCKER_TLS_CERTDIR": "",
				"DOCKER_HOST":        fmt.Sprintf("tcp://0.0.0.0:%d", port),
			},
			Cmd: []string{
				"--host", fmt.Sprintf("tcp://0.0.0.0:%d", port),
				"--tls=false",
			},
		},
		Started: true,
	})
	require.NoError(t, err)

	defer func() {
		if terr := container.Terminate(ctx); terr != nil {
			t.Logf("failed to terminate container: %v", terr)
		}
	}()

	// Get docker daemon address.
	containerPort, err := container.MappedPort(ctx, nat.Port(fmt.Sprintf("%d/tcp", port)))
	require.NoError(t, err)

	// Create docker service with mock config.
	ds, err := sidecar.NewDockerSidecar(logger, mockSidecarConfig, mockInstallerConfig)
	require.NoError(t, err)

	// Set docker host to test container.
	t.Setenv("DOCKER_HOST", fmt.Sprintf("tcp://localhost:%s", containerPort.Port()))
	t.Setenv("CONTRIBUTOOR_CONFIG_PATH", tmpDir)

	// Helper function for container health check.
	checkContainerHealth := func(t *testing.T, ds sidecar.DockerSidecar, expectRunning bool) {
		t.Helper()
		// Single check if container is running - docker-compose defines a healthcheck which will cover us.
		running, err := ds.IsRunning()
		require.NoError(t, err)
		require.Equal(t, expectRunning, running, "Container running state does not match expected state")
	}

	verifyContainerNetwork := func(t *testing.T, ds sidecar.DockerSidecar, network string) {
		t.Helper()
		cmd := exec.Command("docker", "container", "inspect", "--format", "{{range $net,$v := .NetworkSettings.Networks}}{{printf \"%s\" $net}}{{end}}", "contributoor")
		output, err := cmd.Output()
		require.NoError(t, err)
		require.Contains(t, string(output), network, "Container is connected to incorrect network")
	}

	t.Run("lifecycle_without_metrics", func(t *testing.T) {
		// Boot up the container.
		require.NoError(t, ds.Start())
		checkContainerHealth(t, ds, true)

		// Verify the container is using the default network.
		verifyContainerNetwork(t, ds, "_default")

		// Cleanup
		require.NoError(t, ds.Stop())
		checkContainerHealth(t, ds, false)
	})
	t.Run("lifecycle_with_metrics", func(t *testing.T) {
		cfgWithMetrics := &config.Config{
			Version:               "latest",
			ContributoorDirectory: tmpDir,
			RunMethod:             config.RunMethod_RUN_METHOD_DOCKER,
			MetricsAddress:        "0.0.0.0:9090",
		}

		mockSidecarConfig.EXPECT().Get().Return(cfgWithMetrics).AnyTimes()

		// Boot up the container.
		require.NoError(t, ds.Start())
		checkContainerHealth(t, ds, true)

		// Verify the container is using the default network.
		verifyContainerNetwork(t, ds, "_default")

		// Cleanup.
		require.NoError(t, ds.Stop())
		checkContainerHealth(t, ds, false)
	})

	t.Run("lifecycle_with_external_container", func(t *testing.T) {
		// Start a container directly with docker (not via compose) of the same name. This mimics
		// a container the installer isn't aware of.
		cmd := exec.Command("docker", "run", "-d", "--name", "contributoor", "busybox",
			"sh", "-c", "while true; do echo 'Container is running'; sleep 1; done")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "failed to start container: %s", string(output))

		// IsRunning should detect the external container.
		checkContainerHealth(t, ds, true)

		// Stop should be able to handle the external container.
		require.NoError(t, ds.Stop())

		// Verify container is stopped.
		checkContainerHealth(t, ds, false)

		// Finally, test normal compose lifecycle works after cleaning up external container.
		require.NoError(t, ds.Start())
		checkContainerHealth(t, ds, true)

		// Verify the container is using the default network.
		verifyContainerNetwork(t, ds, "_default")

		// Cleanup.
		require.NoError(t, ds.Stop())
		checkContainerHealth(t, ds, false)
	})

	t.Run("lifecycle_with_custom_network", func(t *testing.T) {
		// Create a custom network first.
		customNetwork := "test_network"
		cmd := exec.Command("docker", "network", "create", customNetwork)
		require.NoError(t, cmd.Run())
		defer exec.Command("docker", "network", "rm", customNetwork).Run() //nolint:errcheck // test.

		cfgWithNetwork := &config.Config{
			Version:               "latest",
			ContributoorDirectory: tmpDir,
			RunMethod:             config.RunMethod_RUN_METHOD_DOCKER,
			DockerNetwork:         customNetwork,
		}

		// Create new mock and DockerSidecar instance for this test.
		mockSidecarConfigCustom := mock.NewMockConfigManager(ctrl)
		mockSidecarConfigCustom.EXPECT().Get().Return(cfgWithNetwork).AnyTimes()
		mockSidecarConfigCustom.EXPECT().GetConfigPath().Return(filepath.Join(tmpDir, "config.yaml")).AnyTimes()

		dsCustom, err := sidecar.NewDockerSidecar(logger, mockSidecarConfigCustom, mockInstallerConfig)
		require.NoError(t, err)

		// Boot up the container.
		require.NoError(t, dsCustom.Start())
		checkContainerHealth(t, dsCustom, true)

		// Verify the container is using the custom network.
		verifyContainerNetwork(t, dsCustom, customNetwork)

		// Cleanup.
		require.NoError(t, dsCustom.Stop())
		checkContainerHealth(t, dsCustom, false)
	})
}

func TestGetComposeEnv(t *testing.T) {
	logger := logrus.New()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tests := []struct {
		name            string
		config          *config.Config
		expectedEnvVars map[string]string
	}{
		{
			name: "basic config",
			config: &config.Config{
				Version:               "latest",
				ContributoorDirectory: t.TempDir(),
				RunMethod:             config.RunMethod_RUN_METHOD_DOCKER,
			},
			expectedEnvVars: map[string]string{
				"CONTRIBUTOOR_VERSION": "latest",
			},
		},
		{
			name: "with metrics",
			config: &config.Config{
				Version:               "v1.0.0",
				ContributoorDirectory: t.TempDir(),
				RunMethod:             config.RunMethod_RUN_METHOD_DOCKER,
				MetricsAddress:        "0.0.0.0:9090",
			},
			expectedEnvVars: map[string]string{
				"CONTRIBUTOOR_VERSION":         "v1.0.0",
				"CONTRIBUTOOR_METRICS_ADDRESS": "0.0.0.0",
				"CONTRIBUTOOR_METRICS_PORT":    "9090",
			},
		},
		{
			name: "with docker network",
			config: &config.Config{
				Version:               "v1.0.0",
				ContributoorDirectory: t.TempDir(),
				RunMethod:             config.RunMethod_RUN_METHOD_DOCKER,
				DockerNetwork:         "custom_network",
			},
			expectedEnvVars: map[string]string{
				"CONTRIBUTOOR_VERSION":        "v1.0.0",
				"CONTRIBUTOOR_DOCKER_NETWORK": "custom_network",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSidecarConfig := mock.NewMockConfigManager(mockCtrl)
			mockInstallerConfig := installer.NewConfig()

			// Write out compose files first
			require.NoError(t, os.WriteFile(filepath.Join(tt.config.ContributoorDirectory, "docker-compose.yml"), []byte(composeFile), 0644))
			require.NoError(t, os.WriteFile(filepath.Join(tt.config.ContributoorDirectory, "docker-compose.metrics.yml"), []byte(composePortsFile), 0644))
			require.NoError(t, os.WriteFile(filepath.Join(tt.config.ContributoorDirectory, "docker-compose.health.yml"), []byte(composePortsFile), 0644))
			require.NoError(t, os.WriteFile(filepath.Join(tt.config.ContributoorDirectory, "docker-compose.network.yml"), []byte(composeNetworkFile), 0644))

			// Change working directory to test directory
			require.NoError(t, os.Chdir(tt.config.ContributoorDirectory))

			// Set up mock expectations before creating DockerSidecar
			mockSidecarConfig.EXPECT().Get().Return(tt.config).AnyTimes()
			mockSidecarConfig.EXPECT().GetConfigPath().Return(filepath.Join(tt.config.ContributoorDirectory, "config.yaml")).AnyTimes()

			ds, err := sidecar.NewDockerSidecar(logger, mockSidecarConfig, mockInstallerConfig)
			require.NoError(t, err)

			env := ds.GetComposeEnv()
			require.NotNil(t, env)

			// Convert env slice to map for easier testing
			envMap := make(map[string]string)
			for _, e := range env {
				parts := strings.SplitN(e, "=", 2)
				if len(parts) == 2 {
					envMap[parts[0]] = parts[1]
				}
			}

			// Check config path is set and points to correct directory
			configPath := envMap["CONTRIBUTOOR_CONFIG_PATH"]
			require.Equal(t, filepath.Dir(filepath.Join(tt.config.ContributoorDirectory, "config.yaml")), configPath)

			// Check all expected env vars are present with correct values
			for k, v := range tt.expectedEnvVars {
				require.Equal(t, v, envMap[k], "Environment variable %s has incorrect value", k)
			}
		})
	}
}
