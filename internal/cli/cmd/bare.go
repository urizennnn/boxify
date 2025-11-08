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
	Short: "Run a bare container with legacy method",
	Long: `Creates and enters a container using the legacy bare method.
This command reads from boxify.yaml/boxify.yml and creates a container
with the specified configuration, then enters it using nsenter.`,
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
