//go:build darwin

package scanner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func scanSoftwarePlatform(daysSince int) ([]Software, error) {
	var all []Software

	// pkgutil --pkgs
	out, err := exec.Command("pkgutil", "--pkgs").Output()
	if err == nil {
		for _, pkg := range strings.Split(string(out), "\n") {
			pkg = strings.TrimSpace(pkg)
			if pkg == "" {
				continue
			}
			sw := Software{Name: pkg, Source: "pkgutil"}
			// Try to get install date
			info, err := exec.Command("pkgutil", "--pkg-info", pkg).Output()
			if err == nil {
				for _, line := range strings.Split(string(info), "\n") {
					if strings.HasPrefix(line, "install-time:") {
						parts := strings.SplitN(line, ":", 2)
						if len(parts) == 2 {
							// pkgutil gives Unix timestamp
							var ts int64
							if _, err := fmt.Sscan(strings.TrimSpace(parts[1]), &ts); err == nil {
								sw.InstallDate = time.Unix(ts, 0)
							}
						}
					}
				}
			}
			all = append(all, sw)
		}
	}

	// /Applications folder
	entries, err := os.ReadDir("/Applications")
	if err == nil {
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".app") {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			sw := Software{
				Name:        strings.TrimSuffix(e.Name(), ".app"),
				InstallDate: info.ModTime(),
				InstallLoc:  filepath.Join("/Applications", e.Name()),
				Source:      "applications",
			}
			all = append(all, sw)
		}
	}

	return filterByAge(all, daysSince), nil
}
