package analysis

import (
	"fmt"
	"strings"
	"time"

	"github.com/adamtayara/vigil/internal/scanner"
)

var trustedProcesses = map[string]bool{
	// Browsers
	"chrome.exe": true, "firefox.exe": true, "msedge.exe": true, "brave.exe": true,
	"safari": true, "opera.exe": true, "chromium.exe": true,
	// Communication & productivity
	"spotify.exe": true, "discord.exe": true, "slack.exe": true, "zoom.exe": true,
	"teams.exe": true, "msteams.exe": true, "skype.exe": true,
	// Cloud sync
	"onedrive.exe": true, "dropbox.exe": true, "googledrivefs.exe": true, "box.exe": true,
	// Windows system
	"svchost.exe": true, "lsass.exe": true, "wininit.exe": true, "winlogon.exe": true,
	"services.exe": true, "csrss.exe": true, "smss.exe": true, "system": true, "idle": true,
	"spoolsv.exe": true, "searchindexer.exe": true, "taskhostw.exe": true,
	"runtimebroker.exe": true, "sihost.exe": true, "ctfmon.exe": true,
	// Dev tools
	"node.exe": true, "git.exe": true, "python.exe": true, "java.exe": true,
	"code.exe": true, "windowsterminal.exe": true, "powershell.exe": true,
	// Security
	"msmpeng.exe": true, "nissrv.exe": true, "securityhealthservice.exe": true,
	// Other common
	"steam.exe": true, "epicgameslauncher.exe": true, "gamebar.exe": true,
	"antimalware service executable": true,
}

var suspiciousPorts = map[uint32]string{
	4444: "common reverse shell port",
	1337: "common hacking tool port",
	31337: "common backdoor port",
	6667: "IRC (often used by botnets)",
	6697: "IRC over TLS",
	9001: "Tor default port",
	9050: "Tor SOCKS proxy",
}

func AnalyzeNetwork(conns []scanner.Connection) []Finding {
	var findings []Finding

	processConns := make(map[string][]scanner.Connection)
	for _, c := range conns {
		processConns[c.ProcessName] = append(processConns[c.ProcessName], c)
	}

	for proc, pconns := range processConns {
		procLower := strings.ToLower(proc)
		var external []scanner.Connection
		for _, c := range pconns {
			if scanner.IsExternalIP(c.RemoteAddr) && c.Status == "ESTABLISHED" {
				external = append(external, c)
			}
		}

		if len(external) == 0 {
			continue
		}

		// Non-browser/trusted process with external connections
		if !trustedProcesses[procLower] && len(external) > 0 {
			uniqueIPs := uniqueRemoteIPs(external)
			for _, c := range external[:min(1, len(external))] {
				// Skip reverse DNS lookup during scan for speed; show raw IP
				rdns := ""
				severity := SeverityHeadsUp
				explain := fmt.Sprintf("\"%s\" is communicating with an external server at %s.", proc, c.RemoteAddr)
				if rdns == "" {
					explain += " If you don't recognize this process making outbound connections, it's worth investigating."
				}

				findings = append(findings, Finding{
					ID:       fmt.Sprintf("net-external-%s-%s", proc, c.RemoteAddr),
					Module:   "Network",
					Severity: severity,
					Title:    fmt.Sprintf("External connection from %s", proc),
					Detail:   fmt.Sprintf("%s is connected to %s:%d", proc, c.RemoteAddr, c.RemotePort),
					Explain:  explain,
					Item:     proc,
					Metadata: map[string]string{
						"remote_ip":   c.RemoteAddr,
						"remote_port": fmt.Sprintf("%d", c.RemotePort),
						"unique_ips":  fmt.Sprintf("%d", len(uniqueIPs)),
					},
					Timestamp: time.Now(),
				})
				break
			}
		}

		// Many external connections from one process
		if len(external) > 20 {
			findings = append(findings, Finding{
				ID:       fmt.Sprintf("net-manyconn-%s", proc),
				Module:   "Network",
				Severity: SeverityCheck,
				Title:    fmt.Sprintf("High number of external connections: %s", proc),
				Detail:   fmt.Sprintf("%s has %d active external connections", proc, len(external)),
				Explain:  fmt.Sprintf("\"%s\" has an unusually high number of connections to external servers. This could indicate normal behavior for a server application, or something worth investigating.", proc),
				Item:     proc,
				Metadata: map[string]string{"count": fmt.Sprintf("%d", len(external))},
				Timestamp: time.Now(),
			})
		}
	}

	// Known suspicious port usage
	for _, c := range conns {
		if desc, ok := suspiciousPorts[c.RemotePort]; ok && scanner.IsExternalIP(c.RemoteAddr) {
			findings = append(findings, Finding{
				ID:       fmt.Sprintf("net-suspport-%d-%s", c.RemotePort, c.ProcessName),
				Module:   "Network",
				Severity: SeverityWarning,
				Title:    fmt.Sprintf("Connection on suspicious port %d", c.RemotePort),
				Detail:   fmt.Sprintf("%s is connected to %s on port %d (%s)", c.ProcessName, c.RemoteAddr, c.RemotePort, desc),
				Explain:  fmt.Sprintf("Port %d is commonly associated with %s. This is worth investigating unless you intentionally use software that requires this port.", c.RemotePort, desc),
				Item:     c.ProcessName,
				Metadata: map[string]string{"port": fmt.Sprintf("%d", c.RemotePort), "desc": desc},
				Timestamp: time.Now(),
			})
		}
	}

	return findings
}

func uniqueRemoteIPs(conns []scanner.Connection) []string {
	seen := make(map[string]bool)
	var ips []string
	for _, c := range conns {
		if !seen[c.RemoteAddr] {
			seen[c.RemoteAddr] = true
			ips = append(ips, c.RemoteAddr)
		}
	}
	return ips
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
