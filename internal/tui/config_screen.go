package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	if m.width < 80 || m.height < 24 {
		return stylePending.Render("Terminal too small. Minimum 80x24.")
	}

	configFile := m.cfg.ConfigFile
	if strings.TrimSpace(configFile) == "" {
		// Best-effort fallback when file doesn't exist.
		configFile = m.cfg.ConfigDir + "/config.yaml"
	}

	localDir := m.cfg.TemplatesDir
	if strings.TrimSpace(localDir) == "" {
		localDir = "~/.structify/templates/"
	}

	version := "v0.1.4"

	var box strings.Builder
	if m.err != nil {
		box.WriteString(styleErrorText(m.err.Error()))
		box.WriteString("\n\n")
	}
	box.WriteString(fmt.Sprintf("Config file   %s\n", configFile))
	box.WriteString(fmt.Sprintf("Templates     %s  (%d instalados)\n", localDir, m.localCount))
	box.WriteString(fmt.Sprintf("Log level     %s\n", m.cfg.LogLevel))
	box.WriteString(fmt.Sprintf("Version       %s\n", version))
	box.WriteString("\nPara editar: abre el archivo de config con tu editor.")
	box.WriteString("\n\nesc volver")

	content := strings.Join([]string{
		styleHeader.Render("structify") + "  " + styleCompletedLabel.Render("·  Configuración"),
		"",
		styleActiveBox.Render(box.String()),
	}, "\n")

	return centerContent(m.width, m.height, lipgloss.PlaceHorizontal(m.width, lipgloss.Center, content))
}

func styleErrorText(msg string) string {
	return lipgloss.NewStyle().Foreground(colorError).Bold(true).Render("✗ " + msg)
}

