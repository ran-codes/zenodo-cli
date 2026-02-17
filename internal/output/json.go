package output

import (
	"encoding/json"
	"io"
)

func formatJSON(w io.Writer, data interface{}, fields string) error {
	fieldList := parseFields(fields)

	if len(fieldList) > 0 {
		rows, err := toRows(data)
		if err != nil {
			return err
		}
		rows = filterFields(rows, fieldList)
		data = rows
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
