package test

import (
	"os"
	"testing"
)

// SuppressOutput redirects stdout + stderr to /dev/null during test execution, otherwise its too noisy.
func SuppressOutput(t *testing.T) func() {
	t.Helper()

	// Save original stdout and stderr.
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Open /dev/null + redirect stdout + stderr.
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = devNull
	os.Stderr = devNull

	return func() {
		devNull.Close()

		// Restore original stdout and stderr.
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}
}
