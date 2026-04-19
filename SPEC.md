# Vigil — Feature Specification v1.0

> Privacy-first system health checker for Windows, macOS, and Linux.
> Everything runs locally. Nothing leaves your machine.

---

## Overview

Vigil is a one-shot system diagnostics tool that scans your computer, explains what it finds in plain English, and produces a clean interactive report. It does not send data anywhere, does not require an account, and does not run as a background service unless you explicitly schedule it.

---

## Core Principles

| Principle | Implication |
|---|---|
| Privacy-first | Zero external network calls during a scan. No telemetry. No analytics. |
| Plain language | Every finding has a human-readable explanation. No raw hex, no log dumps. |
| Non-destructive | Vigil reads system state. It never modifies, quarantines, or deletes anything. |
| Lightweight | Single binary. No installer. No runtime dependencies. Launches in under 2 seconds. |
| Open source | MIT licensed. Auditable. No black-box behavior. |

---

## V1 Feature Modules

### 1. Process Audit

**Goal:** Show what's running, how much it's consuming, and flag anything that looks out of place.

**What it collects:**
- All running processes: name, PID, CPU%, RAM (MB), executable path, parent PID
- Process uptime (time since started)
- Process owner (which user account spawned it)

**Analysis:**
- Top 10 CPU consumers and top 10 RAM consumers highlighted
- Process tree reconstruction (parent → child relationships visualized)
- **Suspicious path detection:** Flag processes running from:
  - `%TEMP%` / `/tmp`
  - `AppData\Local\Temp`
  - Unusual deeply-nested paths
  - Paths with random-looking directory names (high character entropy)
- **Anomalous parent-child flags:** Known risky patterns such as:
  - Word / Excel / browser spawning `cmd.exe`, `powershell.exe`, `wscript.exe`
  - System processes (`svchost`, `lsass`) with unexpected children
  - Processes masquerading as system processes but running from wrong paths

**User-facing output (examples):**
> "**Chrome Helper** is using 18% of your CPU. This is normal if you have many tabs open — try closing unused tabs."
> "**svchost.exe** is running from an unusual location (C:\Users\you\AppData\Temp) — system processes shouldn't be here."

---

### 2. Network Connection Audit

**Goal:** Show what your computer is communicating with right now.

**What it collects:**
- All active TCP and UDP connections
- Local port, remote IP, remote port, connection state (ESTABLISHED, LISTEN, etc.)
- Which process owns each connection

**Analysis:**
- Reverse DNS resolution on all external IPs (runs locally via system DNS — no external API)
- Offline IP geolocation (bundled GeoLite2 database — no API calls)
- Flag connections to:
  - IPs with no reverse DNS resolution (common for C2 servers)
  - Multiple simultaneous connections from one process to many different external IPs
  - Unusual ports (non-HTTP/HTTPS outbound from non-browser processes)
  - Known-bad IP ranges (bundled offline list, updated with each release)
- Show which processes are making the most external connections

**Privacy note:** Geolocation uses an embedded offline database. No IP is sent anywhere.

**User-facing output (examples):**
> "**Spotify** is connected to 3 servers in Sweden. This is expected for a music streaming app."
> "**unknown.exe** is connected to an IP address in Russia with no website name. This is worth investigating."

---

### 3. Disk Usage Breakdown

**Goal:** Show where your disk space is going and flag anything abnormal.

**What it collects:**
- All mounted drives: total, used, free space
- Top 15 largest directories (by total size, recursively)
- Top 20 largest individual files
- Temp directory sizes (`%TEMP%`, browser caches, system temp)

**Analysis:**
- Flag drives that are more than 85% full (performance degrades significantly)
- Flag oversized temp/cache folders (often a free easy win — can safely clean)
- Flag the Windows hiberfil.sys and pagefile.sys sizes with explanation
- Flag very large files in unexpected locations

**User-facing output (examples):**
> "Your C: drive is 91% full. Windows needs free space to run well — consider clearing Downloads or old installers."
> "Your temp folder is 8.4 GB. This is safe to delete and could free up significant space."

