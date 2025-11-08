/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/urizennnn/boxify/internal/legacy"
)

// bareCmd represents the bare command
var bareCmd = &cobra.Command{
	Use:   "bare",
	Short: "Run a container using the legacy bare method",
	Long: `Create and enter a container using the legacy bare method.

This command reads configuration from boxify.yaml or boxify.yml in the current
directory and creates a container with the specified settings. After creation,
it automatically enters the container using nsenter, providing an interactive
shell within the isolated environment.

Configuration file should specify:
  • image_name: Container name/identifier
  • memory_limit: Maximum memory (e.g., 100m, 1g)
  • cpu_limit: CPU weight for relative CPU time allocation`,
	Example: `  # Run a container from boxify.yaml configuration
  boxify bare

  # Ensure boxify.yaml exists first
  cp boxify.example.yaml boxify.yaml
  boxify bare`,
	Run: func(cmd *cobra.Command, args []string) {
		legacy.Run()
	},
}

func init() {
	rootCmd.AddCommand(bareCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bareCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bareCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
