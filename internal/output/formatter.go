package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Format writes data in the specified format to the writer.
// Supported formats: "json", "table", "csv".
// fields is an optional comma-separated list of fields to include.
func Format(w io.Writer, data interface{}, format string, fields string) error {
	switch format {
	case "json":
		return formatJSON(w, data, fields)
	case "table":
		return formatTable(w, data, fields)
	case "csv":
		return formatCSV(w, data, fields)
	default:
		return fmt.Errorf("unsupported output format: %q", format)
	}
}

// toRows converts data into a slice of maps for tabular rendering.
// Handles both single items and slices.
func toRows(data interface{}) ([]map[string]interface{}, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Try as array first.
	var rows []map[string]interface{}
	if err := json.Unmarshal(b, &rows); err == nil {
		return rows, nil
	}

	// Try as single object.
	var single map[string]interface{}
	if err := json.Unmarshal(b, &single); err == nil {
		return []map[string]interface{}{single}, nil
	}

	return nil, fmt.Errorf("data cannot be converted to tabular format")
}

// parseFields splits a comma-separated fields string into a slice.
func parseFields(fields string) []string {
	if fields == "" {
		return nil
	}
	parts := strings.Split(fields, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// filterFields keeps only specified fields in each row.
func filterFields(rows []map[string]interface{}, fields []string) []map[string]interface{} {
	if len(fields) == 0 {
		return rows
	}
	result := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		filtered := make(map[string]interface{})
		for _, f := range fields {
			if val, ok := resolveNestedField(row, f); ok {
				filtered[f] = val
			}
		}
		result[i] = filtered
	}
	return result
}

// resolveNestedField resolves dotted field paths like "stats.downloads".
func resolveNestedField(row map[string]interface{}, field string) (interface{}, bool) {
	parts := strings.Split(field, ".")
	var current interface{} = row
	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current, ok = m[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

// detectColumns returns column names from data, preserving a useful default order.
func detectColumns(rows []map[string]interface{}, fields []string) []string {
	if len(fields) > 0 {
		return fields
	}
	if len(rows) == 0 {
		return nil
	}

	// Collect all unique keys from all rows.
	seen := make(map[string]bool)
	var cols []string
	for _, row := range rows {
		for k := range row {
			if !seen[k] {
				seen[k] = true
				cols = append(cols, k)
			}
		}
	}
	return cols
}

// stringify converts a value to a display string.
func stringify(v interface{}) string {
	if v == nil {
		return ""
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice:
		// Flatten arrays of simple objects into comma-separated values.
		// e.g. [{"identifier":"a"},{"identifier":"b"}] → "a, b"
		if items, ok := v.([]interface{}); ok && len(items) > 0 {
			var vals []string
			for _, item := range items {
				m, ok := item.(map[string]interface{})
				if !ok || len(m) != 1 {
					// Not a single-key object array; fall back to JSON.
					b, _ := json.Marshal(v)
					return string(b)
				}
				for _, val := range m {
					vals = append(vals, fmt.Sprintf("%v", val))
				}
			}
			return strings.Join(vals, ", ")
		}
		b, _ := json.Marshal(v)
		return string(b)
	case reflect.Map:
		b, _ := json.Marshal(v)
		return string(b)
	default:
		return fmt.Sprintf("%v", v)
	}
}
