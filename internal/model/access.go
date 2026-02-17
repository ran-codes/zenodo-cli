package model

import "time"

// AccessLink represents a share/access link for a record.
type AccessLink struct {
	ID        string    `json:"id"`
	Token     string    `json:"token,omitempty"`
	Created   time.Time `json:"created"`
	Links     Links     `json:"links,omitempty"`
}

// AccessLinkList is the response from the access links endpoint.
type AccessLinkList struct {
	Hits AccessLinkHits `json:"hits"`
}

// AccessLinkHits contains access link results.
type AccessLinkHits struct {
	Hits  []AccessLink `json:"hits"`
	Total int          `json:"total"`
}
