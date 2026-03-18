package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "createon",
	Short: "Createon - Self-hosted creator platform",
	Long: `Createon is a self-hosted alternative to Patreon using cryptocurrency payments.
Complete platform management through command line interface.`,
}

// Execute runs the root command and handles command-line argument parsing.
// It is the main entry point for the CLI application.
// Returns an error if command execution fails.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "config/server.yaml", "config file")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}
