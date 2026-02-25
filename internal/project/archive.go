package project

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ArchiveProject moves a project directory to the archive.
// Returns the destination path.
func ArchiveProject(projectPath, archiveBase string) (string, error) {
	now := time.Now()
	monthDir := filepath.Join(archiveBase, now.Format("2006-01"))
	projectName := filepath.Base(projectPath)
	dest := filepath.Join(monthDir, projectName)

	if err := os.MkdirAll(monthDir, 0755); err != nil {
		return "", fmt.Errorf("creating archive dir: %w", err)
	}

	if _, err := os.Stat(dest); err == nil {
		dest = fmt.Sprintf("%s-%s", dest, now.Format("150405"))
	}

	if err := os.Rename(projectPath, dest); err != nil {
		return "", fmt.Errorf("moving project: %w", err)
	}

	return dest, nil
}

// DeleteProject removes a project directory entirely.
func DeleteProject(projectPath string) error {
	return os.RemoveAll(projectPath)
}
