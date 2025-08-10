package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Build-time variables set by GoReleaser
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
	builtBy = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Provider Explorer %s\n", version)
		if commit != "unknown" {
			fmt.Printf("Commit: %s\n", commit)
		}
		if date != "unknown" {
			fmt.Printf("Built: %s\n", date)
		}
		if builtBy != "unknown" {
			fmt.Printf("Built by: %s\n", builtBy)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
