package sidecar_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/ethpandaops/contributoor-installer/internal/installer"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar"
	"github.com/ethpandaops/contributoor-installer/internal/sidecar/mock"
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
		cfg    = &sidecar.Config{
			Version:               "latest",
			ContributoorDirectory: tmpDir,
			RunMethod:             sidecar.RunMethodDocker,
		}
	)

	// Create mock config manager
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInstallerConfig := installer.NewConfig()
	mockSidecarConfig := mock.NewMockConfigManager(ctrl)
	mockSidecarConfig.EXPECT().Get().Return(cfg).AnyTimes()
	mockSidecarConfig.EXPECT().GetConfigDir().Return(tmpDir).AnyTimes()
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

	// Create docker service with mock config
	ds, err := sidecar.NewDockerSidecar(logger, mockSidecarConfig, mockInstallerConfig)
	require.NoError(t, err)

	// Set docker host to test container.
	t.Setenv("DOCKER_HOST", fmt.Sprintf("tcp://localhost:%s", containerPort.Port()))

	// Write out dummy compose file.
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte(composeFile), 0644))

	// Run our tests in a real container.
	t.Run("lifecycle", func(t *testing.T) {
		// Ensure Start() executes as expected.
		require.NoError(t, ds.Start())

		// Wait for the container to be healthy.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				// If we timeout, log the container logs so we get some idea of what went wrong.
				logs, err := container.Logs(context.Background())
				if err == nil {
					t.Logf("docker-in-docker container logs:\n%s", logs)
				}

				t.Fatal("timeout waiting for docker-in-docker container to become healthy")
			default:
				// Check if the container is running.
				running, err := ds.IsRunning()
				require.NoError(t, err)

				if running {
					goto containerRunning
				}

				time.Sleep(time.Second)
			}
		}

	containerRunning:
		// Stop container and verify it's not running anymore.
		require.NoError(t, ds.Stop())

		running, err := ds.IsRunning()
		require.NoError(t, err)
		require.False(t, running)
	})
}
