package cmd

import (
	"fmt"

	"github.com/dkd-dobberkau/claude-meister/internal/config"
	"github.com/dkd-dobberkau/claude-meister/internal/project"
	"github.com/dkd-dobberkau/claude-meister/internal/squirrel"
	"github.com/spf13/cobra"
)

var dockerStopCmd = &cobra.Command{
	Use:   "docker-stop",
	Short: "Stop Docker/DDEV containers of all forgotten projects",
	RunE:  runDockerStop,
}

func init() {
	rootCmd.AddCommand(dockerStopCmd)
}

func runDockerStop(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return err
	}

	runner := squirrel.NewRunner(cfg.SquirrelDays, cfg.SquirrelDepth)
	fmt.Println("Scanning projects...")
	status, err := runner.Run()
	if err != nil {
		return err
	}

	allProjects := status.AllProjects()
	found := 0

	for _, p := range allProjects {
		info := project.DetectDocker(p.Path)
		if !info.HasCompose && !info.HasDDEV {
			continue
		}
		found++

		dockerType := "Docker Compose"
		if info.HasDDEV {
			dockerType = "DDEV"
		}
		fmt.Printf("\n  %s (%s)\n", p.ShortName, dockerType)
		fmt.Printf("  Path: %s\n", p.Path)

		result, err := project.StopDocker(p.Path)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			fmt.Printf("  %s\n", result)
		}
	}

	if found == 0 {
		fmt.Println("No projects with Docker/DDEV configuration found.")
	} else {
		fmt.Printf("\nProcessed %d projects with Docker/DDEV.\n", found)
	}
	return nil
}
