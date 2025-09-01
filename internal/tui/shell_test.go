package tui

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestUpgradeWarning(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		wantOutput     bool
	}{
		{
			name:           "no warning when current is latest tag",
			currentVersion: "latest",
			latestVersion:  "v1.0.0",
			wantOutput:     false,
		},
		{
			name:           "no warning when versions match",
			currentVersion: "v1.0.0",
			latestVersion:  "v1.0.0",
			wantOutput:     false,
		},
		{
			name:           "warning when versions differ",
			currentVersion: "v1.0.0",
			latestVersion:  "v1.1.0",
			wantOutput:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout.
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			UpgradeWarning(tt.currentVersion, tt.latestVersion)

			// Restore stdout.
			w.Close()

			os.Stdout = old

			var buf bytes.Buffer

			_, _ = io.Copy(&buf, r)

			output := buf.String()

			if tt.wantOutput {
				if !strings.Contains(output, "You are running an old version") {
					t.Errorf("expected warning message, got none")
				}

				if !strings.Contains(output, tt.latestVersion) {
					t.Errorf("expected version %s in output, not found", tt.latestVersion)
				}
			} else {
				if output != "" {
					t.Errorf("expected no output, got: %s", output)
				}
			}
		})
	}
}