---

### 4. Recently Installed Software

**Goal:** Show what's been added to your computer recently — useful for spotting software you didn't knowingly install.

**What it collects:**
- All installed programs with their install date
- Filtered to last 30 days by default (configurable)
- Publisher name, version, install location

**Platform specifics:**
- **Windows:** Registry `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`
- **macOS:** `/Applications` modification dates + `pkgutil --pkgs`
- **Linux:** Package manager history (`apt`, `dnf`, `pacman` logs)

**Analysis:**
- Sort by install date (newest first)
- Flag software with no publisher / unknown publisher
- Flag software installed from temp or unusual directories
- Flag multiple pieces of software installed within minutes of each other (common with bundleware installers)

**User-facing output (examples):**
> "**Coupon Finder Pro** was installed 3 days ago and has no verified publisher. You may want to check if you intended to install this."
> "**5 programs** were installed within 2 minutes of each other on April 15th — this sometimes happens with bundled software installers."

---

### 5. Scheduled Tasks Audit

**Goal:** Find tasks set to run automatically, especially ones you didn't set up.

**What it collects:**
- All scheduled tasks / cron jobs
- Task name, trigger (what makes it run), action (what it runs), last run time, next run time, created date, task owner

**Platform specifics:**
- **Windows:** Task Scheduler XML via `schtasks /query`
- **macOS:** LaunchAgents and LaunchDaemons (`~/Library/LaunchAgents`, `/Library/LaunchAgents`, `/Library/LaunchDaemons`)
- **Linux:** User and system crontabs, `/etc/cron.d/`, systemd timers

**Analysis:**
- Flag tasks that:
  - Run executables from temp or unusual directories
  - Were created in the last 30 days (new = worth reviewing)
  - Run as SYSTEM with no obvious legitimate purpose
  - Have obfuscated or random-looking names
  - Run very frequently (every minute or every few minutes — common for persistence mechanisms)
- Flag tasks whose executable no longer exists on disk (orphaned tasks — sometimes left by malware)

**User-facing output (examples):**
> "**UpdateTask_a4f2** runs every 5 minutes from your Temp folder. This is unusual and worth investigating — most legitimate update tasks run once a day or less."

---

### 6. Browser Extension Audit

**Goal:** Surface browser extensions and flag ones with excessive permissions or unknown origins.

**Supported browsers:** Chrome, Firefox, Edge, Brave (where installed)

**What it collects:**
- Extension name, version, ID
- Permissions requested
- Install source (Chrome Web Store, Firefox Add-ons, sideloaded/unpacked)
- Last updated date

