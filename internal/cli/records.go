package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

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

Without --community, lists your own records.
With --community (no value), aggregates records across your communities.
With --community=<slug>, lists all records in that community.

Examples:
  zenodo records list
  zenodo records list --status draft
  zenodo records list --community
  zenodo records list --community=my-org`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(appCtx.BaseURL, appCtx.Token)
		status, _ := cmd.Flags().GetString("status")
		community, _ := cmd.Flags().GetString("community")
		communityUsed := cmd.Flags().Changed("community")
		authored, _ := cmd.Flags().GetBool("authored")

		fields := appCtx.Fields

		// --authored: search by ORCID
		if authored {
			if status == "draft" {
				return fmt.Errorf("--authored cannot be used with --status draft (drafts are not available via the search API)")
			}
			orcid := fmt.Sprintf("%v", appCtx.Config.Get("orcid"))
			if orcid == "" || orcid == "<nil>" {
				return fmt.Errorf("ORCID not configured. Run: zenodo config set orcid <your-orcid>")
			}
			if fields == "" {
				fields = "title,community,doi,stats.version_views,stats.version_downloads,created"
			}
			query := fmt.Sprintf("creators.orcid:%s", orcid)
			params := api.RecordListParams{
				Community: community,
			}
			result, err := client.SearchRecords(query, params)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Showing %d of %d authored records\n", len(result.Hits.Hits), result.Hits.Total)
			return output.Format(os.Stdout, result.Hits.Hits, appCtx.Output, fields)
		}

		// --community=<slug>: all records in that community
		if communityUsed && community != "" && community != "*" {
			if fields == "" {
				fields = "community,title,doi,stats.version_views,stats.version_downloads,created"
			}
			params := api.RecordListParams{
				Status:    status,
				Community: community,
			}
			result, err := client.SearchRecords("", params)
			if err != nil {
				return err
			}
			rows, err := injectCommunity(result.Hits.Hits, community)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Showing %d of %d records in %s\n", len(result.Hits.Hits), result.Hits.Total, community)
			return output.Format(os.Stdout, rows, appCtx.Output, fields)
		}

		// --community (no value): aggregate across user's communities
		if communityUsed && (community == "" || community == "*") {
			if fields == "" {
				fields = "community,title,doi,stats.version_views,stats.version_downloads,created"
			}
			communities, err := client.ListUserCommunities("", 0, 0)
			if err != nil {
				return err
			}
			if len(communities.Hits.Hits) == 0 {
				fmt.Fprintln(os.Stderr, "No communities found")
				return nil
			}
			var allRows []map[string]interface{}
			for _, c := range communities.Hits.Hits {
				params := api.RecordListParams{
					Status:    status,
					Community: c.Slug,
				}
				result, err := client.SearchRecords("", params)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not fetch records for %s: %v\n", c.Slug, err)
					continue
				}
				rows, err := injectCommunity(result.Hits.Hits, c.Slug)
				if err != nil {
					return err
				}
				allRows = append(allRows, rows...)
			}
			fmt.Fprintf(os.Stderr, "Total: %d records across %d communities\n", len(allRows), len(communities.Hits.Hits))
			return output.Format(os.Stdout, allRows, appCtx.Output, fields)
		}

		// Default: user's own records
		if fields == "" {
			fields = "title,community,doi,created"
		}
		params := api.RecordListParams{
			Status:    status,
		}
		depositions, err := client.ListUserRecords(params)
		if err != nil {
			return err
		}
		// Normalize community field: extract identifiers from metadata.communities
		rows, err := normalizeCommunities(depositions)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Showing %d records\n", len(depositions))
		return output.Format(os.Stdout, rows, appCtx.Output, fields)
	},
}

// normalizeCommunities converts depositions to maps and extracts metadata.communities
// into a top-level "community" field as a comma-separated string of identifiers.
func normalizeCommunities(data interface{}) ([]map[string]interface{}, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var rows []map[string]interface{}
	if err := json.Unmarshal(b, &rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if mc, ok := row["metadata"]; ok {
			if meta, ok := mc.(map[string]interface{}); ok {
				if communities, ok := meta["communities"]; ok {
					if arr, ok := communities.([]interface{}); ok {
						var slugs []string
						for _, item := range arr {
							if m, ok := item.(map[string]interface{}); ok {
								if id, ok := m["identifier"]; ok {
									slugs = append(slugs, fmt.Sprintf("%v", id))
								}
							}
						}
						row["community"] = strings.Join(slugs, ", ")
					}
				}
			}
		}
	}
	return rows, nil
}

// injectCommunity converts records to maps and adds a top-level "community" field.
func injectCommunity(data interface{}, slug string) ([]map[string]interface{}, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var rows []map[string]interface{}
	if err := json.Unmarshal(b, &rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		row["community"] = slug
	}
	return rows, nil
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
			searchFields = "id,title,doi,stats.version_views,stats.version_downloads,created"
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
			fields = "id,title,doi,stats.version_views,stats.version_downloads,created"
		}
		return output.Format(os.Stdout, result.Hits.Hits, appCtx.Output, fields)
	},
}

func init() {
	// records list flags
	recordsListCmd.Flags().String("status", "", "Filter by status: draft, published")
	recordsListCmd.Flags().String("community", "", "Community slug (omit value to aggregate across your communities)")
	recordsListCmd.Flags().Lookup("community").NoOptDefVal = "*"
	recordsListCmd.Flags().Bool("authored", false, "List records where you are a creator (by ORCID)")


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
