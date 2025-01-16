package sidecar_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
    command: ["sh", "-c", "while true; do echo 'Container is running'; sleep 1; done"]
    healthcheck:
      test: ["CMD-SHELL", "ps aux | grep -v grep | grep 'sleep' || exit 1"]
      interval: 1s
      timeout: 1s
      retries: 3
      start_period: 1s
    networks:
      - contributoor

networks:
  contributoor:
    name: ${CONTRIBUTOOR_DOCKER_NETWORK:-contributoor}
    driver: bridge
    external: true
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
	checkContainerHealth := func(t *testing.T, ds sidecar.DockerSidecar) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var lastLogs []byte
		for {
			select {
			case <-ctx.Done():
				// Get logs only once at timeout
				logs, err := container.Logs(context.Background())
				if err == nil {
					logBytes, _ := io.ReadAll(logs)
					lastLogs = logBytes
				}
				t.Fatalf("timeout waiting for docker-in-docker container to become healthy\nLast logs: %s", string(lastLogs))
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
		checkContainerHealth(t, ds)

		// Verify the container is using the default network
		cmd := exec.Command("docker", "container", "inspect", "--format", "{{range $net,$v := .NetworkSettings.Networks}}{{printf \"%s\" $net}}{{end}}", "contributoor")
		output, err := cmd.Output()
		require.NoError(t, err)
		require.Contains(t, string(output), "contributoor", "Container should be connected to default network")

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
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docker-compose.network.yml"), []byte(composeNetworkFile), 0644))

		require.NoError(t, ds.Start())
		checkContainerHealth(t, ds)

		// Verify the container is using the default network
		cmd := exec.Command("docker", "container", "inspect", "--format", "{{range $net,$v := .NetworkSettings.Networks}}{{printf \"%s\" $net}}{{end}}", "contributoor")
		output, err := cmd.Output()
		require.NoError(t, err)
		require.Contains(t, string(output), "contributoor", "Container should be connected to default network")

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
		checkContainerHealth(t, ds)

		// Verify the container is using the default network
		cmd = exec.Command("docker", "container", "inspect", "--format", "{{range $net,$v := .NetworkSettings.Networks}}{{printf \"%s\" $net}}{{end}}", "contributoor")
		output, err = cmd.Output()
		require.NoError(t, err)
		require.Contains(t, string(output), "contributoor", "Container should be connected to default network")

		require.NoError(t, ds.Stop())

		running, err = ds.IsRunning()
		require.NoError(t, err)
		require.False(t, running)
	})

	t.Run("lifecycle_with_custom_network", func(t *testing.T) {
		// Create a custom network first
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

		// Create new mock and DockerSidecar instance for this test
		mockSidecarConfigCustom := mock.NewMockConfigManager(ctrl)
		mockSidecarConfigCustom.EXPECT().Get().Return(cfgWithNetwork).AnyTimes()
		mockSidecarConfigCustom.EXPECT().GetConfigPath().Return(filepath.Join(tmpDir, "config.yaml")).AnyTimes()

		dsCustom, err := sidecar.NewDockerSidecar(logger, mockSidecarConfigCustom, mockInstallerConfig)
		require.NoError(t, err)

		// Write out compose file
		composeFilePath := filepath.Join(tmpDir, "docker-compose.yml")
		require.NoError(t, os.WriteFile(composeFilePath, []byte(composeFile), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docker-compose.network.yml"), []byte(composeNetworkFile), 0644))

		require.NoError(t, dsCustom.Start())
		checkContainerHealth(t, dsCustom)

		// Verify the container is using the custom network
		verifyCmd := exec.Command("docker", "container", "inspect", "--format", "{{range $net,$v := .NetworkSettings.Networks}}{{printf \"%s\" $net}}{{end}}", "contributoor")
		verifyOutput, err := verifyCmd.Output()
		require.NoError(t, err)
		require.Contains(t, string(verifyOutput), customNetwork, "Container should be connected to custom network")

		require.NoError(t, dsCustom.Stop())
		running, err := dsCustom.IsRunning()
		require.NoError(t, err)
		require.False(t, running)
	})
}
