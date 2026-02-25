package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectDocker_ComposeFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte("services:"), 0644)

	info := DetectDocker(dir)
	if !info.HasCompose {
		t.Error("expected HasCompose to be true")
	}
}

func TestDetectDocker_DDEV(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".ddev"), 0755)

	info := DetectDocker(dir)
	if !info.HasDDEV {
		t.Error("expected HasDDEV to be true")
	}
}

func TestDetectDocker_None(t *testing.T) {
	dir := t.TempDir()
	info := DetectDocker(dir)
	if info.HasCompose || info.HasDDEV {
		t.Error("expected no docker detection in empty dir")
	}
}
