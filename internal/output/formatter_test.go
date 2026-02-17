package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

type testRecord struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	DOI   string `json:"doi"`
}

var sampleRecords = []testRecord{
	{ID: 1, Title: "First Record", DOI: "10.5281/1"},
	{ID: 2, Title: "Second Record", DOI: "10.5281/2"},
}

func TestFormatJSON(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, sampleRecords, "json", "")
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	var result []testRecord
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 records, got %d", len(result))
	}
	if result[0].Title != "First Record" {
		t.Errorf("title = %q, want %q", result[0].Title, "First Record")
	}
}

func TestFormatJSON_WithFields(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, sampleRecords, "json", "id,title")
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if _, ok := result[0]["doi"]; ok {
		t.Error("doi field should be filtered out")
	}
	if _, ok := result[0]["id"]; !ok {
		t.Error("id field should be present")
	}
}

func TestFormatTable(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, sampleRecords, "table", "")
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "First Record") {
		t.Error("table should contain 'First Record'")
	}
	if !strings.Contains(out, "Second Record") {
		t.Error("table should contain 'Second Record'")
	}
}

func TestFormatTable_WithFields(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, sampleRecords, "table", "id,title")
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "ID") {
		t.Error("table should contain ID header")
	}
	if !strings.Contains(out, "TITLE") {
		t.Error("table should contain TITLE header")
	}
}

func TestFormatCSV(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, sampleRecords, "csv", "id,title,doi")
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 { // header + 2 data rows
		t.Errorf("expected 3 CSV lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "id,title,doi" {
		t.Errorf("CSV header = %q, want %q", lines[0], "id,title,doi")
	}
}

func TestFormatCSV_WithFields(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, sampleRecords, "csv", "id")
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if lines[0] != "id" {
		t.Errorf("CSV header = %q, want %q", lines[0], "id")
	}
}

func TestFormatSingleObject(t *testing.T) {
	single := testRecord{ID: 42, Title: "Solo", DOI: "10.5281/42"}

	var buf bytes.Buffer
	err := Format(&buf, single, "json", "")
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	// Should still be valid JSON.
	var result interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
}

func TestNestedFields(t *testing.T) {
	data := []map[string]interface{}{
		{
			"id":    1,
			"title": "Test",
			"stats": map[string]interface{}{
				"downloads": 100,
				"views":     200,
			},
		},
	}

	var buf bytes.Buffer
	err := Format(&buf, data, "csv", "id,stats.downloads")
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "100") {
		t.Error("CSV should contain nested stats.downloads value 100")
	}
}

func TestUnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	err := Format(&buf, sampleRecords, "xml", "")
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestParseFields(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"id", 1},
		{"id,title,doi", 3},
		{" id , title ", 2},
	}
	for _, tt := range tests {
		got := parseFields(tt.input)
		if len(got) != tt.want {
			t.Errorf("parseFields(%q) = %d fields, want %d", tt.input, len(got), tt.want)
		}
	}
}
