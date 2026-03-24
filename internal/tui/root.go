package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/engine"
	"github.com/jamt29/structify/internal/template"
)

type screen int

const (
	screenMenu screen = iota
	screenNew
	screenTemplates
	screenGitHub
	screenConfig
)

type RootModel struct {
	screen screen

	menu            MenuModel
	app             *App
	templatesScreen *TemplatesModel
	githubScreen    *GitHubModel
	configScreen    *ConfigModel

	templates []*template.Template
	engine    *engine.Engine

	width  int
	height int

	err error
}

func NewRootModel(templates []*template.Template, eng *engine.Engine) RootModel {
	m := NewMenuModel()
	m.exitOnSelect = false

	return RootModel{
		screen:          screenMenu,
		menu:            m,
		templates:       templates,
		engine:          eng,
		templatesScreen: nil,
		githubScreen:    nil,
		configScreen:    nil,
		width:           80,
		height:          24,
	}
}

func (r RootModel) Init() tea.Cmd {
	sizeCmd := initialWindowSizeCmd()
	switch r.screen {
	case screenMenu:
		return tea.Batch(r.menu.Init(), sizeCmd)
	case screenNew:
		if r.app != nil {
			return tea.Batch(r.app.Init(), sizeCmd)
		}
		return sizeCmd
	case screenTemplates:
		if r.templatesScreen != nil {
			return tea.Batch(r.templatesScreen.Init(), sizeCmd)
		}
		return sizeCmd
	case screenGitHub:
		if r.githubScreen != nil {
			return tea.Batch(r.githubScreen.Init(), sizeCmd)
		}
		return sizeCmd
	case screenConfig:
		if r.configScreen != nil {
			return tea.Batch(r.configScreen.Init(), sizeCmd)
		}
		return sizeCmd
	default:
		return sizeCmd
	}
}

func (r RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Keep window sizing consistent across screens.
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		r.width, r.height = ws.Width, ws.Height

		var cmds []tea.Cmd

		// Always update the menu model (it exists as a value).
		newMenu, cmd := r.menu.Update(msg)
		r.menu = newMenu.(MenuModel)
		cmds = append(cmds, cmd)

		// Update all active sub-models (non-nil pointers).
		if r.app != nil {
			newApp, cmd := r.app.Update(msg)
			r.app = newApp.(*App)
			cmds = append(cmds, cmd)
		}
		if r.templatesScreen != nil {
			newM, cmd := r.templatesScreen.Update(msg)
			r.templatesScreen = newM.(*TemplatesModel)
			cmds = append(cmds, cmd)
		}
		if r.githubScreen != nil {
			newM, cmd := r.githubScreen.Update(msg)
			r.githubScreen = newM.(*GitHubModel)
			cmds = append(cmds, cmd)
		}
		if r.configScreen != nil {
			newM, cmd := r.configScreen.Update(msg)
			r.configScreen = newM.(*ConfigModel)
			cmds = append(cmds, cmd)
		}

		return r, tea.Batch(cmds...)
	}

	switch r.screen {
	case screenMenu:
		prevSelected := r.menu.selected
		newMenu, cmd := r.menu.Update(msg)
		r.menu = newMenu.(MenuModel)

		// Only transition on the "enter" key: selected moves from -1 -> >=0.
		if prevSelected < 0 && r.menu.selected >= 0 {
			action := r.menu.SelectedAction()
			r.menu.selected = -1

			switch action {
			case ActionNew:
				app, err := newApp(r.templates, r.engine)
				if err != nil {
					r.err = err
					return r, tea.Quit
				}
				r.app = app
				r.screen = screenNew
				ws := tea.WindowSizeMsg{Width: r.width, Height: r.height}
				newApp, _ := r.app.Update(ws)
				r.app = newApp.(*App)
				return r, r.app.Init()
			case ActionTemplates:
				r.templatesScreen = NewTemplatesModel(r.templates)
				r.screen = screenTemplates
				return r, r.templatesScreen.Init()
			case ActionGitHub:
				r.githubScreen = NewGitHubModel()
				r.screen = screenGitHub
				return r, r.githubScreen.Init()
			case ActionConfig:
				r.configScreen = NewConfigModel(r.templates)
				r.screen = screenConfig
				return r, r.configScreen.Init()
			default:
				r.screen = screenMenu
				r.menu = NewMenuModel()
				r.menu.exitOnSelect = false
				return r, r.menu.Init()
			}
		}

		return r, cmd

	case screenNew:
		if r.app == nil {
			r.screen = screenMenu
			r.menu = NewMenuModel()
			r.menu.exitOnSelect = false
			return r, r.menu.Init()
		}

		newApp, cmd := r.app.Update(msg)
		r.app = newApp.(*App)

		if r.app.Done() {
			r.screen = screenMenu
			r.menu = NewMenuModel()
			r.menu.exitOnSelect = false
			r.app = nil
			return r, r.menu.Init()
		}

		return r, cmd

	case screenTemplates:
		if r.templatesScreen == nil {
			r.screen = screenMenu
			r.menu = NewMenuModel()
			r.menu.exitOnSelect = false
			return r, r.menu.Init()
		}

		newM, cmd := r.templatesScreen.Update(msg)
		r.templatesScreen = newM.(*TemplatesModel)

		if r.templatesScreen.transitionToNew != nil {
			app, err := newApp(r.templates, r.engine)
			if err != nil {
				r.err = err
				return r, tea.Quit
			}
			app.selected = r.templatesScreen.transitionToNew
			app.answers = dsl.Context{}
			app.prepareInputs()
			app.state = stateInputs
			app.done = false
			app.activeInput = 0
			app.compactForm = len(app.inputs) <= 3

			r.app = app
			ws := tea.WindowSizeMsg{Width: r.width, Height: r.height}
			newApp, _ := r.app.Update(ws)
			r.app = newApp.(*App)
			r.templatesScreen.transitionToNew = nil
			r.templatesScreen.detail = nil
			r.screen = screenNew
			return r, r.app.Init()
		}

		if r.templatesScreen.done {
			r.screen = screenMenu
			r.menu = NewMenuModel()
			r.menu.exitOnSelect = false
			r.templatesScreen = nil
			return r, r.menu.Init()
		}

		return r, cmd

	case screenGitHub:
		if r.githubScreen == nil {
			r.screen = screenMenu
			r.menu = NewMenuModel()
			r.menu.exitOnSelect = false
			return r, r.menu.Init()
		}
		newM, cmd := r.githubScreen.Update(msg)
		r.githubScreen = newM.(*GitHubModel)
		if r.githubScreen.done {
			r.screen = screenMenu
			r.menu = NewMenuModel()
			r.menu.exitOnSelect = false
			r.githubScreen = nil
			return r, r.menu.Init()
		}
		return r, cmd

	case screenConfig:
		if r.configScreen == nil {
			r.screen = screenMenu
			r.menu = NewMenuModel()
			r.menu.exitOnSelect = false
			return r, r.menu.Init()
		}
		newM, cmd := r.configScreen.Update(msg)
		r.configScreen = newM.(*ConfigModel)
		if r.configScreen.done {
			r.screen = screenMenu
			r.menu = NewMenuModel()
			r.menu.exitOnSelect = false
			r.configScreen = nil
			return r, r.menu.Init()
		}
		return r, cmd

	default:
		return r, nil
	}
}

