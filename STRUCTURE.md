# Project Structure

```
vigil/
├── cmd/
│   └── vigil/
│       └── main.go                  # Entry point — parses flags, wires everything together
│
├── internal/
│   │
│   ├── scanner/                     # Data collection — reads system state, never modifies
│   │   ├── processes.go             # Process list, CPU/RAM, paths, parent-child
│   │   ├── network.go               # Active TCP/UDP connections, owning process
│   │   ├── disk.go                  # Drive stats, top dirs/files, temp folder sizes
│   │   ├── software.go              # Installed programs + install dates
│   │   ├── tasks.go                 # Scheduled tasks / cron / launchd
│   │   ├── extensions.go            # Browser extension metadata + permissions
│   │   │
│   │   ├── processes_windows.go     # Windows-specific: WMI/Win32_Process
│   │   ├── processes_darwin.go      # macOS-specific: sysctl, proc_info
│   │   ├── processes_linux.go       # Linux-specific: /proc filesystem
│   │   │
│   │   ├── software_windows.go      # Registry HKLM\...\Uninstall
│   │   ├── software_darwin.go       # pkgutil + /Applications mtime
│   │   └── software_linux.go        # apt/dnf/pacman log parsing
│   │
│   ├── analysis/                    # Analysis — takes raw scan data, produces findings
│   │   ├── finding.go               # Finding struct: severity, title, detail, recommendation
│   │   ├── processes.go             # Flag high CPU/RAM, suspicious paths, bad parent-child
│   │   ├── network.go               # Flag unknown IPs, no-rDNS, unusual ports
│   │   ├── disk.go                  # Flag near-full drives, large temp dirs
│   │   ├── software.go              # Flag unknown publishers, bundleware patterns
│   │   ├── tasks.go                 # Flag new/unusual/temp-path tasks
│   │   ├── extensions.go            # Flag broad permissions, sideloaded extensions
│   │   ├── entropy.go               # Shannon entropy calculation on executables (experimental)
│   │   ├── processtree.go           # Process tree builder + anomaly scoring (experimental)
│   │   └── baseline.go              # Load/save/compare behavioral baseline (experimental)
│   │
│   ├── geo/                         # Offline IP geolocation (no external calls)
│   │   ├── lookup.go                # GeoLite2 database query
│   │   └── embed.go                 # Embeds the .mmdb file into the binary
│   │
│   ├── report/                      # Output generation
│   │   ├── html.go                  # Renders findings into the embedded HTML template
│   │   ├── summary.go               # Computes overall score and per-severity counts
│   │   └── open.go                  # Cross-platform "open in browser" (os/exec)
│   │
│   ├── tui/                         # Terminal UI (charmbracelet/bubbletea)
│   │   ├── model.go                 # Bubbletea Model: scan state, progress, findings so far
│   │   ├── view.go                  # Render function: progress bars, spinner, live counts
│   │   └── styles.go                # Lipgloss styles: colors, borders, layout
│   │
│   └── tray/                        # System tray + scheduler
│       ├── tray.go                  # getlantern/systray setup, menu items
│       ├── schedule.go              # robfig/cron integration for scheduled scans
│       └── notify.go                # Cross-platform native OS notifications
│
├── web/                             # Embedded web dashboard (compiled into binary)
│   ├── index.html                   # Dashboard template (Go template syntax)
│   ├── style.css                    # Dashboard styles
│   ├── app.js                       # Interactivity: sorting, filtering, collapsing, charts
│   └── assets/
│       ├── icons/                   # Severity icons (SVG)
│       └── fonts/                   # Any self-hosted fonts (keep it small)
│
├── rules/                           # Bundled detection rules
│   ├── embed.go                     # go:embed directive for rules directory
│   └── *.yar                        # YARA rules for entropy/pattern matching
│
├── data/                            # Bundled offline databases
│   ├── embed.go                     # go:embed directive
│   ├── GeoLite2-City.mmdb           # Offline IP geolocation (MaxMind, free license)
│   └── known-bad-ips.txt            # Curated blocklist, updated each release
│
├── .github/
│   └── workflows/
│       ├── ci.yml                   # Run tests on every PR
│       └── release.yml              # Cross-compile + publish binaries on tag push
│
├── go.mod
├── go.sum
├── README.md
├── SPEC.md
├── STRUCTURE.md
├── LANGUAGE.md
├── LICENSE                          # MIT
└── CONTRIBUTING.md
```

---

## Key Architecture Decisions

### Separation of Scanner and Analysis

`scanner/` only collects raw data (process list, connection table, etc.). `analysis/` takes that raw data and produces `Finding` objects. This means:
- Scanners are easy to unit test with mock data
- Analysis logic is platform-independent
- Easy to add new analysis rules without touching collection code

### Platform-Specific Files via Build Tags

Go's file naming convention (`_windows.go`, `_darwin.go`, `_linux.go`) means the compiler automatically includes only the right file per platform. No `if runtime.GOOS == "windows"` scattered everywhere.

### Single Binary via `go:embed`

The `web/` dashboard, YARA rules, and GeoLite2 database are all embedded directly into the compiled binary using Go's `embed` package. Users download one file. No assets folder, no missing files.

### Bubbletea for Terminal UI

The scan runs in a goroutine and sends progress updates via a channel to the Bubbletea model. The TUI is non-blocking — scanning happens concurrently with the progress display.

### Web Dashboard Served Locally

After scanning, Vigil starts a local HTTP server on `127.0.0.1:7891` and serves the dashboard. It opens the URL in the default browser, then shuts the server down after the first page load (or after 60 seconds, whichever comes first). The HTML file is also saved to disk so users can reopen it without re-running Vigil.

---

## Key Dependencies

| Package | Purpose |
|---|---|
| `github.com/shirou/gopsutil/v3` | Cross-platform process, CPU, memory, disk, network data |
| `github.com/charmbracelet/bubbletea` | Terminal UI framework (Elm architecture) |
| `github.com/charmbracelet/lipgloss` | Terminal styling (colors, borders, layout) |
| `github.com/charmbracelet/bubbles` | Reusable TUI components (spinner, progress bar) |
| `github.com/getlantern/systray` | Cross-platform system tray icon |
| `github.com/robfig/cron/v3` | Cron-style scheduling for tray mode |
| `github.com/spf13/cobra` | CLI flags and subcommands |
| `github.com/oschwald/geoip2-golang` | Read MaxMind GeoLite2 database for offline geolocation |
| `golang.org/x/sys` | Low-level OS calls not covered by gopsutil |

All dependencies are pure Go (no CGO required), enabling trivial cross-compilation.
