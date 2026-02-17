package api

import (
	"fmt"

	"github.com/ran-codes/zenodo-cli/internal/model"
)

// ListAccessLinks returns the share/access links for a record.
func (c *Client) ListAccessLinks(recordID int) ([]model.AccessLink, error) {
	var result model.AccessLinkList
	if err := c.Get(fmt.Sprintf("/records/%d/access/links", recordID), nil, &result); err != nil {
		return nil, err
	}
	return result.Hits.Hits, nil
}