func initialWindowSizeCmd() tea.Cmd {
	return func() tea.Msg {
		w, h, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil || w <= 0 || h <= 0 {
			return nil
		}
		return tea.WindowSizeMsg{Width: w, Height: h}
	}
}

func (r RootModel) View() string {
	var content string
	switch r.screen {
	case screenMenu:
		content = r.menu.ViewContent()
	case screenNew:
		if r.app == nil {
			return ""
		}
		content = r.app.ViewContent()
	case screenTemplates:
		if r.templatesScreen == nil {
			return ""
		}
		content = r.templatesScreen.ViewContent()
	case screenGitHub:
		if r.githubScreen == nil {
			return ""
		}
		content = r.githubScreen.ViewContent()
	case screenConfig:
		if r.configScreen == nil {
			return ""
		}
		content = r.configScreen.ViewContent()
	default:
		return ""
	}

	// RootModel is the only place applying alignment/centering.
	switch r.screen {
	case screenMenu:
		return centerContent(r.width, r.height, content)
	case screenTemplates, screenNew:
		// Default for these screens: horizontal-only.
		switch r.screen {
		case screenTemplates:
			return centerContentHorizontal(r.width, content)
		case screenNew:
			if r.app == nil {
				return centerContentHorizontal(r.width, content)
			}
			switch r.app.state {
			case stateProgress:
				return centerContentHorizontal(r.width, content)
			case stateSelectTemplate, stateInputs, stateConfirm, stateDone, stateError:
				return centerContent(r.width, r.height, content)
			default:
				return centerContentHorizontal(r.width, content)
			}
		}
	}

	// screenGitHub/screenConfig: center both.
	return centerContent(r.width, r.height, content)
}

// Run lanza la sesión TUI completa (welcome + menú + sub-pantallas).
func Run(templates []*template.Template, eng *engine.Engine) error {
	if eng == nil {
		return fmt.Errorf("engine is nil")
	}

	m := NewRootModel(templates, eng)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return fmt.Errorf("running root ui: %w", err)
	}

	fm, ok := final.(RootModel)
	if !ok {
		return fmt.Errorf("unexpected root model")
	}
	return fm.err
}
