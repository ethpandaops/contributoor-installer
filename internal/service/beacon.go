package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// BeaconService provides methods to interact with a beacon node's REST API.
type BeaconService interface {
	// GetNodeIdentity returns the identity of the beacon node.
	GetNodeIdentity(ctx context.Context) (*NodeIdentity, error)
	// GetSyncStatus returns the sync status of the beacon node.
	GetSyncStatus(ctx context.Context) (*SyncStatus, error)
	// GetHealth returns the health status of the beacon node.
	GetHealth(ctx context.Context) (*HealthStatus, error)
	// GetBeaconInfo fetches all beacon node info in one call.
	// Errors are stored in BeaconInfo.Error rather than returned.
	GetBeaconInfo(ctx context.Context) *BeaconInfo
}

// NodeIdentity represents the response from /eth/v1/node/identity.
//
//nolint:tagliatelle // Ethereum Beacon API uses snake_case.
type NodeIdentity struct {
	PeerID             string   `json:"peer_id"`
	ENR                string   `json:"enr"`
	P2PAddresses       []string `json:"p2p_addresses"`
	DiscoveryAddresses []string `json:"discovery_addresses"`
	Metadata           struct {
		SeqNumber string `json:"seq_number"`
		Attnets   string `json:"attnets"`
		Syncnets  string `json:"syncnets"`
	} `json:"metadata"`
}

// SyncStatus represents the response from /eth/v1/node/syncing.
//
//nolint:tagliatelle // Ethereum Beacon API uses snake_case.
type SyncStatus struct {
	HeadSlot     string `json:"head_slot"`
	SyncDistance string `json:"sync_distance"`
	IsSyncing    bool   `json:"is_syncing"`
	IsOptimistic bool   `json:"is_optimistic"`
	ELOffline    bool   `json:"el_offline"`
}

// HealthStatus represents the parsed health response.
type HealthStatus struct {
	StatusCode int
	IsHealthy  bool
	IsSyncing  bool
}

// BeaconInfo aggregates all beacon node information.
type BeaconInfo struct {
	Identity *NodeIdentity
	Sync     *SyncStatus
	Health   *HealthStatus
	Network  string
	Error    error
}

// beaconService implements BeaconService.
type beaconService struct {
	log     *logrus.Logger
	client  *http.Client
	address string
}

// NewBeaconService creates a new BeaconService instance.
func NewBeaconService(log *logrus.Logger, address string) BeaconService {
	return &beaconService{
		log:     log,
		client:  &http.Client{Timeout: 5 * time.Second},
		address: strings.TrimSuffix(address, "/"),
	}
}

// GetNodeIdentity fetches the node identity from the beacon node.
func (s *beaconService) GetNodeIdentity(ctx context.Context) (*NodeIdentity, error) {
	var response struct {
		Data NodeIdentity `json:"data"`
	}

	if err := s.doGet(ctx, "/eth/v1/node/identity", &response); err != nil {
		return nil, fmt.Errorf("failed to get node identity: %w", err)
	}

	return &response.Data, nil
}

// GetSyncStatus fetches the sync status from the beacon node.
func (s *beaconService) GetSyncStatus(ctx context.Context) (*SyncStatus, error) {
	var response struct {
		Data SyncStatus `json:"data"`
	}

	if err := s.doGet(ctx, "/eth/v1/node/syncing", &response); err != nil {
		return nil, fmt.Errorf("failed to get sync status: %w", err)
	}

	return &response.Data, nil
}

// GetHealth fetches the health status from the beacon node.
func (s *beaconService) GetHealth(ctx context.Context) (*HealthStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.address+"/eth/v1/node/health", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req) //nolint:gosec // Address is from user config, not untrusted input.
	if err != nil {
		return nil, fmt.Errorf("failed to get health: %w", err)
	}

	defer resp.Body.Close()

	health := &HealthStatus{
		StatusCode: resp.StatusCode,
		IsHealthy:  resp.StatusCode == http.StatusOK,
		IsSyncing:  resp.StatusCode == http.StatusPartialContent,
	}

	return health, nil
}

// GetBeaconInfo fetches all beacon node info and returns aggregated results.
func (s *beaconService) GetBeaconInfo(ctx context.Context) *BeaconInfo {
	info := &BeaconInfo{}

	// If no address configured, return early.
	if s.address == "" {
		info.Error = fmt.Errorf("no beacon node address configured")

		return info
	}

	// Fetch health first as it's the quickest indicator of connectivity.
	health, err := s.GetHealth(ctx)
	if err != nil {
		info.Error = fmt.Errorf("beacon node unreachable: %w", err)

		return info
	}

	info.Health = health

	// Fetch sync status.
	sync, err := s.GetSyncStatus(ctx)
	if err != nil {
		s.log.WithError(err).Debug("Failed to get sync status")
	} else {
		info.Sync = sync
	}

	// Fetch identity.
	identity, err := s.GetNodeIdentity(ctx)
	if err != nil {
		s.log.WithError(err).Debug("Failed to get node identity")
	} else {
		info.Identity = identity
	}

	// Try to determine network from spec.
	network, err := s.getNetwork(ctx)
	if err != nil {
		s.log.WithError(err).Debug("Failed to get network")
	} else {
		info.Network = network
	}

	return info
}

// getNetwork fetches the network name from the beacon node spec.
func (s *beaconService) getNetwork(ctx context.Context) (string, error) {
	var response specResponse

	if err := s.doGet(ctx, "/eth/v1/config/spec", &response); err != nil {
		return "", fmt.Errorf("failed to get spec: %w", err)
	}

	return response.Data.ConfigName, nil
}

// specResponse represents the response from /eth/v1/config/spec.
//
//nolint:tagliatelle // Ethereum Beacon API uses SCREAMING_SNAKE_CASE for config.
type specResponse struct {
	Data struct {
		ConfigName string `json:"CONFIG_NAME"`
	} `json:"data"`
}

// doGet performs a GET request and decodes the JSON response.
func (s *beaconService) doGet(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.address+path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req) //nolint:gosec // Address is from user config, not untrusted input.
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
