package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/template"
)

func TestTemplateItem_Methods(t *testing.T) {
	tpl := &template.Template{
		Manifest: &dsl.Manifest{
			Name:          "My Template",
			Architecture: "clean",
			Language:      "go",
			Description:   "Desc",
			Tags:          []string{"go", "clean"},
		},
	}

	ti := templateItem{t: tpl}
	if got := ti.Title(); got != "My Template" {
		t.Fatalf("Title: got %q", got)
	}
	wantDesc := "clean · go\nDesc"
	if got := ti.Description(); got != wantDesc {
		t.Fatalf("Description: got %q want %q", got, wantDesc)
	}

	gotFilter := ti.FilterValue()
	if gotFilter == "" {
		t.Fatalf("FilterValue should not be empty")
	}
	if !strings.Contains(gotFilter, "my template") {
		t.Fatalf("FilterValue should contain name; got %q", gotFilter)
	}
	if !strings.Contains(gotFilter, "clean ·") && !strings.Contains(gotFilter, "clean") {
		// loose check: exact string depends on joined spacing
		t.Fatalf("FilterValue should contain architecture/language; got %q", gotFilter)
	}
}

func TestSelectorModel_ViewAndUpdate(t *testing.T) {
	tpl := &template.Template{
		Manifest: &dsl.Manifest{Name: "A", Language: "go", Architecture: "clean", Description: "d"},
	}
	items := []list.Item{
		templateItem{t: tpl},
	}
	l := list.New(items, list.NewDefaultDelegate(), 80, 10)
	m := selectorModel{list: l}

	// View when not selected.
	if v := m.View(); v == "" {
		t.Fatalf("expected non-empty View when not selected")
	}

	// Update with enter should select the current item.
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	sm, ok := m2.(selectorModel)
	if !ok {
		t.Fatalf("unexpected model type %T", m2)
	}
	if sm.selected == nil {
		t.Fatalf("expected selection after enter")
	}
	if sm.selected != tpl {
		t.Fatalf("unexpected selected template")
	}

	// View when selected.
	if v := sm.View(); v != "" {
		t.Fatalf("expected empty View when selected, got %q", v)
	}
}

func TestSummary_NextStepsAndCtxString(t *testing.T) {
	ctx := map[string]interface{}{"project_name": "myapp"}
	if got := ctxString(ctx, "project_name"); got != "myapp" {
		t.Fatalf("ctxString: got %q", got)
	}
	if got := ctxString(nil, "project_name"); got != "" {
		t.Fatalf("ctxString: expected empty, got %q", got)
	}

	steps := nextSteps("go", "myapp")
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}
	if !strings.Contains(steps[0], "cd myapp") {
		t.Fatalf("unexpected go step[0]=%q", steps[0])
	}

	steps = nextSteps("unknown", "")
	if len(steps) != 1 {
		t.Fatalf("expected 1 step for unknown, got %d", len(steps))
	}
	if steps[0] == "" {
		t.Fatalf("expected non-empty next steps")
	}
}

func TestProgressModel_UpdateAndView(t *testing.T) {
	spin := spinner.New()
	m := progressModel{
		spin:  spin,
		phase: phaseFiles,
		steps: []string{"s1", "s2"},
		done:  map[string]string{},
	}

	v := m.View()
	if !strings.Contains(v, "Generating files...") {
		t.Fatalf("expected files phase view, got %q", v)
	}

	// Move to steps phase.
	m2, _ := m.Update(progressMsgFilesDone{created: []string{"a"}, skipped: []string{"b"}})
	pm := m2.(progressModel)
	if pm.phase != phaseSteps {
		t.Fatalf("expected phaseSteps, got %v", pm.phase)
	}

	// Step start + skip/success/failure messages.
	m3, _ := pm.Update(progressMsgStepStart{name: "s1"})
	pm = m3.(progressModel)
	if pm.status == "" {
		t.Fatalf("expected status to be set")
	}

	m4, _ := pm.Update(progressMsgStepSkipped{name: "s2"})
	pm = m4.(progressModel)
	if pm.done["s2"] != "skipped" {
		t.Fatalf("expected s2 skipped, got %q", pm.done["s2"])
	}

	m5, _ := pm.Update(progressMsgStepSuccess{name: "s1"})
	pm = m5.(progressModel)
	if pm.done["s1"] != "ok" {
		t.Fatalf("expected s1 ok, got %q", pm.done["s1"])
	}

	vs := pm.View()
	if !strings.Contains(vs, "✓ s1") {
		t.Fatalf("expected success marker for s1, got %q", vs)
	}
	if !strings.Contains(vs, " (skipped)") {
		t.Fatalf("expected skipped marker for s2, got %q", vs)
	}

	// Done phase without failure.
	res := &template.ScaffoldResult{}
	m6, _ := pm.Update(progressMsgDone{result: res, err: nil})
	pm = m6.(progressModel)
	if pm.phase != phaseDone {
		t.Fatalf("expected phaseDone, got %v", pm.phase)
	}
	if pm.result != res {
		t.Fatalf("expected result to be set")
	}
}

func TestTeaStepObserver_SafeSendAndHooks(t *testing.T) {
	var got []tea.Msg
	o := teaStepObserver{send: func(msg tea.Msg) { got = append(got, msg) }}

	o.OnStepStart(dsl.Step{Name: "one"}, "")
	o.OnStepSkipped(dsl.Step{Name: "two"})
	o.OnStepSuccess(dsl.Step{Name: "three"}, "")
	o.OnStepFailure(dsl.Step{Name: "four"}, fmt.Errorf("boom"), "")

	if len(got) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(got))
	}
}

func TestProgressHelpers_failedSteps_and_dirExistsAndEmpty(t *testing.T) {
	okErr := fmt.Errorf("x")
	in := []template.StepResult{
		{Name: "a", Error: nil},
		{Name: "b", Error: okErr},
		{Name: "c", Error: nil},
	}
	failed := failedSteps(in)
	if len(failed) != 1 || failed[0].Name != "b" {
		t.Fatalf("unexpected failedSteps output: %#v", failed)
	}

	emptyDir := t.TempDir()
	exists, empty, err := dirExistsAndEmpty(emptyDir)
	if err != nil {
		t.Fatalf("dirExistsAndEmpty returned error: %v", err)
	}
	if !exists || !empty {
		t.Fatalf("expected exists=true empty=true, got exists=%v empty=%v", exists, empty)
	}

	nonEmptyDir := filepath.Join(t.TempDir(), "nonempty")
	if err := os.MkdirAll(nonEmptyDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nonEmptyDir, "file.txt"), []byte("x"), 0o600); err != nil {
		t.Fatalf("writefile: %v", err)
	}

	exists, empty, err = dirExistsAndEmpty(nonEmptyDir)
	if err != nil {
		t.Fatalf("dirExistsAndEmpty returned error: %v", err)
	}
	if !exists || empty {
		t.Fatalf("expected exists=true empty=false, got exists=%v empty=%v", exists, empty)
	}
}

