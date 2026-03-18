package tui

import (
	"testing"

	"github.com/jamt29/structify/internal/dsl"
)

func TestEvalWhen_EmptyAlwaysTrue(t *testing.T) {
	ok, err := evalWhen("", dsl.Context{"a": "x"})
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if !ok {
		t.Fatalf("expected true")
	}
}

func TestEvalWhen_UsesContext(t *testing.T) {
	ok, err := evalWhen(`transport == "http"`, dsl.Context{"transport": "http"})
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if !ok {
		t.Fatalf("expected true")
	}
}

func TestDefaultOrZero_String(t *testing.T) {
	in := dsl.Input{ID: "name", Type: "string"}
	if v := defaultOrZero(in); v != "" {
		t.Fatalf("expected empty string, got %#v", v)
	}
	in.Default = "abc"
	if v := defaultOrZero(in); v != "abc" {
		t.Fatalf("expected default, got %#v", v)
	}
}

func TestDefaultOrZero_Bool(t *testing.T) {
	in := dsl.Input{ID: "b", Type: "bool"}
	if v := defaultOrZero(in); v != false {
		t.Fatalf("expected false, got %#v", v)
	}
	in.Default = true
	if v := defaultOrZero(in); v != true {
		t.Fatalf("expected true, got %#v", v)
	}
}

func TestValidateProvided_StringRegex(t *testing.T) {
	in := dsl.Input{ID: "project_name", Type: "string", Required: true, Validate: "^[a-z]+$"}
	if err := validateProvided(in, "abc"); err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
	if err := validateProvided(in, "Abc"); err == nil {
		t.Fatalf("expected regex error")
	}
}

func TestValidateProvided_Enum(t *testing.T) {
	in := dsl.Input{ID: "orm", Type: "enum", Required: true, Options: []string{"gorm", "sqlx"}}
	if err := validateProvided(in, "gorm"); err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
	if err := validateProvided(in, "none"); err == nil {
		t.Fatalf("expected invalid option error")
	}
}

