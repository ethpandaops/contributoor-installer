package validate

import "testing"

func TestValidateMetricsAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "empty address",
			address: "",
			wantErr: false,
		},
		{
			name:    "just port",
			address: "9090",
			wantErr: false,
		},
		{
			name:    "colon port",
			address: ":9090",
			wantErr: false,
		},
		{
			name:    "localhost with port",
			address: "localhost:9090",
			wantErr: false,
		},
		{
			name:    "ip with port",
			address: "127.0.0.1:9090",
			wantErr: false,
		},
		{
			name:    "http url with port",
			address: "http://127.0.0.1:9090",
			wantErr: false,
		},
		{
			name:    "https url with port",
			address: "https://127.0.0.1:9090",
			wantErr: false,
		},
		{
			name:    "missing port",
			address: "127.0.0.1",
			wantErr: true,
		},
		{
			name:    "http url without port",
			address: "http://127.0.0.1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetricsAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMetricsAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
