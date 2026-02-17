package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func init() {
	// Disable color in tests for predictable output.
	color.NoColor = true
}

type testMeta struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

func TestDiffMetadata_NoChanges(t *testing.T) {
	old := testMeta{Title: "Test", Description: "Desc"}
	new := testMeta{Title: "Test", Description: "Desc"}

	var buf bytes.Buffer
	changed, err := DiffMetadata(&buf, old, new)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if changed {
		t.Error("expected no changes")
	}
	if !strings.Contains(buf.String(), "No changes") {
		t.Errorf("output = %q", buf.String())
	}
}

func TestDiffMetadata_UpdateField(t *testing.T) {
	old := testMeta{Title: "Old Title", Description: "Desc"}
	new := testMeta{Title: "New Title", Description: "Desc"}

	var buf bytes.Buffer
	changed, err := DiffMetadata(&buf, old, new)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !changed {
		t.Error("expected changes")
	}
	out := buf.String()
	if !strings.Contains(out, "Old Title") || !strings.Contains(out, "New Title") {
		t.Errorf("output = %q", out)
	}
}

func TestDiffMetadata_AddField(t *testing.T) {
	old := testMeta{Title: "Test"}
	new := testMeta{Title: "Test", Description: "Added desc"}

	var buf bytes.Buffer
	changed, err := DiffMetadata(&buf, old, new)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !changed {
		t.Error("expected changes")
	}
	if !strings.Contains(buf.String(), "Added desc") {
		t.Errorf("output = %q", buf.String())
	}
}

func TestDiffMetadata_RemoveField(t *testing.T) {
	old := testMeta{Title: "Test", Description: "Will be removed"}
	new := testMeta{Title: "Test"}

	var buf bytes.Buffer
	changed, err := DiffMetadata(&buf, old, new)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !changed {
		t.Error("expected changes")
	}
	if !strings.Contains(buf.String(), "Will be removed") {
		t.Errorf("output = %q", buf.String())
	}
}

func TestDiffMetadata_ArrayChange(t *testing.T) {
	old := testMeta{Title: "Test", Keywords: []string{"a", "b"}}
	new := testMeta{Title: "Test", Keywords: []string{"a", "c"}}

	var buf bytes.Buffer
	changed, err := DiffMetadata(&buf, old, new)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !changed {
		t.Error("expected changes")
	}
}
