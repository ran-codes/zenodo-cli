package api

import (
	"net/url"
	"strconv"

	"github.com/ran-codes/zenodo-cli/internal/model"
)

// SearchLicenses searches available licenses.
func (c *Client) SearchLicenses(q string, page, size int) (*model.LicenseSearchResult, error) {
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

	var result model.LicenseSearchResult
	if err := c.Get("/licenses/", query, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
