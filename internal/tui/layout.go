package tui

import "github.com/charmbracelet/lipgloss"

// CenteringMode selects how RootModel / App place content in the terminal.
type CenteringMode int

const (
	CenterBoth CenteringMode = iota
	CenterHOnly
	CenterNone
)

func centerContent(width int, height int, content string) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func centerContentHorizontal(width int, content string) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, content)
}

// ApplyScreenCentering applies the shared centering policy (RootModel + RunApp).
func ApplyScreenCentering(mode CenteringMode, width, height int, content string) string {
	switch mode {
	case CenterBoth:
		return centerContent(width, height, content)
	case CenterHOnly:
		return centerContentHorizontal(width, content)
	default:
		return content
	}
}

// AppCenteringMode matches the screenNew row in the centering table (v0.5.0).
func AppCenteringMode(s state) CenteringMode {
	switch s {
	case stateProgress:
		return CenterHOnly
	default:
		return CenterBoth
	}
}
