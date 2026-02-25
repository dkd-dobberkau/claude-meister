// cmd/clean.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/olivercodes/claude-meister/internal/config"
	"github.com/olivercodes/claude-meister/internal/project"
	"github.com/olivercodes/claude-meister/internal/squirrel"
	"github.com/spf13/cobra"
)

var cleanDryRun bool

var cleanCmd = &cobra.Command{
	Use:   "clean <project-name>",
	Short: "Clean up git state of a project",
	Long:  "Shows the git status of a project and offers options to commit, stash, or discard changes.",
	Args:  cobra.ExactArgs(1),
	RunE:  runClean,
}

func init() {
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "Show what would be done without doing it")
	rootCmd.AddCommand(cleanCmd)
}

func runClean(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	runner := squirrel.NewRunner(cfg.SquirrelDays, cfg.SquirrelDepth)
	status, err := runner.Run()
	if err != nil {
		return fmt.Errorf("running squirrel: %w", err)
	}

	p := status.FindByName(projectName)
	if p == nil {
		return fmt.Errorf("project %q not found in squirrel results", projectName)
	}

	gitStatus, err := project.GitStatus(p.Path)
	if err != nil {
		return fmt.Errorf("getting git status: %w", err)
	}

	// Display status
	fmt.Printf("\n  Project:  %s\n", p.ShortName)
	fmt.Printf("  Path:     %s\n", p.Path)
	fmt.Printf("  Branch:   %s\n", gitStatus.Branch)

	if !gitStatus.Dirty {
		fmt.Println("\n  Status: clean - nothing to do!")
		return nil
	}

	fmt.Printf("\n  Modified:  %d files\n", len(gitStatus.ModifiedFiles))
	for _, f := range gitStatus.ModifiedFiles {
		fmt.Printf("    M  %s\n", f)
	}
	fmt.Printf("  Untracked: %d files\n", len(gitStatus.UntrackedFiles))
	for _, f := range gitStatus.UntrackedFiles {
		fmt.Printf("    ?  %s\n", f)
	}
	fmt.Printf("  Staged:    %d files\n", len(gitStatus.StagedFiles))
	for _, f := range gitStatus.StagedFiles {
		fmt.Printf("    A  %s\n", f)
	}
	if gitStatus.StashCount > 0 {
		fmt.Printf("  Stashes:   %d\n", gitStatus.StashCount)
	}

	// Ask what to do
	fmt.Println("\n  What do you want to do?")
	fmt.Println("  [1] Commit all changes")
	fmt.Println("  [2] Stash changes")
	fmt.Println("  [3] Discard all changes")
	fmt.Println("  [4] Skip (do nothing)")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\n  Choice: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if cleanDryRun {
		fmt.Println("\n  [dry-run] Would execute the selected action")
		return nil
	}

	switch choice {
	case "1":
		fmt.Print("  Commit message: ")
		msg, _ := reader.ReadString('\n')
		msg = strings.TrimSpace(msg)
		if msg == "" {
			msg = fmt.Sprintf("chore: claude-meister cleanup of %s", p.ShortName)
		}
		if err := project.GitCommitAll(p.Path, msg); err != nil {
			return fmt.Errorf("committing: %w", err)
		}
		fmt.Println("  Done! Changes committed.")

	case "2":
		if err := project.GitStash(p.Path); err != nil {
			return fmt.Errorf("stashing: %w", err)
		}
		fmt.Println("  Done! Changes stashed.")

	case "3":
		fmt.Printf("  Are you sure? This will discard ALL changes in %s [y/N]: ", p.ShortName)
		confirm, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(confirm)) != "y" {
			fmt.Println("  Aborted.")
			return nil
		}
		if err := project.GitDiscardAll(p.Path); err != nil {
			return fmt.Errorf("discarding: %w", err)
		}
		fmt.Println("  Done! All changes discarded.")

	case "4":
		fmt.Println("  Skipped.")

	default:
		fmt.Println("  Invalid choice, skipped.")
	}

	return nil
}
