package cli

import (
	"fmt"
	"os"

	"github.com/ran-codes/zenodo-cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value. Keys use dotted notation.

When the key is "token", the value is stored in the OS keyring for the
active profile. If the keyring is unavailable, it falls back to the config file.

Examples:
  zenodo config set token <your-api-token>
  zenodo config set default_profile sandbox
  zenodo config set profiles.production.base_url https://zenodo.org/api`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Special handling: store tokens in keyring.
		if key == "token" {
			profile, _ := cmd.Flags().GetString("profile")
			profile = config.ResolveProfile(profile)
			if profile == "" {
				profile = cfg.DefaultProfile()
			}

			kr := config.NewKeyring()
			if kr.Available() {
				if err := kr.SetToken(profile, value); err != nil {
					return fmt.Errorf("storing token in keyring: %w", err)
				}
				fmt.Fprintf(os.Stderr, "Token stored in OS keyring for profile %q\n", profile)
				return nil
			}

			// Fallback: store in config file with warning.
			fmt.Fprintf(os.Stderr, "Warning: OS keyring not available, storing token in config file\n")
			cfg.SetProfileValue(profile, "token", value)
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Token stored in config file for profile %q\n", profile)
			return nil
		}

		cfg.Set(key, value)

		if err := cfg.Save(); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Set %s = %s\n", key, value)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Get a configuration value. Keys use dotted notation.

Examples:
  zenodo config get default_profile
  zenodo config get profiles.production.base_url`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		val := cfg.Get(key)
		if val == nil {
			return fmt.Errorf("key %q not found", key)
		}

		fmt.Println(val)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	rootCmd.AddCommand(configCmd)
}
