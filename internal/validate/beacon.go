package validate

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ValidateBeaconNodeAddress checks if a beacon node is accessible and healthy.
func ValidateBeaconNodeAddress(address string) error {
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		return fmt.Errorf("beacon node address must start with http:// or https://")
	}

	// Skip health check if using Docker network hostname (non-localhost).
	host := strings.TrimPrefix(strings.TrimPrefix(address, "http://"), "https://")
	host = strings.Split(host, ":")[0]

	if !strings.HasPrefix(host, "127.0.0.1") && !strings.HasPrefix(host, "localhost") {
		return nil
	}

	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(fmt.Sprintf("%s/eth/v1/node/health", address))
	if err != nil {
		return fmt.Errorf("we're unable to connect to your beacon node: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("beacon node returned status %d", resp.StatusCode)
	}

	return nil
}
