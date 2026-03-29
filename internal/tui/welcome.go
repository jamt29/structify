package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/jamt29/structify/internal/buildinfo"
)

const welcomeASCII = `
 ██████╗████████╗██████╗ ██╗   ██╗ ██████╗████████╗██╗███████╗██╗   ██╗
██╔════╝╚══██╔══╝██╔══██╗██║   ██║██╔════╝╚══██╔══╝██║██╔════╝╚██╗ ██╔╝
╚█████╗    ██║   ██████╔╝██║   ██║██║        ██║   ██║█████╗   ╚████╔╝
 ╚═══██╗   ██║   ██╔══██╗██║   ██║██║        ██║   ██║██╔══╝    ╚██╔╝
██████╔╝   ██║   ██║  ██║╚██████╔╝╚██████╗   ██║   ██║██║        ██║
╚═════╝    ╚═╝   ╚═╝  ╚═╝ ╚═════╝  ╚═════╝   ╚═╝   ╚═╝╚═╝        ╚═╝
`

// WelcomeView returns the welcome header (ASCII + tagline + version) as one block
// with children aligned to the horizontal center of the block (for use inside the menu).
func WelcomeView(width int) string {
	_ = width // reserved for future responsive art; block is capped by menu MaxWidth
	art := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render(strings.TrimSpace(welcomeASCII))
	tagline := styleWelcomeTagline.Render("Scaffold opinionated projects in seconds")
	version := styleWelcomeVersion.Render(buildinfo.Version)
	return lipgloss.JoinVertical(lipgloss.Center,
		art,
		"",
		tagline,
		version,
	)
}
