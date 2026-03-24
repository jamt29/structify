package config

import (
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	charmlog "github.com/charmbracelet/log"
	"golang.org/x/term"
)

// NewLogger creates the standard Structify logger (stderr) for non-interactive CLI output.
func NewLogger(verbose bool) *charmlog.Logger {
	logger := charmlog.New(os.Stderr)
	if verbose {
		logger.SetLevel(charmlog.DebugLevel)
		logger.SetReportTimestamp(true)
	} else {
		logger.SetLevel(charmlog.InfoLevel)
		logger.SetReportTimestamp(false)
	}
	styles := charmlog.DefaultStyles()
	styles.Levels[charmlog.InfoLevel] = lipgloss.NewStyle().Foreground(lipgloss.Color("#98C379"))
	styles.Levels[charmlog.WarnLevel] = lipgloss.NewStyle().Foreground(lipgloss.Color("#E5C07B"))
	styles.Levels[charmlog.ErrorLevel] = lipgloss.NewStyle().Foreground(lipgloss.Color("#E06C75"))
	logger.SetStyles(styles)
	return logger
}

// UseStructuredLogOut reports whether stdout-style output should use charmbracelet/log on stderr
// instead of formatting to w. It is true only when w is an *os.File that is not a terminal
// (pipe, redirect, CI). *bytes.Buffer and other writers keep the legacy fmt path so tests
// that capture cmd.SetOut(buf) keep stable output.
func UseStructuredLogOut(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return !term.IsTerminal(int(f.Fd()))
}