**Platform specifics:**
- Reads directly from browser profile directories (no browser interaction needed)
- Chrome/Edge/Brave: `%LOCALAPPDATA%\Google\Chrome\User Data\Default\Extensions\`
- Firefox: `%APPDATA%\Mozilla\Firefox\Profiles\*/extensions/`

**Analysis:**
- Flag extensions with broad permissions:
  - "Access data on all websites"
  - "Read browsing history"
  - "Read and change all your data on all websites"
- Flag sideloaded extensions (not from official store) — these bypass review processes
- Flag extensions that haven't been updated in over 2 years (abandoned, potential risk)
- Flag extensions with very few users and broad permissions

**User-facing output (examples):**
> "**PDF Converter Pro** (Chrome) can read your data on every website you visit. Extensions with this permission can see passwords and banking info — only keep it if you trust it."
> "**This extension** is not from the Chrome Web Store. Unofficial extensions are not reviewed for safety."

---

## Experimental Features (V1 Optional)

### Visual Process Tree

An interactive tree in the HTML dashboard showing parent-child process relationships. Suspicious relationships highlighted in orange/red. Click a process node to see its full details.

### Entropy Analysis

Scans executables in standard install locations (`Program Files`, `Program Files (x86)`, `AppData\Local`) for high Shannon entropy (> 7.2 out of 8.0). High entropy in an executable suggests it may be packed or encrypted — a common technique used by malware to evade signature detection.

- Only scans `.exe` and `.dll` files
- Skips files > 50MB (performance)
- Shows entropy score as a visual meter in the report
- Explains what entropy means in plain language

### Behavioral Baseline (V1 Optional, Opt-in)

After the first scan, Vigil can save a local snapshot of your "normal" state. On subsequent scans, it compares against this baseline and highlights:
- New processes that weren't running before
- New network connections to external IPs
- New scheduled tasks
- Changes in resource consumption patterns

Baseline data stored locally at `~/.vigil/baseline.json`. Never transmitted anywhere.

---

## Severity System

| Level | Label | Color | Meaning |
|---|---|---|---|
| 0 | All Clear | Green | Nothing unusual found |
| 1 | Heads Up | Blue | Worth knowing, no action needed |
| 2 | Worth Checking | Yellow | Unusual but may be legitimate |
| 3 | Investigate | Orange | Genuinely suspicious pattern |
| 4 | Act Now | Red | Strong indicator of a problem |

Every finding is assigned one of these levels. The report opens at a summary showing counts per level.

---

## Report Output

### Interactive HTML Dashboard

- Auto-opens in default browser after scan completes
- Saved as `vigil-report-YYYY-MM-DD.html` in user's home directory (or configured output dir)
- Self-contained single HTML file (no external resources — works offline)
- Sections: Summary → Processes → Network → Disk → Software → Tasks → Extensions
- Collapsible sections, sortable tables, severity filters
- "What does this mean?" expandable explanations on every finding
- Print-friendly stylesheet for PDF export via browser

### Live Terminal Progress

During the scan, Vigil shows a real-time progress view in the terminal using styled components:
- Overall progress bar
- Current module being scanned
- Live count of findings as they're discovered
- Animated spinner with module name
- Color-coded severity counts updating in real time

---

## Scheduled Scans (System Tray)

When Vigil is running in tray mode:
- Small icon in system tray (Windows taskbar / macOS menu bar)
- Right-click menu: "Scan Now", "View Last Report", "Schedule Settings", "Quit Vigil"
- Schedule options: Manual only / Daily / Weekly / Monthly / Custom (cron-style)
- On scan completion: native OS notification with summary ("3 items need your attention — click to view report")
- Clicking notification opens the report in browser

---

## Privacy Guarantees

- **No network calls during scan.** Vigil does not contact any server. The only network activity during a scan is the DNS lookups it performs to resolve IPs you're already connected to — using your system's existing DNS resolver.
- **No telemetry.** No usage data, crash reports, or analytics of any kind.
- **No cloud storage.** Reports are saved locally only.
- **Offline-capable.** Geolocation uses a bundled GeoLite2 database. YARA rules are bundled. No API keys required.
- **Read-only.** Vigil never modifies, quarantines, or deletes files.
- **LLM explanations (optional).** If the Ollama integration is enabled, queries go to your local Ollama instance only — nothing leaves your machine.

---

## Out of Scope for V1

- Real-time monitoring / continuous background scanning (V2)
- File quarantine or remediation actions (by design — Vigil only diagnoses)
- Full antivirus signature scanning (too large, use Windows Defender / ClamAV for this)
- Memory dump analysis
- Startup items (explicitly deferred — covered partially by Scheduled Tasks)
- Mobile or network device scanning

---

## Target Platforms

| Platform | Build Target | Notes |
|---|---|---|
| Windows 11/10 | `GOOS=windows GOARCH=amd64` | Primary target |
| macOS (Intel) | `GOOS=darwin GOARCH=amd64` | Full support |
| macOS (Apple Silicon) | `GOOS=darwin GOARCH=arm64` | Full support |
| Linux (x64) | `GOOS=linux GOARCH=amd64` | Full support |
| Linux (ARM64) | `GOOS=linux GOARCH=arm64` | Raspberry Pi etc. |
