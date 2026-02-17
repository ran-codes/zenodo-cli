package model

import "time"

// Community represents a Zenodo community.
type Community struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	PageURL     string    `json:"page,omitempty"`
	CurationPolicy string `json:"curation_policy,omitempty"`
	LogoURL     string    `json:"logo_url,omitempty"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Links       Links     `json:"links,omitempty"`
}

// CommunitySearchResult is the paginated response from communities search.
type CommunitySearchResult struct {
	Hits  CommunityHits `json:"hits"`
	Links Links         `json:"links,omitempty"`
}

// CommunityHits contains the search result hits and total count.
type CommunityHits struct {
	Hits  []Community `json:"hits"`
	Total int         `json:"total"`
}
