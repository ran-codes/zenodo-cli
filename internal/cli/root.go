package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/ran-codes/zenodo-cli/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zenodo",
	Short: "CLI for the Zenodo REST API",
	Long: `A command-line tool for managing metadata, searching records,
and querying assets on Zenodo.

Get started:
  zenodo config set token <YOUR_TOKEN>   Set your API token
  zenodo records list                    List your records
  zenodo records search "query"          Search published records
  zenodo records get <id>                Get record details

Exit codes: 0=success, 1=API error, 2=auth error, 3=validation error,
            4=rate limit, 5=user cancelled`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Set up logging.
		verbose, _ := cmd.Flags().GetBool("verbose")
		level := slog.LevelWarn
		if verbose {
			level = slog.LevelDebug
		}
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

		// Load config.
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Check config file permissions.
		if warning := config.CheckConfigPermissions(); warning != nil {
			slog.Warn(warning.Error())
		}

		// Resolve profile.
		profileFlag, _ := cmd.Flags().GetString("profile")
		profile := config.ResolveProfile(profileFlag)
		if profile == "" {
			profile = cfg.DefaultProfile()
		}

		// Set up keyring.
		kr := config.NewKeyring()
		if !kr.Available() {
			slog.Debug("OS keyring not available, using config file fallback for tokens")
		}

		// Auto-migrate plaintext tokens from config to keyring.
		kr.MigrateToken(cfg, profile)

		// Resolve token.
		tokenFlag, _ := cmd.Flags().GetString("token")
		if tokenFlag != "" {
			fmt.Fprintln(os.Stderr, "Warning: passing token via --token flag is insecure; prefer ZENODO_TOKEN env var or `zenodo config set token`")
		}
		token := config.ResolveTokenFull(tokenFlag, kr, cfg, profile)

		// Resolve base URL.
		sandbox, _ := cmd.Flags().GetBool("sandbox")
		baseURL := cfg.ResolveBaseURL(profile, sandbox)

		// Resolve output format.
		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
				output = "table"
			} else {
				output = "json"
			}
		}

		fields, _ := cmd.Flags().GetString("fields")

		// Populate shared context.
		appCtx = AppContext{
			Config:  cfg,
			Keyring: kr,
			Profile: profile,
			Token:   token,
			BaseURL: baseURL,
			Output:  output,
			Fields:  fields,
			Verbose: verbose,
		}

		slog.Debug("resolved context",
			"profile", profile,
			"base_url", baseURL,
			"output", output,
			"has_token", token != "",
		)

		return nil
	},
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	rootCmd.PersistentFlags().String("token", "", "API token (prefer ZENODO_TOKEN env var or keyring)")
	rootCmd.PersistentFlags().String("profile", "", "Config profile to use")
	rootCmd.PersistentFlags().Bool("sandbox", false, "Use Zenodo sandbox environment")
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output format: json, table, csv (default: table for TTY, json for pipe)")
	rootCmd.PersistentFlags().String("fields", "", "Comma-separated list of fields to display")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
}

// Execute runs the root command.
func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
	return err
}
