package validate

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateMetricsAddress validates the metrics address.
func ValidateMetricsAddress(address string) error {
	// Empty address is valid (disables metrics).
	if address == "" {
		return nil
	}

	// If it's just a port, prepend with colon.
	if !strings.Contains(address, ":") {
		address = ":" + address
	}

	// If it's just a port or host:port without scheme, prepend http://.
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		address = "http://" + address
	}

	u, err := url.Parse(address)
	if err != nil {
		return fmt.Errorf("invalid metrics address: %v", err)
	}

	if u.Port() == "" {
		return fmt.Errorf("metrics address must include a port")
	}

	return nil
}
