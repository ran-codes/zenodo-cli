package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zenodo",
	Short: "CLI for the Zenodo REST API",
	Long:  "A command-line tool for managing metadata, searching records, and querying assets on Zenodo.",
	// No Run: root command prints help by default via cobra.
}

func init() {
	rootCmd.PersistentFlags().String("token", "", "API token (prefer ZENODO_TOKEN env var or keyring)")
	rootCmd.PersistentFlags().String("profile", "", "Config profile to use")
	rootCmd.PersistentFlags().Bool("sandbox", false, "Use Zenodo sandbox environment")
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output format: json, table, csv (default: table for TTY, json for pipe)")
	rootCmd.PersistentFlags().String("fields", "", "Comma-separated list of fields to display")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
