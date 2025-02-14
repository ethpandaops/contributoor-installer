package validate

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateBeaconNodeAddress(t *testing.T) {
	tests := []struct {
		name    string
		servers []*httptest.Server
		address string
		wantErr bool
	}{
		{
			name: "single valid beacon node",
			servers: []*httptest.Server{
				httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/eth/v1/node/health" {
						t.Errorf("expected path /eth/v1/node/health, got %s", r.URL.Path)
					}
					w.WriteHeader(http.StatusOK)
				})),
			},
			wantErr: false,
		},
		{
			name: "multiple valid nodes",
			servers: []*httptest.Server{
				httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})),
				httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})),
			},
			wantErr: false,
		},
		{
			name:    "multiple nodes with spaces",
			address: "http://localhost:5053,  http://localhost:5054  ",
			servers: []*httptest.Server{
				httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})),
				httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})),
			},
			wantErr: false,
		},
		{
			name: "all nodes unhealthy but valid URLs",
			servers: []*httptest.Server{
				httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusServiceUnavailable)
				})),
				httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusServiceUnavailable)
				})),
			},
			wantErr: true,
		},
		{
			name:    "single invalid url scheme",
			address: "abc://localhost:5052",
			wantErr: true,
		},
		{
			name:    "missing scheme",
			address: "localhost:5052",
			wantErr: true,
		},
		{
			name:    "mixed valid and invalid schemes",
			address: "abc://localhost:5052,http://localhost:5053",
			wantErr: true,
		},
		{
			name:    "unreachable address",
			address: "http://localhost:1",
			wantErr: true,
		},
		{
			name:    "multiple unreachable addresses",
			address: "http://localhost:1,http://localhost:2",
			wantErr: true,
		},
		{
			name:    "empty address",
			address: "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			address: "  ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				for _, server := range tt.servers {
					if server != nil {
						server.Close()
					}
				}
			}()

			address := tt.address
			if len(tt.servers) > 0 {
				addresses := make([]string, len(tt.servers))
				for i, server := range tt.servers {
					addresses[i] = server.URL
				}
				// Add spaces for the "multiple nodes with spaces" test
				if tt.name == "multiple nodes with spaces" {
					address = strings.Join(addresses, ", ")
				} else {
					address = strings.Join(addresses, ",")
				}
			}

			err := ValidateBeaconNodeAddress(address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBeaconNodeAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
