package analysis

import (
	"fmt"
	"time"

	"github.com/adamtayara/vigil/internal/scanner"
)

func AnalyzeDisk(info scanner.DiskInfo) []Finding {
	var findings []Finding

	for _, d := range info.Drives {
		if d.Total == 0 {
			continue
		}

		if d.PercentUsed >= 90 {
			findings = append(findings, Finding{
				ID:       fmt.Sprintf("disk-critical-%s", d.Path),
				Module:   "Disk",
				Severity: SeverityCritical,
				Title:    fmt.Sprintf("Drive %s is nearly full (%.0f%%)", d.Path, d.PercentUsed),
				Detail:   fmt.Sprintf("%s: %s used of %s total (%s free)", d.Path, scanner.FormatBytesU(d.Used), scanner.FormatBytesU(d.Total), scanner.FormatBytesU(d.Free)),
				Explain:  fmt.Sprintf("Your drive at %s is almost completely full (%.0f%% used). Windows and applications need free space to run — this is likely causing slowdowns and could cause data loss. Free up space immediately.", d.Path, d.PercentUsed),
				Item:     d.Path,
				Metadata: map[string]string{
					"used":    scanner.FormatBytesU(d.Used),
					"total":   scanner.FormatBytesU(d.Total),
					"free":    scanner.FormatBytesU(d.Free),
					"percent": fmt.Sprintf("%.0f%%", d.PercentUsed),
				},
				Timestamp: time.Now(),
			})
		} else if d.PercentUsed >= 80 {
			findings = append(findings, Finding{
				ID:       fmt.Sprintf("disk-warn-%s", d.Path),
				Module:   "Disk",
				Severity: SeverityCheck,
				Title:    fmt.Sprintf("Drive %s is getting full (%.0f%%)", d.Path, d.PercentUsed),
				Detail:   fmt.Sprintf("%s: %s used of %s (%s free)", d.Path, scanner.FormatBytesU(d.Used), scanner.FormatBytesU(d.Total), scanner.FormatBytesU(d.Free)),
				Explain:  fmt.Sprintf("Your drive is %.0f%% full. It's still working fine, but consider freeing up space — performance can degrade when drives are over 85%% full.", d.PercentUsed),
				Item:     d.Path,
				Metadata: map[string]string{"percent": fmt.Sprintf("%.0f%%", d.PercentUsed), "free": scanner.FormatBytesU(d.Free)},
				Timestamp: time.Now(),
			})
		}
	}

	// Large temp folder
	const warnTempBytes = 1 * 1024 * 1024 * 1024 // 1 GB
	if info.TempSize > warnTempBytes {
		findings = append(findings, Finding{
			ID:      "disk-temp",
			Module:  "Disk",
			Severity: SeverityHeadsUp,
			Title:   fmt.Sprintf("Large temp folder: %s", scanner.FormatBytes(info.TempSize)),
			Detail:  fmt.Sprintf("Your temp folder (%s) contains %s of files", info.TempPath, scanner.FormatBytes(info.TempSize)),
			Explain: fmt.Sprintf("Your temporary files folder is %s in size. These files are safe to delete and freeing them could recover useful disk space. You can do this through Windows Settings → Storage → Temporary Files.", scanner.FormatBytes(info.TempSize)),
			Item:    info.TempPath,
			Metadata: map[string]string{"size": scanner.FormatBytes(info.TempSize), "path": info.TempPath},
			Timestamp: time.Now(),
		})
	}

	return findings
}
