<div align="center">

# Vigil

### A privacy-first system health checker for Windows, macOS, and Linux.

*Scan your computer. Get plain-English answers. Nothing leaves your machine.*

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8.svg?style=flat-square&logo=go&logoColor=white)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey.svg?style=flat-square)](https://github.com/adamtayara/vigil/releases)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://github.com/adamtayara/vigil/pulls)
[![Made with Go](https://img.shields.io/badge/Made%20with-Go-00ADD8?style=flat-square&logo=go&logoColor=white)](https://golang.org)

[**Quick Start**](#quick-start) ·
[**Features**](#what-vigil-checks) ·
[**Privacy**](#privacy-guarantees) ·
[**Architecture**](#architecture) ·
[**Roadmap**](#roadmap)

</div>

---

## The Problem

Your computer is constantly being changed by software you didn't install, processes you don't recognize, and extensions that quietly request more permissions than they need. Most diagnostic tools fall into one of three buckets:

- **Bloated commercial suites** that phone home with your data and upsell you on subscriptions
- **Command-line forensics tools** built for security professionals who already know what they're looking for
- **"PC cleaners"** that show scary-looking warnings to sell you a fix

Vigil is the tool I wished existed: a single binary that reads your system, explains what it sees in plain English, and never sends a byte of your data anywhere.

---

## What Vigil Checks

<table>
<tr>
<th align="left">Module</th>
<th align="left">What it looks for</th>
</tr>
<tr>
<td><b>Processes</b></td>
<td>CPU/RAM hogs, suspicious executable paths, unusual parent-child relationships</td>
</tr>
<tr>
<td><b>Network</b></td>
<td>Active connections, what's phoning home, unrecognized external IPs</td>
</tr>
<tr>
<td><b>Disk</b></td>
<td>Where your space went, oversized temp folders, near-full drives</td>
</tr>
<tr>
<td><b>Installed Software</b></td>
<td>Recently added programs, unknown publishers, bundleware patterns</td>
</tr>
<tr>
<td><b>Scheduled Tasks</b></td>
<td>Tasks you didn't set up, tasks running from unusual locations</td>
</tr>
<tr>
<td><b>Browser Extensions</b></td>
<td>Extensions with broad permissions, sideloaded or unreviewed extensions</td>
</tr>
</table>

Each finding ships with a severity badge — **All Clear**, **Heads Up**, **Worth Checking**, or **Investigate** — and a "What does this mean?" explanation written for humans, not engineers.

---

## Quick Start

Download the latest release for your platform from the [**Releases**](../../releases) page, then:

```bash
# Windows
.\vigil.exe

# macOS
./vigil-macos

# Linux
./vigil-linux
```

That's it. Vigil scans your system in 15–45 seconds, then opens an interactive report in your browser.

> **No installer. No runtime dependencies. No account.** One binary, one command.

---

## What the Report Looks Like

The report is a self-contained HTML file — no server, no JavaScript framework, no external resources. It includes:

- **Health score dashboard** with severity counts at a glance
- **Per-module sections** with sortable tables and color-coded findings
- **Plain-English explanations** for every flagged item
- **Full process / connection / software inventories** for power users who want the raw data

The report saves to `vigil-report-YYYY-MM-DD.html` in your home folder. Email it to yourself, share it with a more technical friend, or just delete it — it's yours.

---

## Background Mode

Vigil can sit quietly in your system tray and run scheduled scans in the background.

```bash
.\vigil.exe tray                    # start in tray mode
.\vigil.exe schedule daily          # set scan frequency: daily / weekly / monthly / off
```

Right-click the tray icon to scan immediately, view the last report, change the schedule, or quit. When a scheduled scan finishes, you get a native OS notification — click it to open the report.

---

## Privacy Guarantees

Vigil makes **zero external network connections** during a scan. This isn't a marketing claim — it's an architectural choice:

| Concern | How Vigil handles it |
|---|---|
| IP geolocation | Bundled offline database. No IP ever leaves the machine. |
| DNS lookups | Uses your system resolver (same as your browser). |
| Telemetry | None. No analytics, no crash reports, no version pings. |
| Reports | Saved locally only. Never uploaded. |
| Updates | Manual. You decide when (and whether) to download a new version. |

The entire codebase is under 4,000 lines of Go. You can read every line of what it does in an afternoon.

---

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   cmd/      │     │  internal/  │     │  internal/  │
│   vigil     │────▶│   scanner   │────▶│  analysis   │
│ (CLI entry) │     │ (read state)│     │ (heuristics)│
└─────────────┘     └─────────────┘     └─────────────┘
       │                                        │
       │            ┌─────────────┐             │
       └───────────▶│  internal/  │◀────────────┘
                    │   report    │
                    │ (HTML view) │
                    └─────────────┘
                            │
                    ┌───────┴────────┐
                    │                │
              ┌──────────┐    ┌──────────┐
              │   tui/   │    │  tray/   │
              │ (live    │    │ (system  │
              │  scan)   │    │  tray)   │
              └──────────┘    └──────────┘
```

**Design principles:**

- **Read-only.** Vigil never quarantines, deletes, or modifies anything. Diagnostics, not remediation.
- **Cross-platform without abstraction layers.** OS-specific code lives behind Go build tags (`scanner_windows.go`, `scanner_darwin.go`, `scanner_linux.go`).
- **Self-contained reports.** The HTML report is one file with embedded CSS and JS — no CDN, no internet required to view it later.
- **Heuristic-based, not signature-based.** Vigil flags patterns that *look* suspicious. It doesn't claim to detect specific malware families.

---

## Tech Stack

| Layer | Choice | Why |
|---|---|---|
| Language | **Go 1.23** | Single static binary, great cross-compilation, strong stdlib |
| TUI | **[bubbletea](https://github.com/charmbracelet/bubbletea)** + **[lipgloss](https://github.com/charmbracelet/lipgloss)** | Live progress without a heavy framework |
| System inspection | **[gopsutil](https://github.com/shirou/gopsutil)** | Cross-platform process / network / disk introspection |
| CLI | **[cobra](https://github.com/spf13/cobra)** | Industry-standard subcommand structure |
| Tray | **[systray](https://github.com/getlantern/systray)** | Minimal native tray icon across all three OSes |
| Scheduling | **[robfig/cron](https://github.com/robfig/cron)** | Cron expressions for predictable scheduling |
| Report | **`html/template`** + vanilla CSS/JS | No build step, no dependencies, ships embedded in the binary |

---

## Building From Source

**Prerequisites:** Go 1.23 or newer.

```bash
git clone https://github.com/adamtayara/vigil
cd vigil
go build -o vigil ./cmd/vigil
```

**Cross-compile for all platforms:**

```bash
GOOS=windows GOARCH=amd64 go build -o dist/vigil.exe       ./cmd/vigil
GOOS=darwin  GOARCH=arm64 go build -o dist/vigil-macos-arm ./cmd/vigil
GOOS=darwin  GOARCH=amd64 go build -o dist/vigil-macos-x64 ./cmd/vigil
GOOS=linux   GOARCH=amd64 go build -o dist/vigil-linux     ./cmd/vigil
```

The resulting binary is ~12 MB and has no runtime dependencies.

---

## Project Layout

```
vigil/
├── cmd/vigil/              # CLI entry point
├── internal/
│   ├── scanner/            # OS-level data collection (one file per OS where needed)
│   ├── analysis/           # Heuristics that turn raw data into findings
│   ├── report/             # HTML report generation, embedded template
│   ├── tui/                # Live terminal UI for one-shot scans
│   └── tray/               # Background tray + cron scheduler
├── web/                    # Standalone CSS/JS (also embedded in report)
├── SPEC.md                 # Full product spec
├── STRUCTURE.md            # Codebase tour
└── LANGUAGE.md             # Plain-English copy guidelines
```

---

## Roadmap

| Version | Focus |
|---|---|
| **v0.1** *(current)* | Six core scan modules, HTML report, live TUI, tray mode, scheduling |
| **v0.2** | Behavioral baseline — save a "this is normal for me" snapshot, diff against it |
| **v0.3** | Entropy analysis to flag packed/encrypted executables |
| **v0.4** | Optional local LLM explanations via Ollama (no cloud) |
| **v1.0** | Continuous background monitoring with change-detection alerts |

---

## Contributing

Pull requests, bug reports, and feature ideas are all welcome. Open an [issue](../../issues) to start the conversation.

If you're picking up something new, look for issues tagged [`good first issue`](../../issues?q=is%3Aissue+label%3A%22good+first+issue%22).

---

## License

[MIT](LICENSE) — do whatever you want with it, just keep the copyright notice.

---

<div align="center">

*Vigil reads your system state at a point in time. It is not a replacement for antivirus software. If you suspect active compromise, consult a security professional.*

**Built by [Adam Tayara](https://github.com/adamtayara)**

</div>
