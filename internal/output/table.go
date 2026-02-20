package output

import (
	"io"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
)

func formatTable(w io.Writer, data interface{}, fields string) error {
	fieldList := parseFields(fields)

	rows, err := toRows(data)
	if err != nil {
		return err
	}
	rows = filterFields(rows, fieldList)

	cols := detectColumns(rows, fieldList)
	if len(cols) == 0 {
		return nil
	}

	// Convert all values to strings first.
	stringRows := make([][]string, len(rows))
	for i, row := range rows {
		record := make([]string, len(cols))
		for j, col := range cols {
			record[j] = stringify(row[col])
		}
		stringRows[i] = record
	}

	// Truncate wide columns dynamically based on terminal width.
	// Columns marked as truncatable share the available space.
	truncatable := map[string]bool{"title": true, "links.doi": true}
	var truncIdxs []int
	for i, c := range cols {
		if truncatable[strings.ToLower(c)] {
			truncIdxs = append(truncIdxs, i)
		}
	}
	if len(truncIdxs) > 0 {
		maxWidth := truncatableMaxWidth(stringRows, cols, truncIdxs)
		if maxWidth > 0 {
			for _, record := range stringRows {
				for _, idx := range truncIdxs {
					record[idx] = truncate(record[idx], maxWidth)
				}
			}
		}
	}

	table := tablewriter.NewWriter(w)
	headers := make([]string, len(cols))
	for i, c := range cols {
		headers[i] = columnHeader(c)
	}
	table.SetHeader(headers)
	table.SetBorder(false)
	table.SetAutoWrapText(false)

	for _, record := range stringRows {
		table.Append(record)
	}

	table.Render()
	return nil
}

// terminalWidth returns the terminal width, or 120 as a fallback.
func terminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 120
	}
	return width
}

// columnAliases maps field names to shorter display headers.
var columnAliases = map[string]string{
	"links.doi":              "DOI",
	"stats.version_views":    "VIEWS",
	"stats.version_downloads": "DOWNLOADS",
}

// columnHeader returns the display header for a column.
func columnHeader(col string) string {
	if alias, ok := columnAliases[strings.ToLower(col)]; ok {
		return alias
	}
	return strings.ToUpper(col)
}

// truncatableMaxWidth calculates the available width per truncatable column
// by subtracting fixed columns from the terminal width, then splitting
// the remaining space evenly among truncatable columns.
func truncatableMaxWidth(rows [][]string, cols []string, truncIdxs []int) int {
	tw := terminalWidth()

	truncSet := make(map[int]bool)
	for _, idx := range truncIdxs {
		truncSet[idx] = true
	}

	// Calculate max width of each fixed column.
	otherWidth := 0
	for i := range cols {
		if truncSet[i] {
			continue
		}
		maxW := len(columnHeader(cols[i])) // header width
		for _, row := range rows {
			if len(row[i]) > maxW {
				maxW = len(row[i])
			}
		}
		otherWidth += maxW
	}

	// Account for table separators: 3 chars per column boundary (" | "),
	// plus 2 chars padding on each side.
	separators := (len(cols) - 1) * 3
	padding := len(cols) * 2
	available := tw - otherWidth - separators - padding

	// Split available space among truncatable columns.
	perCol := available / len(truncIdxs)

	if perCol < 20 {
		return 20 // minimum width per column
	}
	return perCol
}

// truncate shortens a string to maxLen runes, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}
