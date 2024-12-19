package validate

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// ValidateOutputServerAddress validates the output server address.
func ValidateOutputServerAddress(address string) error {
	if address == "" {
		return fmt.Errorf("server address is required for custom server")
	}

	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		return fmt.Errorf("server address must start with http:// or https://")
	}

	return nil
}

// ValidateOutputServerCredentials validates the credentials based on server type.
func ValidateOutputServerCredentials(username, password string, isEthPandaOpsServer bool) error {
	if isEthPandaOpsServer {
		// Credentials are required for ethPandaOps servers.
		if username == "" || password == "" {
			return fmt.Errorf("username and password are required for ethPandaOps servers")
		}

		return nil
	}

	// For custom servers, either both must be provided or neither.
	if (username == "" && password != "") || (username != "" && password == "") {
		return fmt.Errorf("both username and password must be provided if using credentials")
	}

	return nil
}

// IsEthPandaOpsServer checks if the given address is an ethPandaOps server.
func IsEthPandaOpsServer(address string) bool {
	return strings.Contains(address, "platform.ethpandaops.io")
}

// EncodeCredentials creates a base64 encoded string of username:password.
func EncodeCredentials(username, password string) string {
	return base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", username, password)),
	)
}

// DecodeCredentials decodes base64 encoded credentials.
func DecodeCredentials(encoded string) (username, password string, err error) {
	if encoded == "" {
		return "", "", nil
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", err
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid credentials format")
	}

	return parts[0], parts[1], nil
}
