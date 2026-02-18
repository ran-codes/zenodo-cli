package model

import "time"

// Community represents a Zenodo community.
type Community struct {
	ID       string            `json:"id"`
	Slug     string            `json:"slug,omitempty"`
	URL      string            `json:"url,omitempty"`
	Metadata CommunityMetadata `json:"metadata,omitempty"`
	Created  time.Time         `json:"created"`
	Updated  time.Time         `json:"updated"`
	Links    CommunityLinks    `json:"links,omitempty"`
}

// CommunityLinks contains links returned by the communities API.
type CommunityLinks struct {
	Self     string `json:"self,omitempty"`
	SelfHTML string `json:"self_html,omitempty"`
}

// CommunityMetadata contains the community's metadata fields.
type CommunityMetadata struct {
	Title          string `json:"title,omitempty"`
	Description    string `json:"description,omitempty"`
	CurationPolicy string `json:"curation_policy,omitempty"`
	Website        string `json:"website,omitempty"`
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
