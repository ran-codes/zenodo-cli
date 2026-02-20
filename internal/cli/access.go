package cli

// v0.1: Access links commands disabled — read-only release.
// They will be re-enabled in a future version.

/*
import (
	"fmt"
	"os"
	"strconv"

	"github.com/ran-codes/zenodo-cli/internal/api"
	"github.com/ran-codes/zenodo-cli/internal/output"
	"github.com/spf13/cobra"
)

var accessCmd = &cobra.Command{
	Use:   "access",
	Short: "Manage access links",
}

var accessLinksCmd = &cobra.Command{
	Use:   "links",
	Short: "Manage share links for records",
}

var accessLinksListCmd = &cobra.Command{
	Use:   "list <record-id>",
	Short: "List share links for a record",
	Long: `List all share/access links for a record.

Examples:
  zenodo access links list 12345
  zenodo access links list 12345 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid record ID: %s", args[0])
		}

		client := api.NewClient(appCtx.BaseURL, appCtx.Token)
		links, err := client.ListAccessLinks(id)
		if err != nil {
			return err
		}
		return output.Format(os.Stdout, links, appCtx.Output, appCtx.Fields)
	},
}

func init() {
	accessLinksCmd.AddCommand(accessLinksListCmd)
	accessCmd.AddCommand(accessLinksCmd)
	rootCmd.AddCommand(accessCmd)
}
*/
