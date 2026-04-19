package analysis

import "time"

type Severity int

const (
	SeverityClear    Severity = 0
	SeverityHeadsUp  Severity = 1
	SeverityCheck    Severity = 2
	SeverityWarning  Severity = 3
	SeverityCritical Severity = 4
)

func (s Severity) Label() string {
	switch s {
	case SeverityClear:
		return "All Clear"
	case SeverityHeadsUp:
		return "Heads Up"
	case SeverityCheck:
		return "Worth Checking"
	case SeverityWarning:
		return "Investigate"
	case SeverityCritical:
		return "Act Now"
	default:
		return "Unknown"
	}
}

func (s Severity) Color() string {
	switch s {
	case SeverityClear:
		return "#22c55e"
	case SeverityHeadsUp:
		return "#3b82f6"
	case SeverityCheck:
		return "#eab308"
	case SeverityWarning:
		return "#f97316"
	case SeverityCritical:
		return "#ef4444"
	default:
		return "#6b7280"
	}
}

func (s Severity) Icon() string {
	switch s {
	case SeverityClear:
		return "✓"
	case SeverityHeadsUp:
		return "ℹ"
	case SeverityCheck:
		return "⚠"
	case SeverityWarning:
		return "⚑"
	case SeverityCritical:
		return "✕"
	default:
		return "?"
	}
}

type Finding struct {
	ID         string
	Module     string
	Severity   Severity
	Title      string
	Detail     string
	Explain    string
	Item       string
	Metadata   map[string]string
	Timestamp  time.Time
}

type ScanResult struct {
	Findings   []Finding
	ScannedAt  time.Time
	Duration   time.Duration
	Hostname   string
	OS         string
	Counts     SeverityCounts
}

type SeverityCounts struct {
	Clear    int
	HeadsUp  int
	Check    int
	Warning  int
	Critical int
	Total    int
}

func (r *ScanResult) AddFindings(findings []Finding) {
	r.Findings = append(r.Findings, findings...)
}

func (r *ScanResult) Tally() {
	r.Counts = SeverityCounts{}
	for _, f := range r.Findings {
		r.Counts.Total++
		switch f.Severity {
		case SeverityClear:
			r.Counts.Clear++
		case SeverityHeadsUp:
			r.Counts.HeadsUp++
		case SeverityCheck:
			r.Counts.Check++
		case SeverityWarning:
			r.Counts.Warning++
		case SeverityCritical:
			r.Counts.Critical++
		}
	}
}

func (r *ScanResult) HealthScore() int {
	if r.Counts.Total == 0 {
		return 100
	}
	// HeadsUp is informational — no penalty. Check = minor, Warning = moderate, Critical = severe.
	penalty := r.Counts.Critical*30 + r.Counts.Warning*12 + r.Counts.Check*4
	if penalty > 100 {
		penalty = 100
	}
	return 100 - penalty
}

func (r *ScanResult) FindingsByModule(module string) []Finding {
	var out []Finding
	for _, f := range r.Findings {
		if f.Module == module {
			out = append(out, f)
		}
	}
	return out
}

func (r *ScanResult) WorstSeverity() Severity {
	worst := SeverityClear
	for _, f := range r.Findings {
		if f.Severity > worst {
			worst = f.Severity
		}
	}
	return worst
}
