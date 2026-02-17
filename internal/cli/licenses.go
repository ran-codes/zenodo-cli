package cli

import (
	"os"

	"github.com/ran-codes/zenodo-cli/internal/api"
	"github.com/ran-codes/zenodo-cli/internal/output"
	"github.com/spf13/cobra"
)

var licensesCmd = &cobra.Command{
	Use:   "licenses",
	Short: "Search available licenses",
}

var licensesSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search available licenses",
	Long: `Search Zenodo's available licenses.

Examples:
  zenodo licenses search
  zenodo licenses search "creative commons"
  zenodo licenses search "MIT" --output json`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(appCtx.BaseURL, appCtx.Token)

		q := ""
		if len(args) > 0 {
			q = args[0]
		}

		result, err := client.SearchLicenses(q, 0, 0)
		if err != nil {
			return err
		}
		return output.Format(os.Stdout, result.Hits.Hits, appCtx.Output, appCtx.Fields)
	},
}

func init() {
	licensesCmd.AddCommand(licensesSearchCmd)
	rootCmd.AddCommand(licensesCmd)
}
