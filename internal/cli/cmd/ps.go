package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/urizennnn/boxify/pkg/container"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List all running containers",
	Long: `Display a list of all running Boxify containers with their details.

Shows container ID, image, command, creation time, status, and other information
in a Docker-like table format.`,
	Example: `  # List all running containers
  boxify ps`,
	Run: func(cmd *cobra.Command, args []string) {
		containers := container.ListAllContainers()

		if len(containers) == 0 {
			fmt.Println("No containers running")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

		fmt.Fprintln(w, "CONTAINER ID\tIMAGE\tCOMMAND\tCREATED\tSTATUS\tPORTS\tNAMES")

		for _, c := range containers {
			containerID := truncateString(c.ID, 12)
			image := c.Image
			if image == "" {
				image = "<none>"
			}

			command := formatCommand(c.Command)
			created := formatTimeSince(c.CreatedAt)
			status := c.Status

			ports := ""

			name := containerID

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				containerID,
				image,
				command,
				created,
				status,
				ports,
				name,
			)
		}

		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func formatCommand(cmd []string) string {
	if len(cmd) == 0 {
		return ""
	}

	cmdStr := strings.Join(cmd, " ")

	maxLen := 20
	if len(cmdStr) > maxLen {
		return fmt.Sprintf("\"%s...\"", cmdStr[:maxLen-3])
	}

	return fmt.Sprintf("\"%s\"", cmdStr)
}

func formatTimeSince(t time.Time) string {
	if t.IsZero() {
		return "Unknown"
	}

	duration := time.Since(t)

	if duration < time.Minute {
		return fmt.Sprintf("%d seconds ago", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	} else {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	}
}
