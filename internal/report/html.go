package report

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adamtayara/vigil/internal/analysis"
	"github.com/adamtayara/vigil/internal/scanner"
)

const processTopN = 25

func cleanProcesses(procs []scanner.ProcessInfo) (top, rest []scanner.ProcessInfo) {
	cleaned := make([]scanner.ProcessInfo, 0, len(procs))
	for _, p := range procs {
		if strings.TrimSpace(p.Name) == "" {
			continue
		}
		if p.CPUPercent == 0 && p.MemMB == 0 && p.Exe == "" {
			continue
		}
		cleaned = append(cleaned, p)
	}
	sort.Slice(cleaned, func(i, j int) bool {
		if cleaned[i].MemMB == cleaned[j].MemMB {
			return cleaned[i].CPUPercent > cleaned[j].CPUPercent
		}
		return cleaned[i].MemMB > cleaned[j].MemMB
	})
	if len(cleaned) <= processTopN {
		return cleaned, nil
	}
	return cleaned[:processTopN], cleaned[processTopN:]
}

//go:embed tmpl/report.html
var reportTmpl string

//go:embed tmpl/style.css
var reportCSS string

//go:embed tmpl/app.js
var reportJS string

type ReportData struct {
	Result         *analysis.ScanResult
	ProcessTop     []scanner.ProcessInfo
	ProcessHidden  []scanner.ProcessInfo
	ProcessTotal   int
	NetworkRaw     []scanner.Connection
	DiskRaw        scanner.DiskInfo
	SoftRaw        []scanner.Software
	TasksRaw       []scanner.Task
	ExtRaw         []scanner.Extension
	CSS            template.CSS
	JS             template.JS
	GeneratedAt    string
}

func Generate(
	result *analysis.ScanResult,
	procs []scanner.ProcessInfo,
	conns []scanner.Connection,
	disk scanner.DiskInfo,
	soft []scanner.Software,
	tasks []scanner.Task,
	exts []scanner.Extension,
) (string, error) {
	result.Tally()

	procTop, procHidden := cleanProcesses(procs)
	data := ReportData{
		Result:        result,
		ProcessTop:    procTop,
		ProcessHidden: procHidden,
		ProcessTotal:  len(procTop) + len(procHidden),
		NetworkRaw:    conns,
		DiskRaw:       disk,
		SoftRaw:       soft,
		TasksRaw:      tasks,
		ExtRaw:        exts,
		CSS:           template.CSS(reportCSS),
		JS:            template.JS(reportJS),
		GeneratedAt:   time.Now().Format("Jan 2, 2006 at 3:04 PM"),
	}

	funcMap := template.FuncMap{
		"severityLabel": func(s analysis.Severity) string { return s.Label() },
		"severityColor": func(s analysis.Severity) string { return s.Color() },
		"severityInt":   func(s analysis.Severity) int { return int(s) },
		"formatBytes":   scanner.FormatBytes,
		"formatBytesU":  scanner.FormatBytesU,
		"formatTime": func(t time.Time) string {
			if t.IsZero() {
				return "Unknown"
			}
			return t.Format("Jan 2, 2006")
		},
		"truncate": func(s string, n int) string {
			if len(s) <= n {
				return s
			}
			return s[:n] + "..."
		},
		"add":      func(a, b int) int { return a + b },
		"percent":  func(f float64) string { return fmt.Sprintf("%.1f%%", f) },
		"contains": func(s, sub string) bool { return strings.Contains(s, sub) },
		"healthColor": func(score int) string {
			if score >= 80 {
				return "#22c55e"
			}
			if score >= 60 {
				return "#f97316"
			}
			return "#ef4444"
		},
	}

	tmpl, err := template.New("report").Funcs(funcMap).Parse(reportTmpl)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	home, _ := os.UserHomeDir()
	filename := fmt.Sprintf("vigil-report-%s.html", time.Now().Format("2006-01-02-150405"))
	outPath := filepath.Join(home, filename)

	if err := os.WriteFile(outPath, buf.Bytes(), 0600); err != nil {
		return "", fmt.Errorf("writing report: %w", err)
	}

	return outPath, nil
}
