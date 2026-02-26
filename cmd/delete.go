// cmd/delete.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dkd-dobberkau/claude-meister/internal/config"
	"github.com/dkd-dobberkau/claude-meister/internal/project"
	"github.com/dkd-dobberkau/claude-meister/internal/squirrel"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <project-name>",
	Short: "Permanently delete a project",
	Long:  "Permanently deletes a project directory. Requires typing the project name to confirm. Offers to archive first.",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	runner := squirrel.NewRunner(cfg.SquirrelDays, cfg.SquirrelDepth)
	status, err := runner.Run()
	if err != nil {
		return err
	}

	p := status.FindByName(projectName)
	if p == nil {
		return fmt.Errorf("project %q not found", projectName)
	}

	fmt.Printf("\n  Project:  %s\n", p.ShortName)
	fmt.Printf("  Path:     %s\n", p.Path)

	if p.GitDirty {
		fmt.Printf("  WARNING: %d uncommitted files!\n", p.UncommittedFiles)
	}

	reader := bufio.NewReader(os.Stdin)

	// Offer archive first
	fmt.Printf("\n  Would you like to archive instead of delete? [Y/n]: ")
	archiveChoice, _ := reader.ReadString('\n')
	archiveChoice = strings.TrimSpace(strings.ToLower(archiveChoice))
	if archiveChoice == "" || archiveChoice == "y" {
		dest, err := project.ArchiveProject(p.Path, cfg.ArchivePath)
		if err != nil {
			return fmt.Errorf("archiving: %w", err)
		}
		fmt.Printf("  Archived to: %s\n", dest)
		return nil
	}

	// Double confirmation for delete
	fmt.Printf("\n  This will PERMANENTLY delete %s\n", p.Path)
	fmt.Printf("  Type the project name to confirm: ")
	confirmation, _ := reader.ReadString('\n')
	if strings.TrimSpace(confirmation) != p.ShortName {
		fmt.Println("  Names don't match. Aborted.")
		return nil
	}

	if err := project.DeleteProject(p.Path); err != nil {
		return fmt.Errorf("deleting: %w", err)
	}

	fmt.Println("  Deleted.")
	return nil
}
