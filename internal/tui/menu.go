package tui

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var ErrMenuExit = errors.New("menu exited")

type menuAction int
type MenuAction = menuAction

const (
	actionNew menuAction = iota
	actionTemplates
	actionGitHub
	actionConfig
)

const (
	ActionNew       = actionNew
	ActionTemplates = actionTemplates
	ActionGitHub    = actionGitHub
	ActionConfig    = actionConfig
)

type menuItem struct {
	icon   string
	label  string
	desc   string
	action menuAction
}

type MenuModel struct {
	items     []menuItem
	cursor    int
	selected  int
	exitOnSelect bool
	cancelled bool
	width     int
	height    int
}

func NewMenuModel() MenuModel {
	return MenuModel{
		items: []menuItem{
			{icon: "◆", label: "Nuevo proyecto", desc: "crear desde un template", action: actionNew},
			{icon: "▤", label: "Mis templates", desc: "explorar, crear, gestionar", action: actionTemplates},
			{icon: "↓", label: "Explorar GitHub", desc: "instalar templates de la comunidad", action: actionGitHub},
			{icon: "⚙", label: "Configuración", desc: "preferencias globales", action: actionConfig},
		},
		selected: -1,
		exitOnSelect: true,
		width:    100,
		height:   30,
	}
}

func (m MenuModel) Init() tea.Cmd { return nil }

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.items) - 1
			}
		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.items) {
				m.cursor = 0
			}
		case "enter":
			m.selected = m.cursor
			if m.exitOnSelect {
				return m, tea.Quit
			}
			return m, nil
		case "q", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
	if m.width < 80 || m.height < 24 {
		return stylePending.Render("Terminal too small. Minimum 80x24.")
	}

	var items strings.Builder
	for i, it := range m.items {
		row := fmt.Sprintf("%s  %s  %s  %s",
			it.icon,
			styleMenuLabel.Render(padRight(it.label, 20)),
			stylePending.Render(padRight(it.desc, 34)),
			styleMenuArrow.Render("›"),
		)
		if i == m.cursor {
			items.WriteString(styleMenuItemActive.Width(min(74, m.width-6)).Render(row))
		} else {
			items.WriteString(styleMenuItem.Width(min(74, m.width-6)).Render(row))
		}
		items.WriteString("\n")
	}

	content := strings.Join([]string{
		WelcomeView(m.width),
		"",
		items.String(),
		"",
		styleHelpBar.Render(" ↑↓ navegar  enter seleccionar  q salir "),
	}, "\n")

	return centerContent(m.width, m.height, lipgloss.PlaceHorizontal(m.width, lipgloss.Center, content))
}

func (m MenuModel) SelectedAction() menuAction {
	if m.selected < 0 || m.selected >= len(m.items) {
		return actionNew
	}
	return m.items[m.selected].action
}

func RunMenu() (menuAction, error) {
	model := NewMenuModel()
	p := tea.NewProgram(model, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return actionNew, fmt.Errorf("running main menu: %w", err)
	}
	m, ok := final.(MenuModel)
	if !ok {
		return actionNew, fmt.Errorf("unexpected main menu model")
	}
	if m.cancelled || m.selected < 0 {
		return actionNew, ErrMenuExit
	}
	return m.SelectedAction(), nil
}
