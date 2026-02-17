// Package integration contains integration tests that run against the Zenodo sandbox.
// These tests require a ZENODO_SANDBOX_TOKEN environment variable to be set.
// Run with: go test ./test/integration/ -tags=integration -v
//
//go:build integration

package integration

import (
	"os"
	"testing"

	"github.com/ran-codes/zenodo-cli/internal/api"
)

func sandboxClient(t *testing.T) *api.Client {
	t.Helper()
	token := os.Getenv("ZENODO_SANDBOX_TOKEN")
	if token == "" {
		t.Skip("ZENODO_SANDBOX_TOKEN not set, skipping integration test")
	}
	return api.NewClient("https://sandbox.zenodo.org/api", token)
}

func TestIntegration_SearchRecords(t *testing.T) {
	client := sandboxClient(t)

	result, err := client.SearchRecords("test", api.RecordListParams{Size: 5})
	if err != nil {
		t.Fatalf("SearchRecords error: %v", err)
	}
	if result.Hits.Total < 0 {
		t.Error("expected non-negative total")
	}
	t.Logf("Found %d total records (showing %d)", result.Hits.Total, len(result.Hits.Hits))
}

func TestIntegration_SearchCommunities(t *testing.T) {
	client := sandboxClient(t)

	result, err := client.SearchCommunities("", 1, 5)
	if err != nil {
		t.Fatalf("SearchCommunities error: %v", err)
	}
	t.Logf("Found %d communities", result.Hits.Total)
}

func TestIntegration_SearchLicenses(t *testing.T) {
	client := sandboxClient(t)

	result, err := client.SearchLicenses("creative commons", 1, 5)
	if err != nil {
		t.Fatalf("SearchLicenses error: %v", err)
	}
	if result.Hits.Total == 0 {
		t.Error("expected at least one license matching 'creative commons'")
	}
	t.Logf("Found %d licenses", result.Hits.Total)
}
