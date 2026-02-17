package api

import (
	"net/url"
	"strconv"

	"github.com/ran-codes/zenodo-cli/internal/model"
)

// SearchCommunities searches or lists communities.
func (c *Client) SearchCommunities(q string, page, size int) (*model.CommunitySearchResult, error) {
	query := url.Values{}
	if q != "" {
		query.Set("q", q)
	}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if size > 0 {
		query.Set("size", strconv.Itoa(size))
	} else {
		query.Set("size", "100")
	}

	var result model.CommunitySearchResult
	if err := c.Get("/communities", query, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
