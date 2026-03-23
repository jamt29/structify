package tui

import "github.com/charmbracelet/lipgloss"

func centerContent(width int, height int, content string) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func centerContentHorizontal(width int, content string) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, content)
}
