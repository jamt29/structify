package config

import "testing"

func TestLoad_ReturnsDirs(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cfg.ConfigDir == "" || cfg.TemplatesDir == "" {
		t.Fatalf("expected non-empty dirs: %#v", cfg)
	}
}

