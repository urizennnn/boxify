/*
Copyright © 2025 urizennnn <igamerryt@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "boxify",
	Short: "A lightweight container runtime built from scratch",
	Long: `Boxify is a lightweight container runtime built from scratch in Go.
It provides process isolation using Linux namespaces, resource management
with cgroups, and networking capabilities - similar to Docker but simpler.

Features:
  • Process isolation with Linux namespaces (PID, UTS, IPC, NET, MNS)
  • Resource management via cgroups v2 (CPU and memory limits)
  • Networking with virtual ethernet pairs and bridge networking
  • Overlay filesystem for container isolation
  • Background daemon architecture for container lifecycle management

Usage:
  boxify [command]

Examples:
  # List running containers
  boxify ps

  # Run a container using legacy bare method
  boxify bare

  # Display help for a command
  boxify [command] --help`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Configure Cobra to show help by default when no command is provided
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
