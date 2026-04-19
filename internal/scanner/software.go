package scanner

import "time"

// ScanSoftware returns recently installed software.
// Platform-specific implementations in software_windows.go, software_darwin.go, software_linux.go
func ScanSoftware(daysSince int) ([]Software, error) {
	return scanSoftwarePlatform(daysSince)
}

func filterByAge(all []Software, daysSince int) []Software {
	if daysSince <= 0 {
		return all
	}
	cutoff := time.Now().AddDate(0, 0, -daysSince)
	var out []Software
	for _, s := range all {
		if s.InstallDate.IsZero() || s.InstallDate.After(cutoff) {
			out = append(out, s)
		}
	}
	return out
}
