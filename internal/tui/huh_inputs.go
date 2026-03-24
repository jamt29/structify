package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/jamt29/structify/internal/dsl"
)

func heapString(v string) *string {
	p := new(string)
	*p = v
	return p
}

func heapBool(v bool) *bool {
	p := new(bool)
	*p = v
	return p
}

func heapStringSlice(v []string) *[]string {
	p := new([]string)
	*p = append([]string{}, v...)
	return p
}

// BuildHuhForm builds a generic form from manifest inputs and static context.
func BuildHuhForm(inputs []dsl.Input, ctx dsl.Context) (*huh.Form, error) {
	holder := &huhBindingHolder{
		strings: map[string]*string{},
		bools:   map[string]*bool{},
		multis:  map[string]*[]string{},
		contextFn: func() dsl.Context {
			return ctx
		},
	}
	return buildHuhFormWithBindings(inputs, holder)
}

type huhBindingHolder struct {
	strings   map[string]*string
	bools     map[string]*bool
	multis    map[string]*[]string
	contextFn func() dsl.Context
}

func buildHuhFormForApp(inputs []dsl.Input, app *App) (*huh.Form, error) {
	holder := &huhBindingHolder{
		strings: map[string]*string{},
		bools:   map[string]*bool{},
		multis:  map[string]*[]string{},
		contextFn: func() dsl.Context {
			return app.currentWhenContext()
		},
	}
	for id, v := range app.huhString {
		holder.strings[id] = heapString(v)
	}
	for id, v := range app.huhBool {
		holder.bools[id] = heapBool(v)
	}
	for id, v := range app.huhMulti {
		holder.multis[id] = heapStringSlice(v)
	}

	form, err := buildHuhFormWithBindings(inputs, holder)
	if err != nil {
		return nil, err
	}
	for id, p := range holder.strings {
		app.huhString[id] = *p
	}
	for id, p := range holder.bools {
		app.huhBool[id] = *p
	}
	for id, p := range holder.multis {
		app.huhMulti[id] = append([]string{}, *p...)
	}
	return form, nil
}

func buildHuhFormWithBindings(inputs []dsl.Input, holder *huhBindingHolder) (*huh.Form, error) {
	if holder == nil {
		return nil, fmt.Errorf("huh holder is nil")
	}
	groups := make([]*huh.Group, 0, len(inputs))
	for _, in := range inputs {
		id := strings.TrimSpace(in.ID)
		if id == "" {
			continue
		}
		prompt := strings.TrimSpace(in.Prompt)
		if prompt == "" {
			prompt = id
		}
		kind := strings.ToLower(strings.TrimSpace(in.Type))
		var field huh.Field

		switch kind {
		case "string", "path":
			ptr := holder.strings[id]
			if ptr == nil {
				ptr = heapString("")
				holder.strings[id] = ptr
			}
			def, _ := ApplyDefault(in, holder.contextFn())
			f := huh.NewInput().
				Key(id).
				Title(prompt).
				Placeholder(def).
				Value(ptr).
				Validate(func(s string) error {
					return ValidateInputValue(in, s)
				})
			field = f
		case "enum":
			ptr := holder.strings[id]
			if ptr == nil {
				ptr = heapString("")
				holder.strings[id] = ptr
			}
			options := make([]huh.Option[string], 0, len(in.Options))
			for _, opt := range in.Options {
				options = append(options, huh.NewOption(opt, opt))
			}
			f := huh.NewSelect[string]().
				Key(id).
				Title(prompt).
				Options(options...).
				Value(ptr)
			field = f
		case "bool":
			ptr := holder.bools[id]
			if ptr == nil {
				ptr = heapBool(false)
				holder.bools[id] = ptr
			}
			f := huh.NewConfirm().
				Key(id).
				Title(prompt).
				Value(ptr)
			field = f
		case "multiselect":
			ptr := holder.multis[id]
			if ptr == nil {
				ptr = heapStringSlice(nil)
				holder.multis[id] = ptr
			}
			options := make([]huh.Option[string], 0, len(in.Options))
			for _, opt := range in.Options {
				options = append(options, huh.NewOption(opt, opt))
			}
			f := huh.NewMultiSelect[string]().
				Key(id).
				Title(prompt).
				Options(options...).
				Value(ptr)
			field = f
		default:
			continue
		}

		inCopy := in
		group := huh.NewGroup(field)
		group.WithHideFunc(func() bool {
			ask, err := ShouldAskInput(inCopy, holder.contextFn())
			if err != nil {
				return false
			}
			return !ask
		})
		groups = append(groups, group)
	}

	form := huh.NewForm(groups...).
		WithTheme(structifyHuhTheme()).
		WithShowHelp(false)
	return form, nil
}

func structifyHuhTheme() *huh.Theme {
	t := huh.ThemeBase()
	t.Focused.Title = lipgloss.NewStyle().Foreground(colorText).Bold(true)
	t.Focused.Base = lipgloss.NewStyle().
		PaddingLeft(1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderLeft(true).
		BorderForeground(colorActive)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(colorPrimary)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(colorPrimary)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(colorPrimary)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(colorPrimary)

	t.Blurred = t.Focused
	t.Blurred.Base = t.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.Title = t.Blurred.Title.Foreground(colorMuted)
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()
	t.Group.Title = t.Focused.Title
	return t
}
