package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/olivercodes/claude-meister/internal/config"
	"github.com/olivercodes/claude-meister/internal/squirrel"
	"github.com/olivercodes/claude-meister/internal/tui"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Interactive TUI overview of all projects",
	Long:  "Opens an interactive terminal UI showing all projects from squirrel, organized by category. Navigate with arrow keys, press Enter for details.",
	RunE:  runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	runner := squirrel.NewRunner(cfg.SquirrelDays, cfg.SquirrelDepth)
	fmt.Fprintln(os.Stderr, "Scanning projects with squirrel...")
	status, err := runner.Run()
	if err != nil {
		return fmt.Errorf("running squirrel: %w", err)
	}

	model := tui.NewModel(status)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
