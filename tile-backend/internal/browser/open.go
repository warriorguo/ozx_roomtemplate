// Package browser provides a best-effort, cross-platform launcher for the
// user's default web browser. Failure to open a browser is never fatal — the
// caller should still log the URL prominently so the user can paste it
// manually.
package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Open attempts to launch the user's default browser pointed at url and
// returns immediately. Errors surface as a returned value rather than panics
// so callers can log them.
func Open(url string) error {
	cmd, args := command(url)
	if cmd == "" {
		return fmt.Errorf("browser: unsupported platform %s", runtime.GOOS)
	}
	c := exec.Command(cmd, args...)
	return c.Start()
}

func command(url string) (string, []string) {
	switch runtime.GOOS {
	case "darwin":
		return "open", []string{url}
	case "linux":
		return "xdg-open", []string{url}
	case "windows":
		// rundll32 is universally present and handles HTTP URLs via the
		// FileProtocolHandler; no shell quoting concerns.
		return "rundll32", []string{"url.dll,FileProtocolHandler", url}
	default:
		return "", nil
	}
}
