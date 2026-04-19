package analysis

import (
	"fmt"
	"strings"
	"time"

	"github.com/adamtayara/vigil/internal/scanner"
)

var suspiciousTaskPaths = []string{
	`\temp\`, `/tmp/`, `\appdata\local\temp\`,
	`\users\public\`, `\downloads\`, `/downloads/`,
}

func AnalyzeTasks(tasks []scanner.Task) []Finding {
	var findings []Finding

	for _, t := range tasks {
		actionLower := strings.ToLower(t.Action)

		// Running from suspicious path
		for _, bad := range suspiciousTaskPaths {
			if strings.Contains(actionLower, strings.ToLower(bad)) {
				findings = append(findings, Finding{
					ID:       fmt.Sprintf("task-susppath-%s", slugify(t.Name)),
					Module:   "Scheduled Tasks",
					Severity: SeverityWarning,
					Title:    fmt.Sprintf("Scheduled task running from suspicious location"),
					Detail:   fmt.Sprintf("Task \"%s\" runs: %s", t.Name, t.Action),
					Explain:  fmt.Sprintf("The scheduled task \"%s\" runs a program from a temporary or downloads folder. Legitimate scheduled tasks run from Program Files or System folders. This is worth investigating.", t.Name),
					Item:     t.Name,
					Metadata: map[string]string{"action": t.Action, "trigger": t.Trigger},
					Timestamp: time.Now(),
				})
				break
			}
		}

		// Very frequent trigger (every minute or close)
		trigger := strings.ToLower(t.Trigger)
		if strings.Contains(trigger, "minute") || strings.Contains(trigger, "every 1") {
			findings = append(findings, Finding{
				ID:       fmt.Sprintf("task-frequent-%s", slugify(t.Name)),
				Module:   "Scheduled Tasks",
				Severity: SeverityCheck,
				Title:    fmt.Sprintf("Task runs very frequently: %s", t.Name),
				Detail:   fmt.Sprintf("Task \"%s\" trigger: %s", t.Name, t.Trigger),
				Explain:  fmt.Sprintf("\"%s\" is scheduled to run very frequently (every minute or more). Legitimate update tasks usually run daily or weekly. Malware often uses frequent triggers to maintain persistence.", t.Name),
				Item:     t.Name,
				Metadata: map[string]string{"trigger": t.Trigger},
				Timestamp: time.Now(),
			})
		}

		// Recently created task (last 30 days)
		if !t.Created.IsZero() && time.Since(t.Created) < 30*24*time.Hour {
			findings = append(findings, Finding{
				ID:       fmt.Sprintf("task-new-%s", slugify(t.Name)),
				Module:   "Scheduled Tasks",
				Severity: SeverityHeadsUp,
				Title:    fmt.Sprintf("Recently created task: %s", t.Name),
				Detail:   fmt.Sprintf("Task \"%s\" was created %s ago. Action: %s", t.Name, formatDuration(time.Since(t.Created)), t.Action),
				Explain:  fmt.Sprintf("A new scheduled task was created recently. This may be from a recent software installation — just verify you recognize \"%s\".", t.Name),
				Item:     t.Name,
				Metadata: map[string]string{"created": t.Created.Format("Jan 2, 2006"), "action": t.Action},
				Timestamp: time.Now(),
			})
		}

		// PowerShell with encoded command (obfuscation)
		if strings.Contains(actionLower, "powershell") && (strings.Contains(actionLower, "-enc") || strings.Contains(actionLower, "-encodedcommand")) {
			findings = append(findings, Finding{
				ID:       fmt.Sprintf("task-encoded-%s", slugify(t.Name)),
				Module:   "Scheduled Tasks",
				Severity: SeverityWarning,
				Title:    fmt.Sprintf("Scheduled task uses obfuscated PowerShell"),
				Detail:   fmt.Sprintf("Task \"%s\" uses -EncodedCommand: %s", t.Name, t.Action),
				Explain:  "This scheduled task runs PowerShell with an encoded (hidden) command. Legitimate software rarely uses this technique — it's a common way to hide malicious activity.",
				Item:     t.Name,
				Metadata: map[string]string{"action": t.Action},
				Timestamp: time.Now(),
			})
		}
	}

	return findings
}

func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}
