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

var configDeleteCmd = &cobra.Command{
	Use:   "delete <key>",
	Short: "Delete a configuration value",
	Long: `Delete a configuration value. Keys use dotted notation.

When the key is "token", the value is removed from both the OS keyring
and the config file for the active profile.

Examples:
  zenodo config delete orcid
  zenodo config delete token`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Special handling: remove token from keyring and config.
		if key == "token" {
			profile, _ := cmd.Flags().GetString("profile")
			profile = config.ResolveProfile(profile)
			if profile == "" {
				profile = cfg.DefaultProfile()
			}

			kr := config.NewKeyring()
			if kr.Available() {
				if err := kr.DeleteToken(profile); err != nil {
					return fmt.Errorf("removing token from keyring: %w", err)
				}
			}

			// Also remove from config file in case it was stored there.
			cfg.SetProfileValue(profile, "token", "")
			cfg.Delete("profiles." + profile + ".token")
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Token deleted for profile %q\n", profile)
			return nil
		}

		if cfg.Get(key) == nil {
			return fmt.Errorf("key %q not found", key)
		}

		cfg.Delete(key)
		if err := cfg.Save(); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Deleted %s\n", key)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all configuration values",
	Long: `Show all known configuration keys and their current values.

Sensitive values like tokens are masked. Unset keys show MISSING.

Examples:
  zenodo config list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		profile := appCtx.Profile

		// Known config keys to display.
		keys := []struct {
			label string
			value string
		}{
			{"profile", profile},
			{"base_url", cfg.ResolveBaseURL(profile, false)},
			{"token", formatToken(appCtx.Token)},
			{"orcid", formatValue(cfg.Get("orcid"))},
		}

		for _, k := range keys {
			fmt.Printf("%-12s %s\n", k.label+":", k.value)
		}
		return nil
	},
}

func formatToken(token string) string {
	if token == "" {
		return "MISSING"
	}
	return config.MaskToken(token)
}

func formatValue(val interface{}) string {
	if val == nil {
		return "MISSING"
	}
	s := fmt.Sprintf("%v", val)
	if s == "" {
		return "MISSING"
	}
	return s
}

var configProfilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List configured profiles",
	Long: `List all configured profiles and show which is the default.

Examples:
  zenodo config profiles`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		defaultProfile := cfg.DefaultProfile()
		names := cfg.ProfileNames()

		if len(names) == 0 {
			fmt.Fprintln(os.Stderr, "No profiles configured.")
			return nil
		}

		for _, name := range names {
			marker := "  "
			if name == defaultProfile {
				marker = "* "
			}
			baseURL := cfg.ProfileBaseURL(name)
			fmt.Printf("%s%s (%s)\n", marker, name, baseURL)
		}
		return nil
	},
}

var configUseCmd = &cobra.Command{
	Use:   "use <profile>",
	Short: "Switch the default profile",
	Long: `Set a profile as the default.

Examples:
  zenodo config use sandbox
  zenodo config use production`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Verify profile exists.
		found := false
		for _, p := range cfg.ProfileNames() {
			if p == name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("profile %q not found; available profiles: %v", name, cfg.ProfileNames())
		}

		cfg.SetDefaultProfile(name)
		if err := cfg.Save(); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Default profile set to %q\n", name)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configProfilesCmd)
	configCmd.AddCommand(configUseCmd)
	rootCmd.AddCommand(configCmd)
}
