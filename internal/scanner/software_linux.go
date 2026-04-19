//go:build linux

package scanner

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
	"time"
)

func scanSoftwarePlatform(daysSince int) ([]Software, error) {
	var all []Software

	// Try dpkg log
	if entries := parseDpkgLog(); len(entries) > 0 {
		all = append(all, entries...)
	}

	// Try rpm
	if entries := parseRPM(); len(entries) > 0 {
		all = append(all, entries...)
	}

	// Try pacman log
	if entries := parsePacmanLog(); len(entries) > 0 {
		all = append(all, entries...)
	}

	return filterByAge(all, daysSince), nil
}

func parseDpkgLog() []Software {
	f, err := os.Open("/var/log/dpkg.log")
	if err != nil {
		return nil
	}
	defer f.Close()

	var out []Software
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, " install ") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}
		dateStr := parts[0] + " " + parts[1]
		t, err := time.Parse("2006-01-02 15:04:05", dateStr)
		if err != nil {
			continue
		}
		name := parts[3]
		if idx := strings.Index(name, ":"); idx > 0 {
			name = name[:idx]
		}
		out = append(out, Software{Name: name, InstallDate: t, Source: "dpkg"})
	}
	return out
}

func parseRPM() []Software {
	out, err := exec.Command("rpm", "-qa", "--queryformat", "%{NAME}|%{VENDOR}|%{VERSION}|%{INSTALLTIME:date}\n").Output()
	if err != nil {
		return nil
	}
	var res []Software
	for _, line := range strings.Split(string(out), "\n") {
		parts := strings.SplitN(line, "|", 4)
		if len(parts) < 4 {
			continue
		}
		t, _ := time.Parse("Mon Jan _2 15:04:05 2006", strings.TrimSpace(parts[3]))
		res = append(res, Software{
			Name:        parts[0],
			Publisher:   parts[1],
			Version:     parts[2],
			InstallDate: t,
			Source:      "rpm",
		})
	}
	return res
}

func parsePacmanLog() []Software {
	f, err := os.Open("/var/log/pacman.log")
	if err != nil {
		return nil
	}
	defer f.Close()

	var out []Software
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "[ALPM] installed") {
			continue
		}
		// [2024-01-15T12:00:00+0000] [ALPM] installed packagename (version)
		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}
		dateStr := strings.Trim(parts[0], "[]")
		t, _ := time.Parse("2006-01-02T15:04:05-0700", dateStr)
		name := parts[3]
		out = append(out, Software{Name: name, InstallDate: t, Source: "pacman"})
	}
	return out
}
