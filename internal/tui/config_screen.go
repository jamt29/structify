package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jamt29/structify/internal/buildinfo"
	"github.com/jamt29/structify/internal/config"
	"github.com/jamt29/structify/internal/template"
)

type ConfigModel struct {
	width  int
	height int
	done   bool

	cfg config.Config
	err error

	localCount int
}

func NewConfigModel(templates []*template.Template) *ConfigModel {
	cfg, err := config.Load()
	m := &ConfigModel{
		width: 100, height: 30,
		cfg: cfg,
		err: err,
	}
	for _, t := range templates {
		if t == nil {
			continue
		}
		if t.Source == "local" {
			m.localCount++
		}
	}
	return m
}

func (m *ConfigModel) Init() tea.Cmd { return nil }

func (m *ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = ws.Width, ws.Height
		return m, nil
	}
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.done = true
			return m, nil
		}
	}
	return m, nil
}

func (m *ConfigModel) View() string {
	return m.ViewContent()
}

func expandHome(p string) string {
	h, err := os.UserHomeDir()
	if err != nil || h == "" {
		return p
	}
	if strings.HasPrefix(p, h) {
		return "~" + strings.TrimPrefix(p, h)
	}
	return p
}

func (m *ConfigModel) ViewContent() string {
	if m.width < 80 || m.height < 24 {
		return stylePending.Render("Terminal too small. Minimum 80x24.")
	}

	configFile := m.cfg.ConfigFile
	if strings.TrimSpace(configFile) == "" {
		configFile = filepath.Join(m.cfg.ConfigDir, "config.yaml")
	}
	configFile = expandHome(configFile)

	templatesDir := expandHome(m.cfg.TemplatesDir)
	if strings.TrimSpace(templatesDir) == "" {
		templatesDir = "~/.structify/templates/"
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1).
		Width(min(m.width-6, MaxWidthConfig))

	titleStyle := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)

	general := strings.Join([]string{
		titleStyle.Render("Información general"),
		"",
		fmt.Sprintf("  %-16s %s", "Versión", buildinfo.Version),
		fmt.Sprintf("  %-16s %s", "Config file", configFile),
		fmt.Sprintf("  %-16s %s", "Templates dir", templatesDir),
		fmt.Sprintf("  %-16s %d instalados", "Templates local", m.localCount),
	}, "\n")

	prefs := strings.Join([]string{
		titleStyle.Render("Preferencias"),
		"",
		fmt.Sprintf("  %-16s %s", "Log level", m.cfg.LogLevel),
		fmt.Sprintf("  %-16s %t", "Non-interactive", m.cfg.NonInteractive),
	}, "\n")

	var top strings.Builder
	if m.err != nil {
		top.WriteString(styleErrorText(m.err.Error()))
		top.WriteString("\n\n")
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		boxStyle.Render(general),
		"",
		boxStyle.Render(prefs),
		"",
		stylePending.Render("Para editar: "+configFile),
		"",
		styleHelpBar.Render("esc volver"),
	)

	top.WriteString(styleHeader.Render("structify") + "  " + styleCompletedLabel.Render("·  Configuración"))
	top.WriteString("\n\n")
	top.WriteString(body)

	return lipgloss.NewStyle().MaxWidth(MaxWidthConfig).Render(top.String())
}

func styleErrorText(msg string) string {
	return lipgloss.NewStyle().Foreground(colorError).Bold(true).Render("✗ " + msg)
}
