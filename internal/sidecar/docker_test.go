package sidecar_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
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
  test:
    image: busybox
    command: ["sh", "-c", "while true; do echo 'Container is running'; sleep 1; done"]
    healthcheck:
      test: ["CMD-SHELL", "ps aux | grep -v grep | grep 'sleep' || exit 1"]
      interval: 1s
      timeout: 1s
      retries: 3
      start_period: 1s
`

const composePortsFile = `
services:
  test:
    ports:
      - "9090:9090"
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
			Version:               "latest",
			ContributoorDirectory: tmpDir,
			RunMethod:             config.RunMethod_RUN_METHOD_DOCKER,
		}
	)

	// Create mock config manager
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInstallerConfig := installer.NewConfig()
	mockSidecarConfig := mock.NewMockConfigManager(ctrl)
	mockSidecarConfig.EXPECT().Get().Return(cfg).AnyTimes()
	mockSidecarConfig.EXPECT().GetConfigPath().Return(filepath.Join(tmpDir, "config.yaml")).AnyTimes()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "docker:dind",
			ExposedPorts: []string{fmt.Sprintf("%d/tcp", port)},
			Privileged:   true,
			WaitingFor: wait.ForAll(
				wait.ForLog("Daemon has completed initialization").WithStartupTimeout(2*time.Minute),
				wait.ForListeningPort(nat.Port(fmt.Sprintf("%d/tcp", port))).WithStartupTimeout(2*time.Minute),
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

	// Set docker host to test container
	t.Setenv("DOCKER_HOST", fmt.Sprintf("tcp://localhost:%s", containerPort.Port()))
	t.Setenv("CONTRIBUTOOR_CONFIG_PATH", tmpDir)

	// Change working directory to our test directory.
	require.NoError(t, os.Chdir(tmpDir))

	// Helper function for container health check.
	checkContainerHealth := func(t *testing.T) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				// Fix data race by using a mutex for logs access
				var logsMutex sync.Mutex
				logsMutex.Lock()
				logs, err := container.Logs(context.Background())
				if err == nil {
					// Convert logs to string before logging
					logsBytes, err := io.ReadAll(logs)
					if err == nil {
						t.Logf("docker-in-docker container logs:\n%s", string(logsBytes))
					}
				}
				logsMutex.Unlock()
				t.Fatal("timeout waiting for docker-in-docker container to become healthy")
			default:
				running, err := ds.IsRunning()
				require.NoError(t, err)
				if running {
					return
				}
				time.Sleep(time.Second)
			}
		}
	}

	t.Run("lifecycle_without_metrics", func(t *testing.T) {
		// Write out compose file.
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte(composeFile), 0644))

		require.NoError(t, ds.Start())
		checkContainerHealth(t)

		require.NoError(t, ds.Stop())
		running, err := ds.IsRunning()
		require.NoError(t, err)
		require.False(t, running)
	})

	t.Run("lifecycle_with_metrics", func(t *testing.T) {
		cfgWithMetrics := &config.Config{
			Version:               "latest",
			ContributoorDirectory: tmpDir,
			RunMethod:             config.RunMethod_RUN_METHOD_DOCKER,
			MetricsAddress:        "0.0.0.0:9090",
		}

		mockSidecarConfig.EXPECT().Get().Return(cfgWithMetrics).AnyTimes()

		// Write out compose files.
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte(composeFile), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docker-compose.ports.yml"), []byte(composePortsFile), 0644))

		require.NoError(t, ds.Start())
		checkContainerHealth(t)

		require.NoError(t, ds.Stop())
		running, err := ds.IsRunning()
		require.NoError(t, err)
		require.False(t, running)
	})

	t.Run("lifecycle_with_external_container", func(t *testing.T) {
		// Write out compose file.
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte(composeFile), 0644))

		// Start a container directly with docker (not via compose) of the same name. This mimics
		// a container the installer isn't aware of.
		cmd := exec.Command("docker", "run", "-d", "--name", "contributoor", "busybox",
			"sh", "-c", "while true; do echo 'Container is running'; sleep 1; done")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "failed to start container: %s", string(output))

		// IsRunning should detect the external container.
		running, err := ds.IsRunning()
		require.NoError(t, err)
		require.True(t, running, "IsRunning should detect externally started container")

		// Stop should be able to handle the external container.
		require.NoError(t, ds.Stop())

		// Verify container is stopped.
		running, err = ds.IsRunning()
		require.NoError(t, err)
		require.False(t, running, "Container should be stopped")

		// Finally, test normal compose lifecycle works after cleaning up external container.
		require.NoError(t, ds.Start())
		checkContainerHealth(t)

		require.NoError(t, ds.Stop())

		running, err = ds.IsRunning()
		require.NoError(t, err)
		require.False(t, running)
	})
}
