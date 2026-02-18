package cli

import (
	"os"

	"github.com/ran-codes/zenodo-cli/internal/api"
	"github.com/ran-codes/zenodo-cli/internal/model"
	"github.com/ran-codes/zenodo-cli/internal/output"
	"github.com/spf13/cobra"
)

var communitiesCmd = &cobra.Command{
	Use:   "communities",
	Short: "Search and list communities",
}

var communitiesListCmd = &cobra.Command{
	Use:   "list [query]",
	Short: "List your communities",
	Long: `List the authenticated user's communities. Use --all to search all communities.

Examples:
  zenodo communities list
  zenodo communities list --all
  zenodo communities list --all "open science"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(appCtx.BaseURL, appCtx.Token)
		all, _ := cmd.Flags().GetBool("all")

		q := ""
		if len(args) > 0 {
			q = args[0]
		}

		var result *model.CommunitySearchResult
		var err error
		if all {
			result, err = client.SearchCommunities(q, 0, 0)
		} else {
			result, err = client.ListUserCommunities(q, 0, 0)
		}
		if err != nil {
			return err
		}
		return output.Format(os.Stdout, result.Hits.Hits, appCtx.Output, appCtx.Fields)
	},
}

func init() {
	communitiesListCmd.Flags().Bool("all", false, "Search all communities instead of just yours")
	communitiesCmd.AddCommand(communitiesListCmd)
	rootCmd.AddCommand(communitiesCmd)
}
