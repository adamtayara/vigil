package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorGreen  = lipgloss.Color("#22c55e")
	colorBlue   = lipgloss.Color("#3b82f6")
	colorYellow = lipgloss.Color("#eab308")
	colorOrange = lipgloss.Color("#f97316")
	colorRed    = lipgloss.Color("#ef4444")
	colorGray   = lipgloss.Color("#6b7280")
	colorWhite  = lipgloss.Color("#f9fafb")
	colorDim    = lipgloss.Color("#9ca3af")

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			MarginBottom(1)

	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#a78bfa")).
			MarginBottom(1)

	styleDim = lipgloss.NewStyle().Foreground(colorDim)

	styleSuccess = lipgloss.NewStyle().Foreground(colorGreen)
	styleInfo    = lipgloss.NewStyle().Foreground(colorBlue)
	styleWarn    = lipgloss.NewStyle().Foreground(colorOrange)
	styleError   = lipgloss.NewStyle().Foreground(colorRed)

	styleBadgeClear    = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	styleBadgeHeadsUp  = lipgloss.NewStyle().Foreground(colorBlue).Bold(true)
	styleBadgeCheck    = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	styleBadgeWarning  = lipgloss.NewStyle().Foreground(colorOrange).Bold(true)
	styleBadgeCritical = lipgloss.NewStyle().Foreground(colorRed).Bold(true)

	styleModuleDone    = lipgloss.NewStyle().Foreground(colorGreen)
	styleModuleRunning = lipgloss.NewStyle().Foreground(colorBlue)
	styleModulePending = lipgloss.NewStyle().Foreground(colorGray)

	styleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151")).
			Padding(0, 1)

	styleSummaryBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4b5563")).
			Padding(1, 2).
			MarginTop(1)
)

func severityStyle(s int) lipgloss.Style {
	switch s {
	case 0:
		return styleBadgeClear
	case 1:
		return styleBadgeHeadsUp
	case 2:
		return styleBadgeCheck
	case 3:
		return styleBadgeWarning
	case 4:
		return styleBadgeCritical
	default:
		return styleDim
	}
}
