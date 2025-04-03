package validate

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"
)

// ValidateBeaconNodeAddress checks if any beacon node in the comma-separated list is accessible and healthy.
func ValidateBeaconNodeAddress(addresses string) error {
	var (
		nodes   = strings.Split(addresses, ",")
		lastErr error
	)

	for _, address := range nodes {
		address = strings.TrimSpace(address)
		if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
			lastErr = fmt.Errorf("beacon node address must start with http:// or https://")

			continue
		}
	}

	if lastErr != nil {
		return lastErr
	}

	for _, address := range nodes {
		address = strings.TrimSpace(address)

		// Skip health check if using Docker network hostname (non-localhost).
		host := strings.TrimPrefix(strings.TrimPrefix(address, "http://"), "https://")
		host = strings.Split(host, ":")[0]

		if !strings.HasPrefix(host, "127.0.0.1") && !strings.HasPrefix(host, "localhost") {
			return nil
		}

		client := &http.Client{Timeout: 5 * time.Second}

		resp, err := client.Get(fmt.Sprintf("%s/eth/v1/node/health", address))
		if err != nil {
			lastErr = fmt.Errorf("unable to connect to beacon node %s: %w", address, err)

			continue
		}

		defer resp.Body.Close()

		// Acceptable status codes are 200 and 206.
		// 200 is the status code for a healthy beacon node.
		// 206 is the status code for a beacon node that is syncing.
		acceptableStatusCodes := []int{http.StatusOK, http.StatusPartialContent}

		if !slices.Contains(acceptableStatusCodes, resp.StatusCode) {
			lastErr = fmt.Errorf("beacon node %s returned status %d", address, resp.StatusCode)

			continue
		}
	}

	if lastErr != nil {
		return lastErr
	}

	return nil
}
