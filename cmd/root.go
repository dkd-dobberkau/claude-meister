package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "claude-meister",
	Short: "Clean up forgotten Claude Code projects",
	Long:  "claude-meister uses squirrel to find forgotten projects and helps you clean them up: commit/discard git changes, stop Docker containers, archive or delete projects.",
}

func Execute() error {
	return rootCmd.Execute()
}
