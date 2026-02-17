package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ran-codes/zenodo-cli/internal/model"
)

func TestSearchRecords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/records" {
			t.Errorf("path = %q, want /records/", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "climate" {
			t.Errorf("q = %q, want climate", r.URL.Query().Get("q"))
		}
		json.NewEncoder(w).Encode(model.RecordSearchResult{
			Hits: model.RecordHits{
				Hits:  []model.Record{{ID: 1, Metadata: model.Metadata{Title: "Climate Study"}}},
				Total: 1,
			},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	result, err := client.SearchRecords("climate", RecordListParams{})
	if err != nil {
		t.Fatalf("SearchRecords() error: %v", err)
	}
	if result.Hits.Total != 1 {
		t.Errorf("total = %d, want 1", result.Hits.Total)
	}
	if result.Hits.Hits[0].Metadata.Title != "Climate Study" {
		t.Errorf("title = %q", result.Hits.Hits[0].Metadata.Title)
	}
}

func TestGetRecord(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/records/12345" {
			t.Errorf("path = %q, want /records/12345", r.URL.Path)
		}
		json.NewEncoder(w).Encode(model.Record{
			ID:       12345,
			Metadata: model.Metadata{Title: "My Record"},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	record, err := client.GetRecord(12345)
	if err != nil {
		t.Fatalf("GetRecord() error: %v", err)
	}
	if record.ID != 12345 {
		t.Errorf("ID = %d, want 12345", record.ID)
	}
}

func TestListUserRecords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/records" {
			t.Errorf("path = %q, want /records", r.URL.Path)
		}
		if r.URL.Query().Get("status") != "draft" {
			t.Errorf("status = %q, want draft", r.URL.Query().Get("status"))
		}
		json.NewEncoder(w).Encode(model.RecordSearchResult{
			Hits: model.RecordHits{
				Hits:  []model.Record{{ID: 1}},
				Total: 1,
			},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	result, err := client.ListUserRecords(RecordListParams{Status: "draft"})
	if err != nil {
		t.Fatalf("ListUserRecords() error: %v", err)
	}
	if result.Hits.Total != 1 {
		t.Errorf("total = %d, want 1", result.Hits.Total)
	}
}

func TestListVersions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/records/100/versions" {
			t.Errorf("path = %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(model.RecordSearchResult{
			Hits: model.RecordHits{
				Hits:  []model.Record{{ID: 100}, {ID: 101}},
				Total: 2,
			},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	result, err := client.ListVersions(100)
	if err != nil {
		t.Fatalf("ListVersions() error: %v", err)
	}
	if len(result.Hits.Hits) != 2 {
		t.Errorf("versions count = %d, want 2", len(result.Hits.Hits))
	}
}

func TestRecordListParams_ToQuery(t *testing.T) {
	p := RecordListParams{
		Page:      2,
		Size:      50,
		Status:    "published",
		Community: "my-org",
		Sort:      "mostrecent",
	}
	q := p.toQuery()

	if q.Get("page") != "2" {
		t.Errorf("page = %q", q.Get("page"))
	}
	if q.Get("size") != "50" {
		t.Errorf("size = %q", q.Get("size"))
	}
	if q.Get("status") != "published" {
		t.Errorf("status = %q", q.Get("status"))
	}
	if q.Get("communities") != "my-org" {
		t.Errorf("communities = %q", q.Get("communities"))
	}
}

func TestPaginateAll(t *testing.T) {
	callCount := 0
	records, total, err := PaginateAll(func(page int) (*model.RecordSearchResult, error) {
		callCount++
		if page == 1 {
			hits := make([]model.Record, 100)
			for i := range hits {
				hits[i] = model.Record{ID: i}
			}
			return &model.RecordSearchResult{
				Hits: model.RecordHits{Hits: hits, Total: 150},
			}, nil
		}
		hits := make([]model.Record, 50)
		for i := range hits {
			hits[i] = model.Record{ID: 100 + i}
		}
		return &model.RecordSearchResult{
			Hits: model.RecordHits{Hits: hits, Total: 150},
		}, nil
	})

	if err != nil {
		t.Fatalf("PaginateAll() error: %v", err)
	}
	if len(records) != 150 {
		t.Errorf("records = %d, want 150", len(records))
	}
	if total != 150 {
		t.Errorf("total = %d, want 150", total)
	}
	if callCount != 2 {
		t.Errorf("calls = %d, want 2", callCount)
	}
}
