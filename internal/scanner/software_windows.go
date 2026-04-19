//go:build windows

package scanner

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/sys/windows/registry"
)

func scanSoftwarePlatform(daysSince int) ([]Software, error) {
	var all []Software
	roots := []struct {
		hive registry.Key
		path string
	}{
		{registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`},
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`},
		{registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`},
	}

	for _, r := range roots {
		entries, err := readUninstallKey(r.hive, r.path)
		if err != nil {
			continue
		}
		all = append(all, entries...)
	}

	return filterByAge(dedup(all), daysSince), nil
}

func readUninstallKey(hive registry.Key, path string) ([]Software, error) {
	k, err := registry.OpenKey(hive, path, registry.READ)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	subkeys, err := k.ReadSubKeyNames(-1)
	if err != nil {
		return nil, err
	}

	var out []Software
	for _, sub := range subkeys {
		sk, err := registry.OpenKey(k, sub, registry.READ)
		if err != nil {
			continue
		}

		name, _, _ := sk.GetStringValue("DisplayName")
		if name == "" {
			sk.Close()
			continue
		}
		publisher, _, _ := sk.GetStringValue("Publisher")
		version, _, _ := sk.GetStringValue("DisplayVersion")
		location, _, _ := sk.GetStringValue("InstallLocation")
		dateStr, _, _ := sk.GetStringValue("InstallDate")
		sk.Close()

		sw := Software{
			Name:       name,
			Publisher:  publisher,
			Version:    version,
			InstallLoc: location,
			Source:     "registry",
		}
		if t, err := parseWindowsDate(dateStr); err == nil {
			sw.InstallDate = t
		}
		out = append(out, sw)
	}
	return out, nil
}

func parseWindowsDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if len(s) == 8 {
		return time.Parse("20060102", s)
	}
	if len(s) == 10 {
		t, err := time.Parse("2006-01-02", s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date: %s", s)
}

func dedup(all []Software) []Software {
	seen := make(map[string]bool)
	var out []Software
	for _, s := range all {
		key := strings.ToLower(s.Name)
		if !seen[key] {
			seen[key] = true
			out = append(out, s)
		}
	}
	return out
}
