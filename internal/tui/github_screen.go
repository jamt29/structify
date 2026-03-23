package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type GitHubModel struct {
	width  int
	height int
	done   bool
}

func NewGitHubModel() *GitHubModel {
	return &GitHubModel{width: 100, height: 30}
}

func (m *GitHubModel) Init() tea.Cmd { return nil }

func (m *GitHubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *GitHubModel) View() string {
	if m.width < 80 || m.height < 24 {
		return stylePending.Render("Terminal too small. Minimum 80x24.")
	}

	box := strings.Join([]string{
		"Instala templates de la comunidad directamente desde GitHub.",
		"",
		"Uso:",
		"  structify template add github.com/<user>/<repo>",
		"  structify template add github.com/<user>/<repo>@v1.0",
		"",
		"Recursos:",
		"  docs/template-format.md — cómo crear un template compatible",
		"  github.com/jamt29/structify — templates oficiales",
		"",
		"esc volver",
	}, "\n")

	content := strings.Join([]string{
		styleHeader.Render("structify") + "  " + styleCompletedLabel.Render("·  Explorar GitHub"),
		"",
		styleActiveBox.Render(box),
	}, "\n")

	return centerContent(m.width, m.height, lipgloss.PlaceHorizontal(m.width, lipgloss.Center, content))
}

