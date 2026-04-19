package analysis

import (
	"fmt"
	"strings"
	"time"

	"github.com/adamtayara/vigil/internal/scanner"
)

var broadPermissions = map[string]string{
	"<all_urls>":        "read your data on every website",
	"tabs":              "read your browser tab URLs and titles",
	"history":           "read your full browsing history",
	"webRequest":        "intercept and modify network requests",
	"webRequestBlocking": "block or modify web traffic",
	"cookies":           "read and modify all cookies (including login sessions)",
	"nativeMessaging":   "communicate with apps on your computer",
	"debugger":          "debug and inspect any web page",
	"pageCapture":       "capture full-page screenshots",
}

func AnalyzeExtensions(exts []scanner.Extension) []Finding {
	var findings []Finding

	for _, e := range exts {
		// Sideloaded extensions
		if strings.Contains(strings.ToLower(e.Source), "sideload") || strings.Contains(strings.ToLower(e.Source), "third-party") {
			findings = append(findings, Finding{
				ID:       fmt.Sprintf("ext-sideload-%s-%s", e.Browser, e.ID),
				Module:   "Browser Extensions",
				Severity: SeverityWarning,
				Title:    fmt.Sprintf("Unreviewed extension in %s: %s", e.Browser, e.Name),
				Detail:   fmt.Sprintf("%s (%s) is not from the official store. Source: %s", e.Name, e.Browser, e.Source),
				Explain:  fmt.Sprintf("\"%s\" in your %s browser was not installed from the official extension store. Extensions installed outside the store bypass Google's/Mozilla's safety review and could be malicious.", e.Name, e.Browser),
				Item:     fmt.Sprintf("%s (%s)", e.Name, e.Browser),
				Metadata: map[string]string{"browser": e.Browser, "source": e.Source, "id": e.ID},
				Timestamp: time.Now(),
			})
		}

		// Broad permissions
		broadFound := []string{}
		for _, perm := range e.Permissions {
			if desc, ok := broadPermissions[perm]; ok {
				broadFound = append(broadFound, desc)
			}
			// host permissions
			if perm == "<all_urls>" || strings.HasPrefix(perm, "http://*/") || strings.HasPrefix(perm, "https://*/") {
				if !contains(broadFound, "read your data on every website") {
					broadFound = append(broadFound, "read your data on every website")
				}
			}
		}

		if len(broadFound) >= 2 {
			findings = append(findings, Finding{
				ID:       fmt.Sprintf("ext-perms-%s-%s", e.Browser, e.ID),
				Module:   "Browser Extensions",
				Severity: SeverityCheck,
				Title:    fmt.Sprintf("Extension with broad permissions: %s", e.Name),
				Detail:   fmt.Sprintf("%s (%s) can: %s", e.Name, e.Browser, strings.Join(broadFound, "; ")),
				Explain:  fmt.Sprintf("\"%s\" has permissions that give it significant access to your browser activity. This doesn't mean it's malicious — but make sure you trust and actually use this extension.", e.Name),
				Item:     fmt.Sprintf("%s (%s)", e.Name, e.Browser),
				Metadata: map[string]string{"browser": e.Browser, "permissions": strings.Join(broadFound, ", ")},
				Timestamp: time.Now(),
			})
		}

		// Very old extension (>3 years since update)
		if !e.UpdatedAt.IsZero() && time.Since(e.UpdatedAt) > 3*365*24*time.Hour {
			findings = append(findings, Finding{
				ID:       fmt.Sprintf("ext-old-%s-%s", e.Browser, e.ID),
				Module:   "Browser Extensions",
				Severity: SeverityHeadsUp,
				Title:    fmt.Sprintf("Outdated extension: %s", e.Name),
				Detail:   fmt.Sprintf("%s (%s) hasn't been updated since %s", e.Name, e.Browser, e.UpdatedAt.Format("Jan 2006")),
				Explain:  fmt.Sprintf("\"%s\" hasn't received an update in over 3 years. Abandoned extensions may have unpatched security vulnerabilities. Consider removing it if you don't actively use it.", e.Name),
				Item:     fmt.Sprintf("%s (%s)", e.Name, e.Browser),
				Metadata: map[string]string{"last_updated": e.UpdatedAt.Format("Jan 2, 2006")},
				Timestamp: time.Now(),
			})
		}
	}

	return findings
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
