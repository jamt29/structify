package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

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

type transitionPhase int

const (
	transitionIdle transitionPhase = iota
	transitionFadeOut
	transitionFadeIn
)

// ~4–5 ticks at 16ms ≈ 64–80ms total per half.
const transitionAlphaDelta = 0.22

type transitionTickMsg struct{}

type rootPendingKind int

const (
	rootPendingNone rootPendingKind = iota
	rootPendingMenuNew
	rootPendingMenuTemplates
	rootPendingMenuGitHub
	rootPendingMenuConfig
	rootPendingMenuDefaultReset
	rootPendingNewToMenu
	rootPendingTemplatesToMenu
	rootPendingGitHubToMenu
	rootPendingConfigToMenu
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

	transPhase transitionPhase
	transAlpha float64
	pending    rootPendingKind
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
		transPhase:      transitionIdle,
		transAlpha:      1,
	}
}

func transitionTickCmd() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(time.Time) tea.Msg { return transitionTickMsg{} })
}

func (r RootModel) startTransition(kind rootPendingKind) (tea.Model, tea.Cmd) {
	r.pending = kind
	r.transPhase = transitionFadeOut
	r.transAlpha = 1.0
	return r, transitionTickCmd()
}

func (r RootModel) applyPendingTransition() (RootModel, tea.Cmd) {
	kind := r.pending
	r.pending = rootPendingNone

	switch kind {
	case rootPendingMenuNew:
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

	case rootPendingMenuTemplates:
		r.templatesScreen = NewTemplatesModel(r.templates)
		r.screen = screenTemplates
		return r, r.templatesScreen.Init()

	case rootPendingMenuGitHub:
		r.githubScreen = NewGitHubModel()
		r.screen = screenGitHub
		return r, r.githubScreen.Init()

	case rootPendingMenuConfig:
		r.configScreen = NewConfigModel(r.templates)
		r.screen = screenConfig
		return r, r.configScreen.Init()

	case rootPendingMenuDefaultReset:
		r.screen = screenMenu
		r.menu = NewMenuModel()
		r.menu.exitOnSelect = false
		return r, r.menu.Init()

	case rootPendingNewToMenu:
		r.screen = screenMenu
		r.menu = NewMenuModel()
		r.menu.exitOnSelect = false
		r.app = nil
		return r, r.menu.Init()

	case rootPendingTemplatesToMenu:
		r.screen = screenMenu
		r.menu = NewMenuModel()
		r.menu.exitOnSelect = false
		r.templatesScreen = nil
		return r, r.menu.Init()

	case rootPendingGitHubToMenu:
		r.screen = screenMenu
		r.menu = NewMenuModel()
		r.menu.exitOnSelect = false
		r.githubScreen = nil
		return r, r.menu.Init()

	case rootPendingConfigToMenu:
		r.screen = screenMenu
		r.menu = NewMenuModel()
		r.menu.exitOnSelect = false
		r.configScreen = nil
		return r, r.menu.Init()

	default:
		return r, nil
	}
}

func (r RootModel) handleTransitionTick() (tea.Model, tea.Cmd) {
	switch r.transPhase {
	case transitionFadeOut:
		r.transAlpha -= transitionAlphaDelta
		if r.transAlpha <= 0 {
			r.transAlpha = 0
			next, cmd := r.applyPendingTransition()
			r = next
			if r.err != nil {
				r.transPhase = transitionIdle
				r.transAlpha = 1
				return r, cmd
			}
			r.transPhase = transitionFadeIn
			return r, tea.Batch(cmd, transitionTickCmd())
		}
		return r, transitionTickCmd()
	case transitionFadeIn:
		r.transAlpha += transitionAlphaDelta
		if r.transAlpha >= 1 {
			r.transAlpha = 1
			r.transPhase = transitionIdle
			return r, nil
		}
		return r, transitionTickCmd()
	default:
		return r, nil
	}
}

// applyRootTransitionAlpha simulates fade: terminals have no real opacity.
func applyRootTransitionAlpha(content string, alpha float64) string {
	if alpha >= 0.99 {
		return content
	}
	if alpha <= 0.02 {
		lines := strings.Split(content, "\n")
		blank := make([]string, len(lines))
		return strings.Join(blank, "\n")
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#5C6370")).Render(content)
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

	if _, ok := msg.(transitionTickMsg); ok {
		return r.handleTransitionTick()
	}
	if r.transPhase != transitionIdle {
		return r, nil
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
				return r.startTransition(rootPendingMenuNew)
			case ActionTemplates:
				return r.startTransition(rootPendingMenuTemplates)
			case ActionGitHub:
				return r.startTransition(rootPendingMenuGitHub)
			case ActionConfig:
				return r.startTransition(rootPendingMenuConfig)
			default:
				return r.startTransition(rootPendingMenuDefaultReset)
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
			return r.startTransition(rootPendingNewToMenu)
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

		if r.templatesScreen.NeedsReload() {
			sel := r.templatesScreen.SelectAfterReloadName()
			r.templatesScreen.ClearReload()
			all, err := engine.ListAll()
			if err == nil {
				r.templates = all
				r.templatesScreen = NewTemplatesModel(r.templates)
				if sel != "" {
					r.templatesScreen.SelectByName(sel)
				}
				ws := tea.WindowSizeMsg{Width: r.width, Height: r.height}
				r.templatesScreen.Update(ws)
			}
		}

		if r.templatesScreen.done {
			return r.startTransition(rootPendingTemplatesToMenu)
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
			return r.startTransition(rootPendingGitHubToMenu)
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
			return r.startTransition(rootPendingConfigToMenu)
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

func (r RootModel) viewCurrentScreen() string {
	switch r.screen {
	case screenMenu:
		return r.menu.ViewContent()
	case screenNew:
		if r.app == nil {
			return ""
		}
		return r.app.ViewContent()
	case screenTemplates:
		if r.templatesScreen == nil {
			return ""
		}
		return r.templatesScreen.ViewContent()
	case screenGitHub:
		if r.githubScreen == nil {
			return ""
		}
		return r.githubScreen.ViewContent()
	case screenConfig:
		if r.configScreen == nil {
			return ""
		}
		return r.configScreen.ViewContent()
	default:
		return ""
	}
}

func (r RootModel) centeringMode() CenteringMode {
	switch r.screen {
	case screenMenu:
		return CenterBoth
	case screenNew:
		if r.app == nil {
			return CenterHOnly
		}
		return AppCenteringMode(r.app.state)
	case screenTemplates:
		if r.templatesScreen == nil {
			return CenterBoth
		}
		switch r.templatesScreen.Mode() {
		case modeList, modeEdit:
			return CenterHOnly
		default:
			return CenterBoth
		}
	case screenGitHub, screenConfig:
		return CenterBoth
	default:
		return CenterBoth
	}
}

func (r RootModel) View() string {
	if r.width == 0 || r.height == 0 {
		return ""
	}
	content := r.viewCurrentScreen()
	if r.transPhase != transitionIdle {
		content = applyRootTransitionAlpha(content, r.transAlpha)
	}
	return ApplyScreenCentering(r.centeringMode(), r.width, r.height, content)
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
