package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBeaconService_GetNodeIdentity(t *testing.T) {
	tests := []struct {
		name           string
		responseCode   int
		responseBody   any
		expectedError  string
		expectedPeerID string
	}{
		{
			name:         "successful response",
			responseCode: http.StatusOK,
			responseBody: map[string]any{
				"data": map[string]any{
					"peer_id": "16Uiu2HAm8maLMjag1TAUM52zPfmLbm8WFzISyrG8Pv1M6pdfNZgj",
					"enr":     "enr:-test",
					"p2p_addresses": []string{
						"/ip4/127.0.0.1/tcp/9000",
					},
				},
			},
			expectedPeerID: "16Uiu2HAm8maLMjag1TAUM52zPfmLbm8WFzISyrG8Pv1M6pdfNZgj",
		},
		{
			name:          "server error",
			responseCode:  http.StatusInternalServerError,
			responseBody:  nil,
			expectedError: "unexpected status code: 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/eth/v1/node/identity", r.URL.Path)
				w.WriteHeader(tt.responseCode)

				if tt.responseBody != nil {
					err := json.NewEncoder(w).Encode(tt.responseBody)
					require.NoError(t, err)
				}
			}))
			defer server.Close()

			svc := NewBeaconService(logrus.New(), server.URL)
			identity, err := svc.GetNodeIdentity(context.Background())

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedPeerID, identity.PeerID)
		})
	}
}

func TestBeaconService_GetSyncStatus(t *testing.T) {
	tests := []struct {
		name              string
		responseCode      int
		responseBody      any
		expectedError     string
		expectedIsSyncing bool
		expectedHeadSlot  string
	}{
		{
			name:         "synced node",
			responseCode: http.StatusOK,
			responseBody: map[string]any{
				"data": map[string]any{
					"head_slot":     "1234567",
					"sync_distance": "0",
					"is_syncing":    false,
					"is_optimistic": false,
					"el_offline":    false,
				},
			},
			expectedIsSyncing: false,
			expectedHeadSlot:  "1234567",
		},
		{
			name:         "syncing node",
			responseCode: http.StatusOK,
			responseBody: map[string]any{
				"data": map[string]any{
					"head_slot":     "1000000",
					"sync_distance": "234567",
					"is_syncing":    true,
					"is_optimistic": false,
					"el_offline":    false,
				},
			},
			expectedIsSyncing: true,
			expectedHeadSlot:  "1000000",
		},
		{
			name:          "server error",
			responseCode:  http.StatusInternalServerError,
			responseBody:  nil,
			expectedError: "unexpected status code: 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/eth/v1/node/syncing", r.URL.Path)
				w.WriteHeader(tt.responseCode)

				if tt.responseBody != nil {
					err := json.NewEncoder(w).Encode(tt.responseBody)
					require.NoError(t, err)
				}
			}))
			defer server.Close()

			svc := NewBeaconService(logrus.New(), server.URL)
			syncStatus, err := svc.GetSyncStatus(context.Background())

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedIsSyncing, syncStatus.IsSyncing)
			assert.Equal(t, tt.expectedHeadSlot, syncStatus.HeadSlot)
		})
	}
}

func TestBeaconService_GetHealth(t *testing.T) {
	tests := []struct {
		name            string
		responseCode    int
		expectedHealthy bool
		expectedSyncing bool
	}{
		{
			name:            "healthy node",
			responseCode:    http.StatusOK,
			expectedHealthy: true,
			expectedSyncing: false,
		},
		{
			name:            "syncing node",
			responseCode:    http.StatusPartialContent,
			expectedHealthy: false,
			expectedSyncing: true,
		},
		{
			name:            "unhealthy node",
			responseCode:    http.StatusServiceUnavailable,
			expectedHealthy: false,
			expectedSyncing: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/eth/v1/node/health", r.URL.Path)
				w.WriteHeader(tt.responseCode)
			}))
			defer server.Close()

			svc := NewBeaconService(logrus.New(), server.URL)
			health, err := svc.GetHealth(context.Background())

			require.NoError(t, err)
			assert.Equal(t, tt.expectedHealthy, health.IsHealthy)
			assert.Equal(t, tt.expectedSyncing, health.IsSyncing)
			assert.Equal(t, tt.responseCode, health.StatusCode)
		})
	}
}

func TestBeaconService_GetBeaconInfo(t *testing.T) {
	t.Run("aggregates all beacon node info", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var resp any

			switch r.URL.Path {
			case "/eth/v1/node/health":
				w.WriteHeader(http.StatusOK)

				return
			case "/eth/v1/node/syncing":
				resp = map[string]any{
					"data": map[string]any{
						"head_slot":     "1234567",
						"sync_distance": "0",
						"is_syncing":    false,
						"is_optimistic": false,
						"el_offline":    false,
					},
				}
			case "/eth/v1/node/identity":
				resp = map[string]any{
					"data": map[string]any{
						"peer_id": "16Uiu2HAm8maLMjag1TAUM52zPfmLbm8WFzISyrG8Pv1M6pdfNZgj",
						"enr":     "enr:-test",
					},
				}
			case "/eth/v1/config/spec":
				resp = map[string]any{
					"data": map[string]any{
						"CONFIG_NAME": "mainnet",
					},
				}
			default:
				w.WriteHeader(http.StatusNotFound)

				return
			}

			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(resp)
			require.NoError(t, err)
		}))
		defer server.Close()

		svc := NewBeaconService(logrus.New(), server.URL)
		info := svc.GetBeaconInfo(context.Background())

		require.Nil(t, info.Error)
		assert.NotNil(t, info.Health)
		assert.True(t, info.Health.IsHealthy)
		assert.NotNil(t, info.Sync)
		assert.False(t, info.Sync.IsSyncing)
		assert.Equal(t, "1234567", info.Sync.HeadSlot)
		assert.NotNil(t, info.Identity)
		assert.Equal(t, "16Uiu2HAm8maLMjag1TAUM52zPfmLbm8WFzISyrG8Pv1M6pdfNZgj", info.Identity.PeerID)
		assert.Equal(t, "mainnet", info.Network)
	})

	t.Run("handles unreachable beacon node", func(t *testing.T) {
		svc := NewBeaconService(logrus.New(), "http://localhost:1")
		info := svc.GetBeaconInfo(context.Background())

		require.NotNil(t, info.Error)
		assert.Contains(t, info.Error.Error(), "beacon node unreachable")
	})

	t.Run("handles empty address", func(t *testing.T) {
		svc := NewBeaconService(logrus.New(), "")
		info := svc.GetBeaconInfo(context.Background())

		require.NotNil(t, info.Error)
		assert.Contains(t, info.Error.Error(), "no beacon node address configured")
	})
}
