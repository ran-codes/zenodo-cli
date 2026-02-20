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

	// Truncate the "title" column dynamically based on terminal width.
	titleIdx := -1
	for i, c := range cols {
		if strings.EqualFold(c, "title") {
			titleIdx = i
			break
		}
	}
	if titleIdx >= 0 {
		maxTitle := titleMaxWidth(stringRows, cols, titleIdx)
		if maxTitle > 0 {
			for _, record := range stringRows {
				record[titleIdx] = truncate(record[titleIdx], maxTitle)
			}
		}
	}

	table := tablewriter.NewWriter(w)
	headers := make([]string, len(cols))
	for i, c := range cols {
		headers[i] = strings.ToUpper(c)
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

// titleMaxWidth calculates the available width for the title column
// by subtracting the width of all other columns from the terminal width.
func titleMaxWidth(rows [][]string, cols []string, titleIdx int) int {
	tw := terminalWidth()

	// Calculate max width of each non-title column.
	otherWidth := 0
	for i := range cols {
		if i == titleIdx {
			continue
		}
		maxW := len(cols[i]) // header width
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

	if available < 20 {
		return 20 // minimum title width
	}
	return available
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
