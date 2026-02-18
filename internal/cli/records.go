package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ran-codes/zenodo-cli/internal/api"
	"github.com/ran-codes/zenodo-cli/internal/model"
	"github.com/ran-codes/zenodo-cli/internal/output"
	"github.com/spf13/cobra"
)

var recordsCmd = &cobra.Command{
	Use:   "records",
	Short: "List, search, and view records",
}

var recordsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your records and drafts",
	Long: `List the authenticated user's records and drafts.

Examples:
  zenodo records list
  zenodo records list --status draft
  zenodo records list --community my-org --all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(appCtx.BaseURL, appCtx.Token)
		status, _ := cmd.Flags().GetString("status")
		community, _ := cmd.Flags().GetString("community")
		all, _ := cmd.Flags().GetBool("all")

		params := api.RecordListParams{
			Status:    status,
			Community: community,
		}

		fields := appCtx.Fields
		if fields == "" {
			fields = "title,metadata.communities,doi_url,created"
		}

		if all {
			var allDepositions []model.Deposition
			const pageSize = 100
			for page := 1; ; page++ {
				params.Page = page
				deps, err := client.ListUserRecords(params)
				if err != nil {
					return err
				}
				allDepositions = append(allDepositions, deps...)
				if len(deps) < pageSize {
					break
				}
			}
			fmt.Fprintf(os.Stderr, "Total: %d records\n", len(allDepositions))
			return output.Format(os.Stdout, allDepositions, appCtx.Output, fields)
		}

		depositions, err := client.ListUserRecords(params)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Showing %d records\n", len(depositions))
		return output.Format(os.Stdout, depositions, appCtx.Output, fields)
	},
}

var recordsSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search published records",
	Long: `Search all published records using Elasticsearch query syntax.

Examples:
  zenodo records search "climate change"
  zenodo records search --community my-org "dataset"
  zenodo records search "publication_date:[2024-01-01 TO 2024-12-31]" --all`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(appCtx.BaseURL, appCtx.Token)
		query := args[0]
		community, _ := cmd.Flags().GetString("community")
		all, _ := cmd.Flags().GetBool("all")

		params := api.RecordListParams{
			Community: community,
		}

		searchFields := appCtx.Fields
		if searchFields == "" {
			searchFields = "id,title,doi,links.html,created"
		}

		if all {
			records, total, err := api.PaginateAll(func(page int) (*model.RecordSearchResult, error) {
				params.Page = page
				return client.SearchRecords(query, params)
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			}
			fmt.Fprintf(os.Stderr, "Total: %d records\n", total)
			return output.Format(os.Stdout, records, appCtx.Output, searchFields)
		}

		result, err := client.SearchRecords(query, params)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Showing %d of %d records\n", len(result.Hits.Hits), result.Hits.Total)
		return output.Format(os.Stdout, result.Hits.Hits, appCtx.Output, searchFields)
	},
}

var recordsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a record by ID",
	Long: `Retrieve full details of a published record.

Examples:
  zenodo records get 12345
  zenodo records get 12345 --output json
  zenodo records get 12345 --format bibtex`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid record ID: %s", args[0])
		}

		client := api.NewClient(appCtx.BaseURL, appCtx.Token)
		format, _ := cmd.Flags().GetString("format")

		// Handle non-JSON formats via Accept header.
		switch format {
		case "bibtex":
			data, err := client.GetRaw(fmt.Sprintf("/records/%d", id), "application/x-bibtex")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		case "datacite":
			data, err := client.GetRaw(fmt.Sprintf("/records/%d", id), "application/vnd.datacite.datacite+xml")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		record, err := client.GetRecord(id)
		if err != nil {
			return err
		}
		return output.Format(os.Stdout, record, appCtx.Output, appCtx.Fields)
	},
}

var recordsVersionsCmd = &cobra.Command{
	Use:   "versions <id>",
	Short: "List all versions of a record",
	Long: `List all versions of a record by its ID.

Examples:
  zenodo records versions 12345`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid record ID: %s", args[0])
		}

		client := api.NewClient(appCtx.BaseURL, appCtx.Token)
		result, err := client.ListVersions(id)
		if err != nil {
			return err
		}
		fields := appCtx.Fields
		if fields == "" {
			fields = "id,title,doi,links.html,created"
		}
		return output.Format(os.Stdout, result.Hits.Hits, appCtx.Output, fields)
	},
}

func init() {
	// records list flags
	recordsListCmd.Flags().String("status", "", "Filter by status: draft, published")
	recordsListCmd.Flags().String("community", "", "Filter by community ID")
	recordsListCmd.Flags().Bool("all", false, "Fetch all pages (up to 10k results)")

	// records search flags
	recordsSearchCmd.Flags().String("community", "", "Filter by community ID")
	recordsSearchCmd.Flags().Bool("all", false, "Fetch all pages (up to 10k results)")

	// records get flags
	recordsGetCmd.Flags().String("format", "", "Response format: json, bibtex, datacite (default: uses --output)")

	recordsCmd.AddCommand(recordsListCmd)
	recordsCmd.AddCommand(recordsSearchCmd)
	recordsCmd.AddCommand(recordsGetCmd)
	recordsCmd.AddCommand(recordsVersionsCmd)
	rootCmd.AddCommand(recordsCmd)
}
