package output

import (
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
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

	table := tablewriter.NewWriter(w)
	headers := make([]string, len(cols))
	for i, c := range cols {
		headers[i] = strings.ToUpper(c)
	}
	table.SetHeader(headers)
	table.SetBorder(false)
	table.SetAutoWrapText(false)

	for _, row := range rows {
		record := make([]string, len(cols))
		for i, col := range cols {
			record[i] = stringify(row[col])
		}
		table.Append(record)
	}

	table.Render()
	return nil
}
