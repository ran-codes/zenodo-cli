package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ran-codes/zenodo-cli/internal/model"
)

func TestGetDeposition(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/deposit/depositions/100" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(model.Deposition{
			ID:       100,
			Metadata: model.Metadata{Title: "Test Deposition"},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	dep, err := client.GetDeposition(100)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if dep.ID != 100 || dep.Metadata.Title != "Test Deposition" {
		t.Errorf("unexpected deposition: %+v", dep)
	}
}

func TestUpdateDeposition(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/deposit/depositions/100" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var payload map[string]model.Metadata
		json.Unmarshal(body, &payload)
		if payload["metadata"].Title != "Updated Title" {
			t.Errorf("title = %q", payload["metadata"].Title)
		}
		json.NewEncoder(w).Encode(model.Deposition{
			ID:       100,
			Metadata: model.Metadata{Title: "Updated Title"},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	dep, err := client.UpdateDeposition(100, model.Metadata{Title: "Updated Title"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if dep.Metadata.Title != "Updated Title" {
		t.Errorf("title = %q", dep.Metadata.Title)
	}
}

func TestEditDeposition(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/deposit/depositions/100/actions/edit" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(model.Deposition{ID: 100, State: "inprogress"})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	dep, err := client.EditDeposition(100)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if dep.State != "inprogress" {
		t.Errorf("state = %q", dep.State)
	}
}

func TestPublishDeposition(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/deposit/depositions/100/actions/publish" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(model.Deposition{ID: 100, State: "done"})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	dep, err := client.PublishDeposition(100)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if dep.State != "done" {
		t.Errorf("state = %q", dep.State)
	}
}

func TestDiscardDeposition(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/deposit/depositions/100/actions/discard" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(model.Deposition{ID: 100, State: "done"})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	dep, err := client.DiscardDeposition(100)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if dep.ID != 100 {
		t.Errorf("id = %d", dep.ID)
	}
}
