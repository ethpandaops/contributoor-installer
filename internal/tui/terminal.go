package tui

import "os"

func init() {
	// Set to xterm-256color as it's widely supported and works well with tview.
	// If whatever terminal we're using doesn't support it, it'll fallback to
	// xterm normally anyway.
	os.Setenv("TERM", "xterm-256color")
}
