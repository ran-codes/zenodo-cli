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
	Use:   "list",
	Short: "List your communities",
	Long: `List the authenticated user's communities, or search all public communities with --query.

Examples:
  zenodo communities list
  zenodo communities list --query "open science"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(appCtx.BaseURL, appCtx.Token)
		query, _ := cmd.Flags().GetString("query")

		var result *model.CommunitySearchResult
		var err error
		if query != "" {
			result, err = client.SearchCommunities(query, 0, 0)
		} else {
			result, err = client.ListUserCommunities("", 0, 0)
		}
		if err != nil {
			return err
		}
		fields := appCtx.Fields
		if fields == "" {
			fields = "metadata.title,links.self_html"
		}
		return output.Format(os.Stdout, result.Hits.Hits, appCtx.Output, fields)
	},
}

func init() {
	communitiesListCmd.Flags().String("query", "", "Search all public communities")
	communitiesCmd.AddCommand(communitiesListCmd)
	rootCmd.AddCommand(communitiesCmd)
}
