package output

import (
	"encoding/csv"
	"io"
)

func formatCSV(w io.Writer, data interface{}, fields string) error {
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

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header row.
	if err := writer.Write(cols); err != nil {
		return err
	}

	// Data rows.
	for _, row := range rows {
		record := make([]string, len(cols))
		for i, col := range cols {
			record[i] = stringify(row[col])
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}
