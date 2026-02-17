package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/r3labs/diff/v3"
)

var (
	addColor    = color.New(color.FgGreen)
	removeColor = color.New(color.FgRed)
	changeColor = color.New(color.FgYellow)
	pathColor   = color.New(color.Bold)
)

// DiffMetadata computes and displays a colored field-by-field diff between
// old and new metadata. Returns true if there are changes.
func DiffMetadata(w io.Writer, old, new interface{}) (bool, error) {
	changelog, err := diff.Diff(old, new)
	if err != nil {
		return false, fmt.Errorf("computing diff: %w", err)
	}

	if len(changelog) == 0 {
		fmt.Fprintln(w, "No changes detected.")
		return false, nil
	}

	fmt.Fprintf(w, "Changes (%d):\n", len(changelog))
	for _, change := range changelog {
		path := formatPath(change.Path)
		switch change.Type {
		case diff.CREATE:
			pathColor.Fprintf(w, "  %s: ", path)
			addColor.Fprintf(w, "+ %v\n", formatValue(change.To))
		case diff.DELETE:
			pathColor.Fprintf(w, "  %s: ", path)
			removeColor.Fprintf(w, "- %v\n", formatValue(change.From))
		case diff.UPDATE:
			pathColor.Fprintf(w, "  %s:\n", path)
			removeColor.Fprintf(w, "    - %v\n", formatValue(change.From))
			addColor.Fprintf(w, "    + %v\n", formatValue(change.To))
		}
	}

	return true, nil
}

func formatPath(path []string) string {
	return strings.Join(path, ".")
}

func formatValue(v interface{}) string {
	if v == nil {
		return "<empty>"
	}
	s := fmt.Sprintf("%v", v)
	// Truncate very long values for readability.
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}
