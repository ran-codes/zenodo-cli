package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for the zenodo CLI.

To load completions:

  Bash:
    $ source <(zenodo completion bash)
    # To load completions for each session, execute once:
    # Linux:
    $ zenodo completion bash > /etc/bash_completion.d/zenodo
    # macOS:
    $ zenodo completion bash > $(brew --prefix)/etc/bash_completion.d/zenodo

  Zsh:
    $ source <(zenodo completion zsh)
    # To load completions for each session, execute once:
    $ zenodo completion zsh > "${fpath[1]}/_zenodo"

  Fish:
    $ zenodo completion fish | source
    # To load completions for each session, execute once:
    $ zenodo completion fish > ~/.config/fish/completions/zenodo.fish

  PowerShell:
    PS> zenodo completion powershell | Out-String | Invoke-Expression
    # To load completions for every new session, add the output to your profile.

Or use "zenodo completion install" to set this up automatically.`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

var completionInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install shell completions into your shell profile",
	Long: `Automatically configure shell completions by appending the
completion script to your shell profile.

Detects the current shell automatically, or use --shell to specify one.

Examples:
  zenodo completion install
  zenodo completion install --shell powershell`,
	RunE: func(cmd *cobra.Command, args []string) error {
		shellFlag, _ := cmd.Flags().GetString("shell")

		shell := shellFlag
		if shell == "" {
			shell = detectShell()
		}
		if shell == "" {
			return fmt.Errorf("could not detect shell; use --shell to specify one (bash, zsh, fish, powershell)")
		}

		profilePath, sourceLine, err := completionConfig(shell)
		if err != nil {
			return err
		}

		// For fish, we write the completion script directly to a file.
		if shell == "fish" {
			return installFishCompletion(profilePath)
		}

		// Check if already installed.
		if alreadyInstalled(profilePath, sourceLine) {
			fmt.Printf("Completions already installed in %s\n", profilePath)
			return nil
		}

		// Show what we'll do and ask for confirmation.
		fmt.Printf("Shell:   %s\n", shell)
		fmt.Printf("File:    %s\n", profilePath)
		fmt.Printf("Append:  %s\n", sourceLine)
		fmt.Print("\nProceed? [y/N] ")

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}

		// Append to profile.
		f, err := os.OpenFile(profilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("opening %s: %w", profilePath, err)
		}
		defer f.Close()

		if _, err := fmt.Fprintf(f, "\n# zenodo shell completions\n%s\n", sourceLine); err != nil {
			return fmt.Errorf("writing to %s: %w", profilePath, err)
		}

		fmt.Printf("\nDone! Restart your shell or run:\n  source %s\n", profilePath)
		return nil
	},
}

func init() {
	completionInstallCmd.Flags().String("shell", "", "Shell to install for (bash, zsh, fish, powershell)")
	completionCmd.AddCommand(completionInstallCmd)
	rootCmd.AddCommand(completionCmd)
}

// detectShell returns the current shell name, or empty string if unknown.
func detectShell() string {
	// Check SHELL env var (Unix).
	if sh := os.Getenv("SHELL"); sh != "" {
		base := filepath.Base(sh)
		switch {
		case strings.Contains(base, "zsh"):
			return "zsh"
		case strings.Contains(base, "bash"):
			return "bash"
		case strings.Contains(base, "fish"):
			return "fish"
		}
	}

	// On Windows, default to powershell.
	if runtime.GOOS == "windows" {
		return "powershell"
	}

	return ""
}

// completionConfig returns (profilePath, sourceLine, error) for the given shell.
func completionConfig(shell string) (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("finding home directory: %w", err)
	}

	switch shell {
	case "bash":
		return filepath.Join(home, ".bashrc"),
			`eval "$(zenodo completion bash)"`,
			nil
	case "zsh":
		return filepath.Join(home, ".zshrc"),
			`eval "$(zenodo completion zsh)"`,
			nil
	case "fish":
		return filepath.Join(home, ".config", "fish", "completions", "zenodo.fish"),
			"", // fish uses a file, not a source line
			nil
	case "powershell":
		profile := powershellProfile()
		if profile == "" {
			return "", "", fmt.Errorf("could not determine PowerShell profile path; is pwsh or powershell installed?")
		}
		return profile,
			"zenodo completion powershell | Out-String | Invoke-Expression",
			nil
	default:
		return "", "", fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shell)
	}
}

// powershellProfile returns the PowerShell $PROFILE path.
func powershellProfile() string {
	// Try pwsh (PowerShell Core) first, then powershell (Windows PowerShell).
	for _, ps := range []string{"pwsh", "powershell"} {
		out, err := exec.Command(ps, "-NoProfile", "-Command", "echo $PROFILE").Output()
		if err == nil {
			p := strings.TrimSpace(string(out))
			if p != "" {
				return p
			}
		}
	}
	return ""
}

// alreadyInstalled checks if the source line is already in the profile file.
func alreadyInstalled(profilePath, sourceLine string) bool {
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), sourceLine)
}

// installFishCompletion writes the fish completion script directly to the completions dir.
func installFishCompletion(path string) error {
	// Check if already exists.
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("Fish completions already installed at %s\n", path)
		return nil
	}

	fmt.Printf("Write fish completions to %s\n", path)
	fmt.Print("Proceed? [y/N] ")

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	// Ensure directory exists.
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()

	if err := rootCmd.GenFishCompletion(f, true); err != nil {
		return fmt.Errorf("generating fish completions: %w", err)
	}

	fmt.Printf("\nDone! Completions will be loaded automatically in new fish sessions.\n")
	return nil
}
