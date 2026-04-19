package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/adamtayara/vigil/internal/analysis"
)

type ModuleStatus struct {
	Name     string
	Done     bool
	Error    bool
	Count    int
	Worst    int
	Duration time.Duration
}

type ModuleDoneMsg struct {
	Name     string
	Findings []analysis.Finding
	Duration time.Duration
	Err      error
}

type ScanDoneMsg struct {
	Result     *analysis.ScanResult
	ReportPath string
}

type ErrorMsg struct {
	Err error
}

type Model struct {
	spinner  spinner.Model
	modules  []ModuleStatus
	current  int
	findings []analysis.Finding
	done     bool
	err      error
	result     *analysis.ScanResult
	reportPath string
	width      int
	startAt    time.Time
}

func NewModel(moduleNames []string) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#a78bfa"))

	modules := make([]ModuleStatus, len(moduleNames))
	for i, name := range moduleNames {
		modules[i] = ModuleStatus{Name: name}
	}

	return Model{
		spinner: sp,
		modules: modules,
		startAt: time.Now(),
		width:   80,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case ModuleDoneMsg:
		for i, mod := range m.modules {
			if mod.Name == msg.Name {
				m.modules[i].Done = true
				m.modules[i].Duration = msg.Duration
				m.modules[i].Count = len(msg.Findings)
				m.modules[i].Error = msg.Err != nil
				if msg.Err == nil {
					m.findings = append(m.findings, msg.Findings...)
					worst := 0
					for _, f := range msg.Findings {
						if int(f.Severity) > worst {
							worst = int(f.Severity)
						}
					}
					m.modules[i].Worst = worst
				}
				m.current = i + 1
				break
			}
		}
		return m, m.spinner.Tick

	case ScanDoneMsg:
		m.done = true
		m.result = msg.Result
		m.reportPath = msg.ReportPath
		return m, tea.Quit

	case ErrorMsg:
		m.err = msg.Err
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return styleError.Render(fmt.Sprintf("Error: %v\n", m.err))
	}

	var b strings.Builder

	// Header
	b.WriteString("\n")
	b.WriteString(styleHeader.Render("  ▶  VIGIL System Health Check"))
	b.WriteString("\n")
	b.WriteString(styleDim.Render(fmt.Sprintf("  Scanning your system... %s\n", formatElapsed(time.Since(m.startAt)))))
	b.WriteString("\n")

	// Module list
	for i, mod := range m.modules {
		line := "  "
		if mod.Done {
			if mod.Error {
				line += styleError.Render("✗") + " "
			} else {
				line += styleSuccess.Render("✓") + " "
			}
			line += styleModuleDone.Render(mod.Name)
			if mod.Count > 0 {
				ws := mod.Worst
				sty := severityStyle(ws)
				line += styleDim.Render(fmt.Sprintf("  %d finding", mod.Count))
				if mod.Count != 1 {
					line += styleDim.Render("s")
				}
				line += "  " + sty.Render(severityLabel(ws))
			} else {
				line += styleDim.Render("  all clear")
			}
			if mod.Duration > 0 {
				line += styleDim.Render(fmt.Sprintf("  (%dms)", mod.Duration.Milliseconds()))
			}
		} else if i == m.current && !m.done {
			line += m.spinner.View() + " "
			line += styleModuleRunning.Render(mod.Name)
			line += styleDim.Render("  scanning...")
		} else {
			line += "  " + styleModulePending.Render(mod.Name)
		}
		b.WriteString(line + "\n")
	}

	if m.done && m.result != nil {
		b.WriteString("\n")
		b.WriteString(renderSummary(m.result, m.reportPath))
	}

	return b.String()
}

func renderSummary(r *analysis.ScanResult, reportPath string) string {
	score := r.HealthScore()
	scoreStyle := styleSuccess
	if score < 80 {
		scoreStyle = styleWarn
	}
	if score < 60 {
		scoreStyle = styleError
	}

	counts := []string{
		styleBadgeCritical.Render(fmt.Sprintf("%d critical", r.Counts.Critical)),
		styleBadgeWarning.Render(fmt.Sprintf("%d investigate", r.Counts.Warning)),
		styleBadgeCheck.Render(fmt.Sprintf("%d worth checking", r.Counts.Check)),
		styleBadgeHeadsUp.Render(fmt.Sprintf("%d heads up", r.Counts.HeadsUp)),
		styleBadgeClear.Render(fmt.Sprintf("%d clear", r.Counts.Clear)),
	}

	content := fmt.Sprintf("Health Score: %s  ·  %s\n\n%s",
		scoreStyle.Bold(true).Render(fmt.Sprintf("%d/100", score)),
		styleDim.Render(r.WorstSeverity().Label()),
		strings.Join(counts, "  "),
	)

	footer := styleSuccess.Render("  Full interactive report opening in your browser.")
	if reportPath != "" {
		footer += "\n  " + styleDim.Render("Saved to: ") + reportPath
	}
	footer += "\n  " + styleDim.Render("If it didn't open, copy the path above into your browser.")

	return styleSummaryBox.Render(content) + "\n\n" + footer + "\n\n"
}

func formatElapsed(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func severityLabel(s int) string {
	labels := []string{"All Clear", "Heads Up", "Worth Checking", "Investigate", "Act Now"}
	if s >= 0 && s < len(labels) {
		return labels[s]
	}
	return ""
}
