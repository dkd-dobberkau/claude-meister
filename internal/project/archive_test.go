package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveProject(t *testing.T) {
	// Create source project
	src := t.TempDir()
	os.WriteFile(filepath.Join(src, "main.go"), []byte("package main"), 0644)
	os.MkdirAll(filepath.Join(src, "sub"), 0755)
	os.WriteFile(filepath.Join(src, "sub", "file.go"), []byte("package sub"), 0644)

	archiveBase := t.TempDir()

	dest, err := ArchiveProject(src, archiveBase)
	if err != nil {
		t.Fatalf("ArchiveProject failed: %v", err)
	}

	// Source should be gone
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("source directory should be removed after archive")
	}

	// Dest should contain the files
	if _, err := os.Stat(filepath.Join(dest, "main.go")); err != nil {
		t.Error("main.go should exist in archive")
	}
	if _, err := os.Stat(filepath.Join(dest, "sub", "file.go")); err != nil {
		t.Error("sub/file.go should exist in archive")
	}
}
