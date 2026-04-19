package report

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenInBrowser(path string) error {
	var attempts []*exec.Cmd
	switch runtime.GOOS {
	case "windows":
		attempts = []*exec.Cmd{
			exec.Command("rundll32", "url.dll,FileProtocolHandler", path),
			exec.Command("cmd", "/c", "start", "", path),
			exec.Command("explorer", path),
		}
	case "darwin":
		attempts = []*exec.Cmd{exec.Command("open", path)}
	default:
		attempts = []*exec.Cmd{exec.Command("xdg-open", path)}
	}

	var firstErr error
	for _, c := range attempts {
		if err := c.Start(); err == nil {
			return nil
		} else if firstErr == nil {
			firstErr = err
		}
	}
	return fmt.Errorf("opening browser: %w", firstErr)
}
