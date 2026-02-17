package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ran-codes/zenodo-cli/internal/model"
)

func TestSearchLicenses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/licenses/" {
			t.Errorf("path = %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(model.LicenseSearchResult{
			Hits: model.LicenseHits{
				Hits:  []model.License{{ID: "cc-by-4.0", Title: "Creative Commons Attribution 4.0"}},
				Total: 1,
			},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	result, err := client.SearchLicenses("creative commons", 0, 0)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result.Hits.Hits[0].ID != "cc-by-4.0" {
		t.Errorf("id = %q", result.Hits.Hits[0].ID)
	}
}
