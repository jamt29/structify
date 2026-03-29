package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// EffectiveMaxWidth caps a layout max width to the terminal so the widest line is
// strictly narrower than termWidth. lipgloss.PlaceHorizontal is a no-op when
// gap := width - contentWidth is <= 0 (e.g. content as wide as the terminal).
func EffectiveMaxWidth(termWidth, layoutCap int) int {
	if layoutCap < 1 {
		layoutCap = 1
	}
	if termWidth <= 2 {
		return layoutCap
	}
	return min(layoutCap, termWidth-2)
}

// CenteringMode selects how RootModel / App place content in the terminal.
type CenteringMode int

const (
	CenterBoth CenteringMode = iota
	CenterHOnly
	CenterNone
)

func centerContent(width int, height int, content string) string {
	if width <= 0 || height <= 0 {
		if width <= 0 {
			return content
		}
		return centerContentHorizontal(width, content)
	}
	ch := lipgloss.Height(content)
	paddingTop := (height - ch) / 2
	if paddingTop < 0 {
		paddingTop = 0
	}
	if paddingTop == 0 && ch > height {
		return centerContentHorizontal(width, content)
	}
	topPad := strings.Repeat("\n", paddingTop)
	return centerContentHorizontal(width, topPad+content)
}

func centerContentHorizontal(width int, content string) string {
	if width <= 0 {
		return content
	}
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

// AppCenteringMode selects centering for the new-project flow (embedded or RunApp).
func AppCenteringMode(_ state) CenteringMode {
	return CenterBoth
}
