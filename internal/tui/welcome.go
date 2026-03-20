package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const welcomeASCII = `██████╗████████╗██████╗ ██╗   ██╗ ██████╗████████╗██╗███████╗██╗   ██╗
██╔════╝╚══██╔══╝██╔══██╗██║   ██║██╔════╝╚══██╔══╝██║██╔════╝╚██╗ ██╔╝
╚█████╗    ██║   ██████╔╝██║   ██║██║        ██║   ██║█████╗   ╚████╔╝
╚═══██╗   ██║   ██╔══██╗██║   ██║██║        ██║   ██║██╔══╝    ╚██╔╝
██████╔╝   ██║   ██║  ██║╚██████╔╝╚██████╗   ██║   ██║██║        ██║
╚═════╝    ╚═╝   ╚═╝  ╚═╝ ╚═════╝  ╚═════╝   ╚═╝   ╚═╝╚═╝        ╚═╝`

func WelcomeView(width int) string {
	art := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render(welcomeASCII)
	tagline := styleWelcomeTagline.Render("Scaffold opinionated projects in seconds")
	version := styleWelcomeVersion.Render("v0.1.4")

	content := strings.Join([]string{
		art,
		"",
		lipgloss.PlaceHorizontal(width, lipgloss.Center, tagline),
		lipgloss.PlaceHorizontal(width, lipgloss.Center, version),
	}, "\n")
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, content)
}
