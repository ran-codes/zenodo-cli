package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ran-codes/zenodo-cli/internal/model"
)

func TestListAccessLinks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/records/123/access/links" {
			t.Errorf("path = %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(model.AccessLinkList{
			Hits: model.AccessLinkHits{
				Hits: []model.AccessLink{{ID: "link-1"}},
				Total: 1,
			},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "tok")
	links, err := client.ListAccessLinks(123)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(links) != 1 {
		t.Errorf("links = %d", len(links))
	}
	if links[0].ID != "link-1" {
		t.Errorf("id = %q", links[0].ID)
	}
}
