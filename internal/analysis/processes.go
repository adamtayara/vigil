package analysis

import (
	"fmt"
	"math"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/adamtayara/vigil/internal/scanner"
)

var suspiciousParentChild = map[string][]string{
	"winword.exe":   {"cmd.exe", "powershell.exe", "wscript.exe", "cscript.exe", "mshta.exe"},
	"excel.exe":     {"cmd.exe", "powershell.exe", "wscript.exe", "cscript.exe", "mshta.exe"},
	"outlook.exe":   {"cmd.exe", "powershell.exe", "wscript.exe", "cscript.exe"},
	"chrome.exe":    {"cmd.exe", "powershell.exe"},
	"firefox.exe":   {"cmd.exe", "powershell.exe"},
	"msedge.exe":    {"cmd.exe", "powershell.exe"},
	"mshta.exe":     {"cmd.exe", "powershell.exe", "wscript.exe"},
	"wscript.exe":   {"cmd.exe", "powershell.exe"},
	"cscript.exe":   {"cmd.exe", "powershell.exe"},
}

var suspiciousPaths = []string{
	`\temp\`, `/tmp/`, `\appdata\local\temp\`, `\users\public\`,
	`\downloads\`, `/downloads/`,
}

var systemProcessPaths = map[string][]string{
	"svchost.exe":  {`\system32\`, `\syswow64\`},
	"lsass.exe":    {`\system32\`},
	"csrss.exe":    {`\system32\`},
	"winlogon.exe": {`\system32\`},
	"services.exe": {`\system32\`},
	"smss.exe":     {`\system32\`},
}

func AnalyzeProcesses(procs []scanner.ProcessInfo) []Finding {
	var findings []Finding

	pidMap := make(map[int32]scanner.ProcessInfo, len(procs))
	for _, p := range procs {
		pidMap[p.PID] = p
	}

	for _, p := range procs {
		exeLower := strings.ToLower(filepath.Base(p.Exe))
		exePathLower := strings.ToLower(p.Exe)

		// Suspicious execution path
		for _, bad := range suspiciousPaths {
			if strings.Contains(exePathLower, strings.ToLower(bad)) {
				findings = append(findings, Finding{
					ID:       fmt.Sprintf("proc-susppath-%d", p.PID),
					Module:   "Processes",
					Severity: SeverityWarning,
					Title:    fmt.Sprintf("Process running from suspicious location"),
					Detail:   fmt.Sprintf("%s (PID %d) is running from: %s", p.Name, p.PID, p.Exe),
					Explain:  fmt.Sprintf("Legitimate programs usually run from Program Files or system directories. \"%s\" is running from a temporary or downloads folder — malware often hides here.", p.Name),
					Item:     p.Name,
					Metadata: map[string]string{"pid": fmt.Sprintf("%d", p.PID), "path": p.Exe},
					Timestamp: time.Now(),
				})
				break
			}
		}

		// System process running from wrong location (Windows)
		if runtime.GOOS == "windows" {
			if validPaths, ok := systemProcessPaths[exeLower]; ok {
				if p.Exe != "" {
					validLocation := false
					for _, vp := range validPaths {
						if strings.Contains(exePathLower, vp) {
							validLocation = true
							break
						}
					}
					if !validLocation {
						findings = append(findings, Finding{
							ID:       fmt.Sprintf("proc-fakesys-%d", p.PID),
							Module:   "Processes",
							Severity: SeverityCritical,
							Title:    "System process running from unusual location",
							Detail:   fmt.Sprintf("%s (PID %d) is at: %s — expected in System32", p.Name, p.PID, p.Exe),
							Explain:  fmt.Sprintf("Windows system files like \"%s\" should only run from C:\\Windows\\System32. This copy is running from a different location, which is a strong sign of malware masquerading as a system process.", p.Name),
							Item:     p.Name,
							Metadata: map[string]string{"pid": fmt.Sprintf("%d", p.PID), "path": p.Exe},
							Timestamp: time.Now(),
						})
					}
				}
			}
		}

		// Suspicious parent-child
		if parent, ok := pidMap[p.PPID]; ok {
			parentLower := strings.ToLower(filepath.Base(parent.Exe))
			childLower := strings.ToLower(filepath.Base(p.Exe))
			if badChildren, ok := suspiciousParentChild[parentLower]; ok {
				for _, bad := range badChildren {
					if childLower == bad {
						findings = append(findings, Finding{
							ID:       fmt.Sprintf("proc-parentchild-%d", p.PID),
							Module:   "Processes",
							Severity: SeverityWarning,
							Title:    "Suspicious program launch chain",
							Detail:   fmt.Sprintf("%s launched %s (PID %d)", parent.Name, p.Name, p.PID),
							Explain:  fmt.Sprintf("Your %s application launched a command-line tool (%s). Office apps and browsers typically shouldn't do this — it's a common sign of a malicious document or extension.", parent.Name, p.Name),
							Item:     fmt.Sprintf("%s → %s", parent.Name, p.Name),
							Metadata: map[string]string{"parent": parent.Name, "child": p.Name, "pid": fmt.Sprintf("%d", p.PID)},
							Timestamp: time.Now(),
						})
						break
					}
				}
			}
		}

		// High entropy process name (random-looking)
		if nameEntropy(p.Name) > 4.5 && len(p.Name) > 8 {
			findings = append(findings, Finding{
				ID:       fmt.Sprintf("proc-entropy-%d", p.PID),
				Module:   "Processes",
				Severity: SeverityCheck,
				Title:    "Process with randomized-looking name",
				Detail:   fmt.Sprintf("%s (PID %d) has a name that appears randomly generated", p.Name, p.PID),
				Explain:  "Some malware uses randomly generated file names to avoid detection. This process name has an unusual character pattern worth checking.",
				Item:     p.Name,
				Metadata: map[string]string{"pid": fmt.Sprintf("%d", p.PID)},
				Timestamp: time.Now(),
			})
		}
	}

	// Top CPU consumers
	topCPU := topByField(procs, func(p scanner.ProcessInfo) float64 { return p.CPUPercent }, 3)
	for _, p := range topCPU {
		if p.CPUPercent < 15 {
			continue
		}
		findings = append(findings, Finding{
			ID:       fmt.Sprintf("proc-highcpu-%d", p.PID),
			Module:   "Processes",
			Severity: SeverityHeadsUp,
			Title:    fmt.Sprintf("High CPU usage: %s", p.Name),
			Detail:   fmt.Sprintf("%s is using %.1f%% CPU", p.Name, p.CPUPercent),
			Explain:  fmt.Sprintf("\"%s\" is using a significant portion of your CPU. This may be normal (video, games, updates) or could indicate a problem if you didn't expect it.", p.Name),
			Item:     p.Name,
			Metadata: map[string]string{"cpu": fmt.Sprintf("%.1f%%", p.CPUPercent), "pid": fmt.Sprintf("%d", p.PID)},
			Timestamp: time.Now(),
		})
	}

	// Top RAM consumers
	topRAM := topByField(procs, func(p scanner.ProcessInfo) float64 { return p.MemMB }, 3)
	for _, p := range topRAM {
		if p.MemMB < 500 {
			continue
		}
		findings = append(findings, Finding{
			ID:       fmt.Sprintf("proc-highram-%d", p.PID),
			Module:   "Processes",
			Severity: SeverityHeadsUp,
			Title:    fmt.Sprintf("High memory usage: %s", p.Name),
			Detail:   fmt.Sprintf("%s is using %.0f MB of RAM", p.Name, p.MemMB),
			Explain:  fmt.Sprintf("\"%s\" is using a lot of memory. This can cause slowdowns if your system is running low. Consider closing it if you're not actively using it.", p.Name),
			Item:     p.Name,
			Metadata: map[string]string{"ram": fmt.Sprintf("%.0f MB", p.MemMB), "pid": fmt.Sprintf("%d", p.PID)},
			Timestamp: time.Now(),
		})
	}

	return findings
}

func nameEntropy(s string) float64 {
	s = strings.TrimSuffix(strings.ToLower(s), filepath.Ext(s))
	if len(s) == 0 {
		return 0
	}
	freq := make(map[rune]float64)
	for _, c := range s {
		freq[c]++
	}
	var entropy float64
	l := float64(len(s))
	for _, count := range freq {
		p := count / l
		entropy -= p * math.Log2(p)
	}
	return entropy
}

func topByField(procs []scanner.ProcessInfo, fn func(scanner.ProcessInfo) float64, n int) []scanner.ProcessInfo {
	sorted := make([]scanner.ProcessInfo, len(procs))
	copy(sorted, procs)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if fn(sorted[j]) > fn(sorted[i]) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}
