package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	tmpl "github.com/jamt29/structify/internal/template"
)

type githubInstallPollMsg struct{}

type GitHubModel struct {
	width  int
	height int
	done   bool

	urlInput textinput.Model
	spin     spinner.Model

	installing  bool
	installErr  error
	installDone bool
	installCh   chan error
}

func NewGitHubModel() *GitHubModel {
	ti := textinput.New()
	ti.Placeholder = "github.com/user/repo"
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 48

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorActive)

	return &GitHubModel{
		width:    100,
		height:   30,
		urlInput: ti,
		spin:     s,
	}
}

func (m *GitHubModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *GitHubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.installing {
		switch msg := msg.(type) {
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spin, cmd = m.spin.Update(msg)
			return m, cmd
		case tea.WindowSizeMsg:
			m.width, m.height = msg.Width, msg.Height
			return m, nil
		case githubInstallPollMsg:
			select {
			case err := <-m.installCh:
				m.installing = false
				m.installDone = true
				m.installErr = err
				return m, nil
			default:
				return m, tea.Batch(
					m.spin.Tick,
					tea.Tick(time.Millisecond*80, func(time.Time) tea.Msg { return githubInstallPollMsg{} }),
				)
			}
		}
		return m, nil
	}

	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = ws.Width, ws.Height
		return m, nil
	}
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if !m.installing {
				m.done = true
				return m, nil
			}
		case "enter":
			if m.installing {
				return m, nil
			}
			raw := strings.TrimSpace(m.urlInput.Value())
			if raw == "" {
				return m, nil
			}
			ref, err := tmpl.ParseGitHubURL(raw)
			if err != nil {
				m.installDone = true
				m.installErr = err
				return m, nil
			}
			m.installing = true
			m.installDone = false
			m.installErr = nil
			m.installCh = make(chan error, 1)
			go func() {
				client := tmpl.NewGitHubClient()
				m.installCh <- tmpl.InstallFromGitHub(client, ref, tmpl.InstallFromGitHubOptions{})
			}()
			return m, tea.Batch(
				m.spin.Tick,
				tea.Tick(time.Millisecond*80, func(time.Time) tea.Msg { return githubInstallPollMsg{} }),
			)
		}
	}

	if !m.installing {
		var cmd tea.Cmd
		m.urlInput, cmd = m.urlInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *GitHubModel) View() string {
	return m.ViewContent()
}

func (m *GitHubModel) ViewContent() string {
	if m.width < 80 || m.height < 24 {
		return stylePending.Render("Terminal too small. Minimum 80x24.")
	}

	var b strings.Builder
	b.WriteString(styleHeader.Render("structify") + "  " + styleCompletedLabel.Render("·  Explorar GitHub"))
	b.WriteString("\n\n")
	b.WriteString(stylePending.Render("Instala templates de la comunidad desde GitHub."))
	b.WriteString("\n\n")
	b.WriteString(styleCompletedLabel.Render("URL del template:"))
	b.WriteString("\n")
	b.WriteString(styleActiveBox.Width(min(m.width-8, 56)).Render(m.urlInput.View()))
	b.WriteString("\n\n")

	if m.installing {
		b.WriteString(lipgloss.NewStyle().Foreground(colorActive).Render(m.spin.View()))
		b.WriteString(" ")
		b.WriteString(stylePending.Render("clonando e instalando…"))
		b.WriteString("\n")
	} else if m.installDone {
		if m.installErr != nil {
			b.WriteString(lipgloss.NewStyle().Foreground(colorError).Render("✗ "+m.installErr.Error()) + "\n")
		} else {
			b.WriteString(styleCheckmark.Render("✓ Template instalado correctamente.") + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styleHelpBar.Render("enter instalar  esc volver"))
	b.WriteString("\n")
	b.WriteString(stylePending.Render(strings.Repeat("─", min(m.width-4, 52))))
	b.WriteString("\n\n")
	b.WriteString(styleCompletedLabel.Render("Recursos:") + "\n")
	b.WriteString(stylePending.Render("· docs/template-format.md") + "\n")
	b.WriteString(stylePending.Render("· github.com/jamt29/structify") + "\n")

	return lipgloss.NewStyle().MaxWidth(EffectiveMaxWidth(m.width, MaxWidthGitHub)).Align(lipgloss.Left).Render(b.String())
}
