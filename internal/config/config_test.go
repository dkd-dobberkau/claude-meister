package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.SquirrelDays != 30 {
		t.Errorf("expected days=30, got %d", cfg.SquirrelDays)
	}
	if cfg.SquirrelDepth != "deep" {
		t.Errorf("expected depth=deep, got %q", cfg.SquirrelDepth)
	}
	if cfg.ArchivePath == "" {
		t.Error("expected non-empty archive path")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("missing config file should return defaults, got error: %v", err)
	}
	if cfg.SquirrelDays != 30 {
		t.Errorf("expected default days=30, got %d", cfg.SquirrelDays)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := []byte(`archive_path: /tmp/my-archive
squirrel_days: 7
squirrel_depth: quick
ignore_paths:
  - /Users/test/important
`)
	if err := os.WriteFile(cfgPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.ArchivePath != "/tmp/my-archive" {
		t.Errorf("expected archive path /tmp/my-archive, got %q", cfg.ArchivePath)
	}
	if cfg.SquirrelDays != 7 {
		t.Errorf("expected days=7, got %d", cfg.SquirrelDays)
	}
	if len(cfg.IgnorePaths) != 1 || cfg.IgnorePaths[0] != "/Users/test/important" {
		t.Errorf("unexpected ignore paths: %v", cfg.IgnorePaths)
	}
}
