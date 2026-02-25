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

var archiveCmd = &cobra.Command{
	Use:   "archive <project-name>",
	Short: "Archive a project to the archive directory",
	Args:  cobra.ExactArgs(1),
	RunE:  runArchive,
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}

func runArchive(cmd *cobra.Command, args []string) error {
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
	fmt.Printf("  Archive:  %s\n\n", cfg.ArchivePath)

	if p.GitDirty {
		fmt.Printf("  WARNING: Project has %d uncommitted files!\n", p.UncommittedFiles)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("  Move project to archive? [y/N]: ")
	confirm, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(confirm)) != "y" {
		fmt.Println("  Aborted.")
		return nil
	}

	dest, err := project.ArchiveProject(p.Path, cfg.ArchivePath)
	if err != nil {
		return fmt.Errorf("archiving: %w", err)
	}

	fmt.Printf("  Archived to: %s\n", dest)

	ackCmd := fmt.Sprintf("squirrel ack %s", p.ShortName)
	fmt.Printf("  Run '%s' to remove from squirrel list\n", ackCmd)

	return nil
}
