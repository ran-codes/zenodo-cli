package api

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/ran-codes/zenodo-cli/internal/model"
)

// ListUserRecords returns the authenticated user's records and drafts.
func (c *Client) ListUserRecords(params RecordListParams) (*model.RecordSearchResult, error) {
	query := params.toQuery()
	var result model.RecordSearchResult
	if err := c.Get("/api/records", query, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SearchRecords searches published records with an Elasticsearch query.
func (c *Client) SearchRecords(q string, params RecordListParams) (*model.RecordSearchResult, error) {
	query := params.toQuery()
	if q != "" {
		query.Set("q", q)
	}
	var result model.RecordSearchResult
	if err := c.Get("/records/", query, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRecord retrieves a single published record by ID.
func (c *Client) GetRecord(id int) (*model.Record, error) {
	var result model.Record
	if err := c.Get(fmt.Sprintf("/records/%d", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListVersions returns all versions of a record.
func (c *Client) ListVersions(id int) (*model.RecordSearchResult, error) {
	var result model.RecordSearchResult
	if err := c.Get(fmt.Sprintf("/api/records/%d/versions", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RecordListParams holds query parameters for listing/searching records.
type RecordListParams struct {
	Page      int
	Size      int
	Status    string
	Community string
	Sort      string
}

func (p RecordListParams) toQuery() url.Values {
	q := url.Values{}
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	if p.Size > 0 {
		q.Set("size", strconv.Itoa(p.Size))
	} else {
		q.Set("size", "100") // Max page size for authenticated requests.
	}
	if p.Status != "" {
		q.Set("status", p.Status)
	}
	if p.Community != "" {
		q.Set("communities", p.Community)
	}
	if p.Sort != "" {
		q.Set("sort", p.Sort)
	}
	return q
}

// PaginateAll fetches all pages of results up to the 10k ceiling.
// It calls fetchPage repeatedly, which should return (result, error).
func PaginateAll(fetchPage func(page int) (*model.RecordSearchResult, error)) ([]model.Record, int, error) {
	var allRecords []model.Record
	const maxResults = 10000
	const pageSize = 100

	for page := 1; ; page++ {
		result, err := fetchPage(page)
		if err != nil {
			return nil, 0, err
		}

		allRecords = append(allRecords, result.Hits.Hits...)

		// Check if we've fetched all results or hit the ceiling.
		if len(allRecords) >= result.Hits.Total || len(result.Hits.Hits) < pageSize {
			return allRecords, result.Hits.Total, nil
		}
		if len(allRecords) >= maxResults {
			return allRecords, result.Hits.Total, fmt.Errorf("results truncated at %d (total: %d); narrow your search", maxResults, result.Hits.Total)
		}
	}
}
