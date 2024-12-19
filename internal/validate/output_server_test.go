package validate

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestValidateOutputServerAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid http address",
			address: "http://example.com",
			wantErr: false,
		},
		{
			name:    "valid https address",
			address: "https://platform.ethpandaops.io",
			wantErr: false,
		},
		{
			name:    "empty address",
			address: "",
			wantErr: true,
			errMsg:  "server address is required for custom server",
		},
		{
			name:    "missing protocol",
			address: "example.com",
			wantErr: true,
			errMsg:  "server address must start with http:// or https://",
		},
		{
			name:    "invalid protocol",
			address: "abc://example.com",
			wantErr: true,
			errMsg:  "server address must start with http:// or https://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOutputServerAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOutputServerAddress() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateOutputServerAddress() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestValidateOutputServerCredentials(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		password      string
		isEthPandaOps bool
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "valid ethpandaops credentials",
			username:      "user",
			password:      "pass",
			isEthPandaOps: true,
			wantErr:       false,
		},
		{
			name:          "missing username for ethpandaops",
			username:      "",
			password:      "pass",
			isEthPandaOps: true,
			wantErr:       true,
			errMsg:        "username and password are required for ethPandaOps servers",
		},
		{
			name:          "missing password for ethpandaops",
			username:      "user",
			password:      "",
			isEthPandaOps: true,
			wantErr:       true,
			errMsg:        "username and password are required for ethPandaOps servers",
		},
		{
			name:          "valid custom server no auth",
			username:      "",
			password:      "",
			isEthPandaOps: false,
			wantErr:       false,
		},
		{
			name:          "valid custom server with auth",
			username:      "user",
			password:      "pass",
			isEthPandaOps: false,
			wantErr:       false,
		},
		{
			name:          "custom server missing password",
			username:      "user",
			password:      "",
			isEthPandaOps: false,
			wantErr:       true,
			errMsg:        "both username and password must be provided if using credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOutputServerCredentials(tt.username, tt.password, tt.isEthPandaOps)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOutputServerCredentials() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateOutputServerCredentials() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestIsEthPandaOpsServer(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    bool
	}{
		{
			name:    "valid ethpandaops server",
			address: "https://platform.ethpandaops.io/api",
			want:    true,
		},
		{
			name:    "custom server",
			address: "https://example.com",
			want:    false,
		},
		{
			name:    "empty address",
			address: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEthPandaOpsServer(tt.address); got != tt.want {
				t.Errorf("IsEthPandaOpsServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentialsEncodingDecoding(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "valid credentials",
			username: "testuser",
			password: "testpass",
			wantErr:  false,
		},
		{
			name:     "empty credentials",
			username: "",
			password: "",
			wantErr:  false,
		},
		{
			name:     "special characters",
			username: "test@user",
			password: "pass:word!",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeCredentials(tt.username, tt.password)

			// Empty credentials should result in empty encoded string.
			if tt.username == "" && tt.password == "" {
				if encoded != "" {
					t.Error("EncodeCredentials() should return empty string for empty credentials")
				}

				// Test decoding empty string
				u, p, err := DecodeCredentials("")
				if err != nil || u != "" || p != "" {
					t.Error("DecodeCredentials() should return empty strings and nil error for empty input")
				}

				return
			}

			if encoded == "" {
				t.Error("EncodeCredentials() returned empty string for non-empty credentials")

				return
			}

			// Test decoding
			username, password, err := DecodeCredentials(encoded)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeCredentials() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr {
				if username != tt.username {
					t.Errorf("DecodeCredentials() username = %v, want %v", username, tt.username)
				}

				if password != tt.password {
					t.Errorf("DecodeCredentials() password = %v, want %v", password, tt.password)
				}
			}
		})
	}

	// Test invalid base64
	t.Run("invalid base64", func(t *testing.T) {
		_, _, err := DecodeCredentials("invalid-base64")
		if err == nil {
			t.Error("DecodeCredentials() should fail with invalid base64")
		}
	})

	// Test invalid format
	t.Run("invalid format", func(t *testing.T) {
		_, _, err := DecodeCredentials(base64.StdEncoding.EncodeToString([]byte("invalid")))
		if err == nil || !strings.Contains(err.Error(), "invalid credentials format") {
			t.Error("DecodeCredentials() should fail with invalid format")
		}
	})
}
