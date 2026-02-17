package model

// License represents a Zenodo license.
type License struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url,omitempty"`
	Family      string `json:"family,omitempty"`
	Domain      string `json:"domain_content,omitempty"`
	DomainData  string `json:"domain_data,omitempty"`
	DomainSW    string `json:"domain_software,omitempty"`
	Maintainer  string `json:"maintainer,omitempty"`
	ODConformance string `json:"od_conformance,omitempty"`
	OSIApproved bool   `json:"osi_approved,omitempty"`
}

// LicenseSearchResult is the paginated response from licenses search.
type LicenseSearchResult struct {
	Hits  LicenseHits `json:"hits"`
	Links Links       `json:"links,omitempty"`
}

// LicenseHits contains the search result hits and total count.
type LicenseHits struct {
	Hits  []License `json:"hits"`
	Total int       `json:"total"`
}
