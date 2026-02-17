package model

// License represents a Zenodo license.
type License struct {
	ID    string            `json:"id"`
	Title map[string]string `json:"title,omitempty"`
	Props LicenseProps      `json:"props,omitempty"`
	Tags  []string          `json:"tags,omitempty"`
}

// LicenseProps contains extra properties for a license.
type LicenseProps struct {
	URL         string `json:"url,omitempty"`
	Scheme      string `json:"scheme,omitempty"`
	OSIApproved string `json:"osi_approved,omitempty"`
}

// TitleString returns the English title, falling back to the first available.
func (l *License) TitleString() string {
	if t, ok := l.Title["en"]; ok {
		return t
	}
	for _, t := range l.Title {
		return t
	}
	return ""
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
