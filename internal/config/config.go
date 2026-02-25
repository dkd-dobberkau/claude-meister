package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration values.
type Config struct {
	ArchivePath   string   `yaml:"archive_path"`
	SquirrelDays  int      `yaml:"squirrel_days"`
	SquirrelDepth string   `yaml:"squirrel_depth"`
	IgnorePaths   []string `yaml:"ignore_paths"`
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		ArchivePath:   filepath.Join(home, "Archive", "projects"),
		SquirrelDays:  30,
		SquirrelDepth: "deep",
		IgnorePaths:   nil,
	}
}

// ConfigPath returns the default config file path.
func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "claude-meister", "config.yaml")
}

// Load reads config from a YAML file. Returns defaults if file doesn't exist.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
