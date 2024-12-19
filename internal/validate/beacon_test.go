package validate

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateBeaconNodeAddress(t *testing.T) {
	tests := []struct {
		name    string
		server  *httptest.Server
		address string
		wantErr bool
	}{
		{
			name: "valid beacon node",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/eth/v1/node/health" {
					t.Errorf("expected path /eth/v1/node/health, got %s", r.URL.Path)
				}

				w.WriteHeader(http.StatusOK)
			})),
			wantErr: false,
		},
		{
			name: "unhealthy beacon node",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			})),
			wantErr: true,
		},
		{
			name:    "invalid url scheme",
			address: "abc://localhost:5052",
			wantErr: true,
		},
		{
			name:    "missing scheme",
			address: "localhost:5052",
			wantErr: true,
		},
		{
			name:    "unreachable address",
			address: "http://localhost:1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if tt.server != nil {
					tt.server.Close()
				}
			}()

			address := tt.address
			if tt.server != nil {
				address = tt.server.URL
			}

			err := ValidateBeaconNodeAddress(address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBeaconNodeAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
