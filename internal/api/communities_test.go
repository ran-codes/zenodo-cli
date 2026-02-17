package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ran-codes/zenodo-cli/internal/model"
)

func TestSearchCommunities(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/communities" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "open science" {
			t.Errorf("q = %q", r.URL.Query().Get("q"))
		}
		json.NewEncoder(w).Encode(model.CommunitySearchResult{
			Hits: model.CommunityHits{
				Hits:  []model.Community{{ID: "open-sci", Metadata: model.CommunityMetadata{Title: "Open Science"}}},
				Total: 1,
			},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	result, err := client.SearchCommunities("open science", 0, 0)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result.Hits.Total != 1 {
		t.Errorf("total = %d", result.Hits.Total)
	}
	if result.Hits.Hits[0].ID != "open-sci" {
		t.Errorf("id = %q", result.Hits.Hits[0].ID)
	}
}
