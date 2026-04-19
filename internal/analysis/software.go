package analysis

import (
	"fmt"
	"sort"
	"time"

	"github.com/adamtayara/vigil/internal/scanner"
)

func AnalyzeSoftware(software []scanner.Software) []Finding {
	var findings []Finding

	// Sort by install date descending
	sort.Slice(software, func(i, j int) bool {
		return software[i].InstallDate.After(software[j].InstallDate)
	})

	// Unknown publisher
	for _, s := range software {
		if s.Publisher == "" || s.Publisher == "N/A" {
			findings = append(findings, Finding{
				ID:       fmt.Sprintf("sw-nopub-%s", slugify(s.Name)),
				Module:   "Software",
				Severity: SeverityCheck,
				Title:    fmt.Sprintf("No verified publisher: %s", s.Name),
				Detail:   fmt.Sprintf("%s was installed with no publisher information", s.Name),
				Explain:  fmt.Sprintf("\"%s\" has no publisher information. Legitimate software from reputable companies always lists who made it. This doesn't mean it's dangerous, but it's worth verifying you intended to install it.", s.Name),
				Item:     s.Name,
				Metadata: map[string]string{"installed": s.InstallDate.Format("Jan 2, 2006"), "version": s.Version},
				Timestamp: time.Now(),
			})
		}
	}

	// Bundleware: multiple installs within 2 minutes of each other
	if clusters := findInstallClusters(software, 2*time.Minute); len(clusters) > 0 {
		for _, cluster := range clusters {
			if len(cluster) < 3 {
				continue
			}
			names := make([]string, 0, len(cluster))
			for _, s := range cluster {
				names = append(names, s.Name)
			}
			findings = append(findings, Finding{
				ID:      fmt.Sprintf("sw-bundle-%d", cluster[0].InstallDate.Unix()),
				Module:  "Software",
				Severity: SeverityCheck,
				Title:   fmt.Sprintf("%d programs installed together on %s", len(cluster), cluster[0].InstallDate.Format("Jan 2")),
				Detail:  fmt.Sprintf("Installed within 2 minutes: %v", names),
				Explain: "Multiple programs were installed at almost the same time. This sometimes happens with 'bundleware' — installers that secretly add extra software. Check if you recognize all of these programs.",
				Item:    fmt.Sprintf("%d programs", len(cluster)),
				Metadata: map[string]string{"date": cluster[0].InstallDate.Format("Jan 2, 2006"), "count": fmt.Sprintf("%d", len(cluster))},
				Timestamp: time.Now(),
			})
		}
	}

	return findings
}

func findInstallClusters(software []scanner.Software, window time.Duration) [][]scanner.Software {
	var clusters [][]scanner.Software
	var current []scanner.Software

	for _, s := range software {
		if s.InstallDate.IsZero() {
			continue
		}
		if len(current) == 0 {
			current = []scanner.Software{s}
			continue
		}
		last := current[len(current)-1]
		if s.InstallDate.After(last.InstallDate.Add(-window)) && s.InstallDate.Before(last.InstallDate.Add(window)) {
			current = append(current, s)
		} else {
			if len(current) > 1 {
				clusters = append(clusters, current)
			}
			current = []scanner.Software{s}
		}
	}
	if len(current) > 1 {
		clusters = append(clusters, current)
	}
	return clusters
}

func slugify(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s) && i < 32; i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			out = append(out, c)
		} else {
			out = append(out, '-')
		}
	}
	return string(out)
}
