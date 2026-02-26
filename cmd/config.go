package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dkd-dobberkau/claude-meister/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or initialize configuration",
	RunE:  runConfig,
}

var configInitFlag bool

func init() {
	configCmd.Flags().BoolVar(&configInitFlag, "init", false, "Create default config file")
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfgPath := config.ConfigPath()

	if configInitFlag {
		if _, err := os.Stat(cfgPath); err == nil {
			return fmt.Errorf("config already exists at %s", cfgPath)
		}
		if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
			return err
		}
		cfg := config.DefaultConfig()
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		if err := os.WriteFile(cfgPath, data, 0644); err != nil {
			return err
		}
		fmt.Printf("Config created at %s\n", cfgPath)
		return nil
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	fmt.Printf("Config file: %s\n\n", cfgPath)
	data, _ := yaml.Marshal(cfg)
	fmt.Println(string(data))
	return nil
}
