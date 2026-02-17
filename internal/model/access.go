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
	Hits []AccessLink `json:"hits"`
}
