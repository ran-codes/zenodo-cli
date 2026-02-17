package cli

import (
	"os"

	"github.com/ran-codes/zenodo-cli/internal/api"
	"github.com/ran-codes/zenodo-cli/internal/output"
	"github.com/spf13/cobra"
)

var communitiesCmd = &cobra.Command{
	Use:   "communities",
	Short: "Search and list communities",
}

var communitiesListCmd = &cobra.Command{
	Use:   "list [query]",
	Short: "List or search communities",
	Long: `List or search Zenodo communities.

Examples:
  zenodo communities list
  zenodo communities list "open science"
  zenodo communities list --output csv`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(appCtx.BaseURL, appCtx.Token)

		q := ""
		if len(args) > 0 {
			q = args[0]
		}

		result, err := client.SearchCommunities(q, 0, 0)
		if err != nil {
			return err
		}
		return output.Format(os.Stdout, result.Hits.Hits, appCtx.Output, appCtx.Fields)
	},
}

func init() {
	communitiesCmd.AddCommand(communitiesListCmd)
	rootCmd.AddCommand(communitiesCmd)
}
