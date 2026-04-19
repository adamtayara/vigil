# Vigil

**A privacy-first system health checker for Windows, macOS, and Linux.**

Vigil scans your computer, explains what it finds in plain English, and produces a clean interactive report — all without sending a single byte of your data anywhere.

No account. No subscription. No cloud. Just answers.

---

## Why Vigil?

Most diagnostic tools are either bloated commercial products that phone home, complex command-line tools built for security professionals, or shallow cleaners that fix nothing. Vigil is different:

- **Everything stays on your machine.** No telemetry, no external API calls, no analytics.
- **Explains findings in plain English.** Not "PID 4820 exhibiting anomalous parent-child fork behavior" — just "this program is running from an unusual location and that's worth checking."
- **Reads your system. Never touches it.** Vigil diagnoses. It doesn't quarantine, delete, or modify anything.
- **One file.** Download, run, done. No installer, no runtime dependencies.

---

## What Vigil Checks

| Module | What it looks for |
|---|---|
| **Processes** | CPU/RAM hogs, suspicious executable paths, unusual parent-child relationships |
| **Network** | Active connections, what's phoning home, unrecognized external IPs |
| **Disk** | Where your space went, oversized temp folders, near-full drives |
| **Installed Software** | Recently added programs, unknown publishers, bundleware patterns |
| **Scheduled Tasks** | Tasks you didn't set up, tasks running from unusual locations |
| **Browser Extensions** | Extensions with broad permissions, sideloaded/unreviewed extensions |

---

## Quick Start

**Download** the latest release for your platform from the [Releases](../../releases) page.

```
# Windows — double-click vigil.exe, or:
.\vigil.exe

# macOS
./vigil-macos

# Linux
./vigil-linux
```

Vigil scans your system (typically 15–45 seconds), then automatically opens an interactive report in your browser.

---

## What the Report Looks Like

The report is a self-contained HTML file that opens in your browser. It shows:

- **Summary dashboard** — overall health score with counts by severity level
- **Per-module sections** — sortable tables, color-coded findings
- **Plain-English explanations** — every finding has a "What does this mean?" section
- **Severity badges** — from All Clear (green) to Act Now (red)

The report is saved as `vigil-report-YYYY-MM-DD.html` in your home folder.

---

## Scheduled Scans

Vigil can sit quietly in your system tray and run scans on a schedule you choose.

```
# Start in tray mode
./vigil --tray
```

Right-click the tray icon to:
- Run a scan immediately
- View your last report
- Set a schedule (daily / weekly / monthly / custom)
- Quit Vigil

When a scheduled scan completes, you'll get a native notification summarizing what was found. Click it to open the report.

---

## Experimental Features

These are off by default and can be enabled with flags:

| Flag | Feature |
|---|---|
| `--entropy` | Scan executables for high entropy (packed/encrypted files) |
| `--process-tree` | Visual process tree with anomaly scoring in the report |
| `--baseline` | Save this scan as your "normal" state for future comparison |
| `--compare` | Compare against saved baseline and highlight changes |
| `--llm` | Use local Ollama to generate plain-English explanations (requires Ollama installed) |

---

## Privacy

Vigil makes **zero external network connections** during a scan.

- IP geolocation uses a bundled offline database (no IP is sent anywhere)
- DNS lookups use your system's existing resolver (the same one your browser uses)
- No usage data, crash reports, or analytics of any kind
- Reports are saved locally — never uploaded
- The optional LLM feature talks only to your local Ollama instance

Vigil is open source. Read every line of what it does.

---

## Building from Source

```bash
# Prerequisites: Go 1.22+
git clone https://github.com/adamtayara/vigil
cd vigil
go build -o vigil ./cmd/vigil

# Cross-compile
GOOS=windows GOARCH=amd64 go build -o vigil.exe ./cmd/vigil
GOOS=darwin  GOARCH=arm64 go build -o vigil-mac   ./cmd/vigil
GOOS=linux   GOARCH=amd64 go build -o vigil-linux ./cmd/vigil
```

---

## Roadmap

- **v1.0** — Core six modules, HTML report, tray mode, scheduled scans
- **v1.1** — Behavioral baseline comparison, entropy analysis
- **v1.2** — Offline LLM explanations via Ollama
- **v2.0** — Continuous background monitoring, change detection alerts

---

## Contributing

Contributions welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) first.

Good first issues are tagged [`good first issue`](../../issues?q=is%3Aissue+label%3A%22good+first+issue%22) in the tracker.

---

## License

MIT — see [LICENSE](LICENSE).

---

*Vigil reads your system state at a point in time. It is not a replacement for antivirus software. If you suspect active compromise, consult a security professional.*
