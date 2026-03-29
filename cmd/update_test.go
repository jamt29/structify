package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/jamt29/structify/internal/buildinfo"
	"github.com/spf13/cobra"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want int
	}{
		{"0.5.0", "0.5.1", -1},
		{"0.5.1", "0.5.1", 0},
		{"0.6.0", "0.5.9", 1},
	}

	for _, tt := range tests {
		got := compareVersions(tt.a, tt.b)
		if got != tt.want {
			t.Fatalf("compareVersions(%q,%q)=%d want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{name: "newer patch", current: "0.5.1", latest: "0.5.2", want: true},
		{name: "newer minor", current: "0.5.1", latest: "0.6.0", want: true},
		{name: "equal", current: "0.6.1", latest: "0.6.1", want: false},
		{name: "current higher", current: "0.7.0", latest: "0.6.1", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNewer(tt.current, tt.latest)
			if got != tt.want {
				t.Fatalf("isNewer(%q,%q)=%v want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestRunUpdate_CheckOnly(t *testing.T) {
	origFetch := fetchLatestReleaseFn
	origRun := runSelfUpdateFn
	origVer := buildinfo.Version
	defer func() {
		fetchLatestReleaseFn = origFetch
		runSelfUpdateFn = origRun
		buildinfo.Version = origVer
	}()

	buildinfo.Version = "0.5.0"
	fetchLatestReleaseFn = func() (*githubRelease, error) { return &githubRelease{TagName: "v0.5.1"}, nil }
	ranInstall := false
	runSelfUpdateFn = func() error {
		ranInstall = true
		return nil
	}

	updateCheck, updateYes = true, false
	var out bytes.Buffer
	cmd := &cobraHarness{out: &out, in: strings.NewReader("")}

	if err := runUpdate(cmd.command(), nil); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ranInstall {
		t.Fatalf("install should not run in --check mode")
	}
	if !strings.Contains(out.String(), "Hay una nueva version disponible.") {
		t.Fatalf("expected check message, got: %s", out.String())
	}
}

func TestRunUpdate_YesPerformsInstall(t *testing.T) {
	origFetch := fetchLatestReleaseFn
	origRun := runSelfUpdateFn
	origVer := buildinfo.Version
	defer func() {
		fetchLatestReleaseFn = origFetch
		runSelfUpdateFn = origRun
		buildinfo.Version = origVer
	}()

	buildinfo.Version = "0.5.0"
	fetchLatestReleaseFn = func() (*githubRelease, error) { return &githubRelease{TagName: "v0.5.1"}, nil }
	ranInstall := false
	runSelfUpdateFn = func() error {
		ranInstall = true
		return nil
	}

	updateCheck, updateYes = false, true
	var out bytes.Buffer
	cmd := &cobraHarness{out: &out, in: strings.NewReader("")}

	if err := runUpdate(cmd.command(), nil); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !ranInstall {
		t.Fatalf("expected install to run")
	}
}

func TestRunUpdate_CancelByPrompt(t *testing.T) {
	origFetch := fetchLatestReleaseFn
	origRun := runSelfUpdateFn
	origVer := buildinfo.Version
	defer func() {
		fetchLatestReleaseFn = origFetch
		runSelfUpdateFn = origRun
		buildinfo.Version = origVer
	}()

	buildinfo.Version = "0.5.0"
	fetchLatestReleaseFn = func() (*githubRelease, error) { return &githubRelease{TagName: "v0.5.1"}, nil }
	ranInstall := false
	runSelfUpdateFn = func() error {
		ranInstall = true
		return nil
	}

	updateCheck, updateYes = false, false
	var out bytes.Buffer
	cmd := &cobraHarness{out: &out, in: strings.NewReader("n\n")}

	if err := runUpdate(cmd.command(), nil); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ranInstall {
		t.Fatalf("install should not run when cancelled")
	}
}

func TestRunUpdate_AlreadyUpdated(t *testing.T) {
	origFetch := fetchLatestReleaseFn
	origVer := buildinfo.Version
	defer func() {
		fetchLatestReleaseFn = origFetch
		buildinfo.Version = origVer
	}()

	buildinfo.Version = "0.5.1"
	fetchLatestReleaseFn = func() (*githubRelease, error) { return &githubRelease{TagName: "v0.5.1"}, nil }

	updateCheck, updateYes = false, false
	var out bytes.Buffer
	cmd := &cobraHarness{out: &out, in: strings.NewReader("")}
	if err := runUpdate(cmd.command(), nil); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !strings.Contains(out.String(), "Ya tienes la ultima version") {
		t.Fatalf("expected up-to-date message, got: %s", out.String())
	}
}

func TestRunUpdate_FetchError(t *testing.T) {
	origFetch := fetchLatestReleaseFn
	defer func() { fetchLatestReleaseFn = origFetch }()

	fetchLatestReleaseFn = func() (*githubRelease, error) { return nil, errors.New("network down") }
	updateCheck, updateYes = false, false
	cmd := (&cobraHarness{out: &bytes.Buffer{}, in: strings.NewReader("")}).command()
	if err := runUpdate(cmd, nil); err == nil {
		t.Fatalf("expected error")
	}
}

func TestRunUpdate_GoNotFoundShowsManualInstructions(t *testing.T) {
	origFetch := fetchLatestReleaseFn
	origRun := runSelfUpdateFn
	origVer := buildinfo.Version
	defer func() {
		fetchLatestReleaseFn = origFetch
		runSelfUpdateFn = origRun
		buildinfo.Version = origVer
	}()

	buildinfo.Version = "0.5.0"
	fetchLatestReleaseFn = func() (*githubRelease, error) { return &githubRelease{TagName: "v0.5.1"}, nil }
	runSelfUpdateFn = func() error { return errGoNotFound }
	updateCheck, updateYes = false, true

	var out bytes.Buffer
	cmd := (&cobraHarness{out: &out, in: strings.NewReader("")}).command()
	if err := runUpdate(cmd, nil); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !strings.Contains(out.String(), "No se encontro 'go' en el PATH.") {
		t.Fatalf("expected PATH guidance, got: %s", out.String())
	}
}

type cobraHarness struct {
	out *bytes.Buffer
	in  *strings.Reader
}

func (h *cobraHarness) command() *cobra.Command {
	c := &cobra.Command{}
	c.SetOut(h.out)
	c.SetIn(h.in)
	return c
}

