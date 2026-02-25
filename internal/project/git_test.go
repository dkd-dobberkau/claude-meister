package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupTestRepo creates a temporary git repo for testing.
func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("setup %v failed: %s %v", args, out, err)
		}
	}

	// Create initial commit
	f := filepath.Join(dir, "README.md")
	os.WriteFile(f, []byte("# test"), 0644)
	c := exec.Command("git", "add", ".")
	c.Dir = dir
	c.Run()
	c = exec.Command("git", "commit", "-m", "initial")
	c.Dir = dir
	c.Run()

	return dir
}

func TestGitStatus_Clean(t *testing.T) {
	dir := setupTestRepo(t)
	status, err := GitStatus(dir)
	if err != nil {
		t.Fatalf("GitStatus failed: %v", err)
	}
	if status.Dirty {
		t.Error("expected clean repo")
	}
	if len(status.ModifiedFiles) != 0 {
		t.Errorf("expected 0 modified files, got %d", len(status.ModifiedFiles))
	}
	if len(status.UntrackedFiles) != 0 {
		t.Errorf("expected 0 untracked files, got %d", len(status.UntrackedFiles))
	}
}

func TestGitStatus_Dirty(t *testing.T) {
	dir := setupTestRepo(t)

	// Create untracked file
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0644)
	// Modify tracked file
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# changed"), 0644)

	status, err := GitStatus(dir)
	if err != nil {
		t.Fatalf("GitStatus failed: %v", err)
	}
	if !status.Dirty {
		t.Error("expected dirty repo")
	}
	if len(status.UntrackedFiles) != 1 {
		t.Errorf("expected 1 untracked file, got %d", len(status.UntrackedFiles))
	}
	if len(status.ModifiedFiles) != 1 {
		t.Errorf("expected 1 modified file, got %d", len(status.ModifiedFiles))
	}
}

func TestGitStatus_Branch(t *testing.T) {
	dir := setupTestRepo(t)
	status, err := GitStatus(dir)
	if err != nil {
		t.Fatalf("GitStatus failed: %v", err)
	}
	if status.Branch == "" {
		t.Error("expected non-empty branch name")
	}
}
