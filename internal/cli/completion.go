package cli

import (
	"os"

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
    # To load completions for every new session, add the output to your profile.`,
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

func init() {
	rootCmd.AddCommand(completionCmd)
}
