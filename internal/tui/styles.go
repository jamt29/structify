package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary = lipgloss.Color("#C678DD")
	colorSuccess = lipgloss.Color("#98C379")
	colorMuted   = lipgloss.Color("#5C6370")
	colorBorder  = lipgloss.Color("#3E4451")
	colorText    = lipgloss.Color("#ABB2BF")
	colorActive  = lipgloss.Color("#61AFEF")
	colorError   = lipgloss.Color("#E06C75")

	styleHeader = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	styleActiveBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorActive).
			Padding(0, 1)

	styleCompletedLabel = lipgloss.NewStyle().
				Foreground(colorMuted)

	styleCompletedValue = lipgloss.NewStyle().
				Foreground(colorText)

	styleCheckmark = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	stylePending = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleHelpBar = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1)

	styleWelcomeTagline = lipgloss.NewStyle().
				Foreground(colorMuted).
				Italic(true)

	styleWelcomeVersion = lipgloss.NewStyle().
				Foreground(colorBorder)

	styleMenuItem = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 1)

	styleMenuItemActive = lipgloss.NewStyle().
				Foreground(colorText).
				Background(lipgloss.Color("#2d2f3f")).
				BorderLeft(true).
				BorderForeground(colorPrimary).
				Padding(0, 1)

	styleMenuLabel = lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true)

	styleMenuArrow = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)
)
