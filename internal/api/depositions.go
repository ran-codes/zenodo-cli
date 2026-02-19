package api

import (
	"fmt"

	"github.com/ran-codes/zenodo-cli/internal/model"
)

// GetDeposition retrieves a deposition by ID.
func (c *Client) GetDeposition(id int) (*model.Deposition, error) {
	var result model.Deposition
	if err := c.Get(fmt.Sprintf("/deposit/depositions/%d", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// v0.1: Write methods disabled — read-only release.
// They will be re-enabled in a future version.

/*
// UpdateDeposition updates the metadata of a deposition (full replacement PUT).
func (c *Client) UpdateDeposition(id int, metadata model.Metadata) (*model.Deposition, error) {
	body := map[string]interface{}{
		"metadata": metadata,
	}
	var result model.Deposition
	if err := c.Put(fmt.Sprintf("/deposit/depositions/%d", id), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// EditDeposition unlocks a published record for editing.
func (c *Client) EditDeposition(id int) (*model.Deposition, error) {
	var result model.Deposition
	if err := c.Post(fmt.Sprintf("/deposit/depositions/%d/actions/edit", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PublishDeposition publishes (or re-publishes) a deposition.
func (c *Client) PublishDeposition(id int) (*model.Deposition, error) {
	var result model.Deposition
	if err := c.Post(fmt.Sprintf("/deposit/depositions/%d/actions/publish", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DiscardDeposition discards changes on an unpublished deposition.
func (c *Client) DiscardDeposition(id int) (*model.Deposition, error) {
	var result model.Deposition
	if err := c.Post(fmt.Sprintf("/deposit/depositions/%d/actions/discard", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
*/
