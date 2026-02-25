package project

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitStatusResult holds the parsed output of git status for a repository.
type GitStatusResult struct {
	Branch         string
	Dirty          bool
	ModifiedFiles  []string
	UntrackedFiles []string
	StagedFiles    []string
	StashCount     int
}

// GitStatus returns the git status of a project directory.
func GitStatus(path string) (*GitStatusResult, error) {
	result := &GitStatusResult{}

	// Get branch name
	branch, err := gitCmd(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("getting branch: %w", err)
	}
	result.Branch = strings.TrimSpace(branch)

	// Get porcelain status
	status, err := gitCmd(path, "status", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("getting status: %w", err)
	}

	for _, line := range strings.Split(status, "\n") {
		if line == "" {
			continue
		}
		if len(line) < 3 {
			continue
		}
		xy := line[:2]
		file := strings.TrimSpace(line[2:])

		switch {
		case strings.HasPrefix(xy, "??"):
			result.UntrackedFiles = append(result.UntrackedFiles, file)
		case xy[0] != ' ' && xy[0] != '?':
			result.StagedFiles = append(result.StagedFiles, file)
		case xy[1] != ' ':
			result.ModifiedFiles = append(result.ModifiedFiles, file)
		}
	}

	result.Dirty = len(result.ModifiedFiles) > 0 ||
		len(result.UntrackedFiles) > 0 ||
		len(result.StagedFiles) > 0

	// Get stash count
	stash, _ := gitCmd(path, "stash", "list")
	if stash != "" {
		result.StashCount = len(strings.Split(strings.TrimSpace(stash), "\n"))
	}

	return result, nil
}

// GitCommitAll stages and commits all changes.
func GitCommitAll(path, message string) error {
	if _, err := gitCmd(path, "add", "."); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	if _, err := gitCmd(path, "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// GitStash stashes all changes with an auto-generated message.
func GitStash(path string) error {
	_, err := gitCmd(path, "stash", "push", "-m", "claude-meister auto-stash")
	return err
}

// GitDiscardAll discards all uncommitted changes including untracked files.
func GitDiscardAll(path string) error {
	if _, err := gitCmd(path, "checkout", "."); err != nil {
		return fmt.Errorf("git checkout: %w", err)
	}
	if _, err := gitCmd(path, "clean", "-fd"); err != nil {
		return fmt.Errorf("git clean: %w", err)
	}
	return nil
}

// gitCmd executes a git command in the specified directory and returns its output.
func gitCmd(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, string(out))
	}
	return string(out), nil
}
