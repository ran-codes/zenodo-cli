package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ran-codes/zenodo-cli/internal/api"
	"github.com/ran-codes/zenodo-cli/internal/model"
	"github.com/ran-codes/zenodo-cli/internal/output"
	"github.com/ran-codes/zenodo-cli/internal/validate"
	"github.com/spf13/cobra"
)

var depositCmd = &cobra.Command{
	Use:   "deposit",
	Short: "Edit and manage depositions",
}

var depositEditCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Unlock a published record for editing",
	Long: `Unlock a published record so its metadata can be updated.

Examples:
  zenodo deposit edit 12345`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid deposition ID: %s", args[0])
		}

		client := api.NewClient(appCtx.BaseURL, appCtx.Token)
		dep, err := client.EditDeposition(id)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Deposition %d unlocked for editing (state: %s)\n", dep.ID, dep.State)
		return nil
	},
}

var depositUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update metadata on a deposition",
	Long: `Update metadata using GET-merge-PUT. Fetches current metadata, merges your
changes, validates, shows a diff, and asks for confirmation before PUT.

Changes can come from:
  --title, --description  Inline field flags
  --file metadata.json    JSON file with partial metadata
  --stdin                 Read JSON metadata from stdin

Examples:
  zenodo deposit update 12345 --title "New Title"
  zenodo deposit update 12345 --file metadata.json
  cat changes.json | zenodo deposit update 12345 --stdin
  zenodo deposit update 12345 --title "New Title" --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid deposition ID: %s", args[0])
		}

		client := api.NewClient(appCtx.BaseURL, appCtx.Token)

		// 1. GET current metadata.
		dep, err := client.GetDeposition(id)
		if err != nil {
			return fmt.Errorf("fetching deposition: %w", err)
		}

		// 2. Build merged metadata.
		merged := dep.Metadata
		if err := applyChanges(cmd, &merged); err != nil {
			return err
		}

		// 3. Validate.
		if errs := validate.Metadata(merged); len(errs) > 0 {
			fmt.Fprintln(os.Stderr, "Validation errors:")
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "  - %s\n", e)
			}
			return fmt.Errorf("metadata validation failed")
		}

		// 4. Show diff.
		changed, err := output.DiffMetadata(os.Stderr, dep.Metadata, merged)
		if err != nil {
			return err
		}
		if !changed {
			return nil
		}

		// 5. Dry run — stop here.
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Fprintln(os.Stderr, "Dry run — no changes applied.")
			return nil
		}

		// 6. Confirm.
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			if !confirm("Apply these changes?") {
				fmt.Fprintln(os.Stderr, "Cancelled.")
				os.Exit(5)
			}
		}

		// 7. PUT merged metadata.
		result, err := client.UpdateDeposition(id, merged)
		if err != nil {
			return fmt.Errorf("updating deposition: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Deposition %d metadata updated.\n", result.ID)
		return nil
	},
}

var depositDiscardCmd = &cobra.Command{
	Use:   "discard <id>",
	Short: "Discard unpublished changes",
	Long: `Discard any unpublished changes on a deposition, reverting to the published version.

Examples:
  zenodo deposit discard 12345`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid deposition ID: %s", args[0])
		}

		client := api.NewClient(appCtx.BaseURL, appCtx.Token)
		dep, err := client.DiscardDeposition(id)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Changes discarded on deposition %d (state: %s)\n", dep.ID, dep.State)
		return nil
	},
}

var depositPublishCmd = &cobra.Command{
	Use:   "publish <id>",
	Short: "Re-publish a deposition after editing",
	Long: `Re-publish a deposition. Shows a diff of changes since last publish
and asks for confirmation.

Examples:
  zenodo deposit publish 12345
  zenodo deposit publish 12345 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid deposition ID: %s", args[0])
		}

		client := api.NewClient(appCtx.BaseURL, appCtx.Token)

		// Get current state to show info.
		dep, err := client.GetDeposition(id)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Publishing deposition %d: %s\n", dep.ID, dep.Metadata.Title)

		// Confirm.
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			if !confirm("Publish this deposition?") {
				fmt.Fprintln(os.Stderr, "Cancelled.")
				os.Exit(5)
			}
		}

		result, err := client.PublishDeposition(id)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Deposition %d published (state: %s)\n", result.ID, result.State)
		if result.DOI != "" {
			fmt.Fprintf(os.Stderr, "DOI: %s\n", result.DOI)
		}
		return nil
	},
}

func init() {
	// deposit update flags
	depositUpdateCmd.Flags().String("title", "", "Set title")
	depositUpdateCmd.Flags().String("description", "", "Set description")
	depositUpdateCmd.Flags().String("file", "", "JSON file with metadata changes")
	depositUpdateCmd.Flags().Bool("stdin", false, "Read metadata changes from stdin")
	depositUpdateCmd.Flags().Bool("dry-run", false, "Show diff without applying changes")
	depositUpdateCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	// deposit publish flags
	depositPublishCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	depositCmd.AddCommand(depositEditCmd)
	depositCmd.AddCommand(depositUpdateCmd)
	depositCmd.AddCommand(depositDiscardCmd)
	depositCmd.AddCommand(depositPublishCmd)
	rootCmd.AddCommand(depositCmd)
}

// applyChanges merges changes from flags, --file, or --stdin into the metadata.
func applyChanges(cmd *cobra.Command, m *model.Metadata) error {
	// Apply inline flags.
	if title, _ := cmd.Flags().GetString("title"); title != "" {
		m.Title = title
	}
	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		m.Description = desc
	}

	// Apply from file.
	filePath, _ := cmd.Flags().GetString("file")
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading metadata file: %w", err)
		}
		if err := mergeJSON(m, data); err != nil {
			return fmt.Errorf("parsing metadata file: %w", err)
		}
	}

	// Apply from stdin.
	useStdin, _ := cmd.Flags().GetBool("stdin")
	if useStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		if err := mergeJSON(m, data); err != nil {
			return fmt.Errorf("parsing stdin metadata: %w", err)
		}
	}

	return nil
}

// mergeJSON unmarshals JSON into the metadata, overwriting only specified fields.
func mergeJSON(m *model.Metadata, data []byte) error {
	// Marshal current state, merge with incoming, unmarshal back.
	current, err := json.Marshal(m)
	if err != nil {
		return err
	}

	var base map[string]interface{}
	if err := json.Unmarshal(current, &base); err != nil {
		return err
	}

	var overlay map[string]interface{}
	if err := json.Unmarshal(data, &overlay); err != nil {
		return err
	}

	// Overlay wins for any key it specifies.
	for k, v := range overlay {
		base[k] = v
	}

	merged, err := json.Marshal(base)
	if err != nil {
		return err
	}

	return json.Unmarshal(merged, m)
}

// confirm prompts the user for y/n confirmation.
func confirm(prompt string) bool {
	fmt.Fprintf(os.Stderr, "%s [y/N] ", prompt)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}
