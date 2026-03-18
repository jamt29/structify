package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds global configuration for Structify.
type Config struct {
	ConfigDir          string
	TemplatesDir       string
	ConfigFile         string
	DefaultTemplateRepo string
	LogLevel           string
	NonInteractive     bool
}

// Load reads configuration from ~/.structify/config.yaml (or an override set in viper)
// and ensures the required directories exist. It returns a populated Config value.
func Load() (Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("determining user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".structify")
	templatesDir := filepath.Join(configDir, "templates")

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return Config{}, fmt.Errorf("creating config dir: %w", err)
	}

	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		return Config{}, fmt.Errorf("creating templates dir: %w", err)
	}

	v := viper.New()

	// Allow external code to override the config file path via viper if needed.
	if v.ConfigFileUsed() == "" {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(configDir)
	}

	// Defaults
	v.SetDefault("defaultTemplateRepo", "github.com/jamt29/structify-templates")
	v.SetDefault("logLevel", "info")
	v.SetDefault("nonInteractive", false)

	// Read configuration file if present; ignore if it does not exist.
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, fmt.Errorf("reading config file: %w", err)
		}
	}

	cfg := Config{
		ConfigDir:          configDir,
		TemplatesDir:       templatesDir,
		ConfigFile:         v.ConfigFileUsed(),
		DefaultTemplateRepo: v.GetString("defaultTemplateRepo"),
		LogLevel:           v.GetString("logLevel"),
		NonInteractive:     v.GetBool("nonInteractive"),
	}

	return cfg, nil
}


