package model

import (
	"encoding/json"
	"time"
)

// Record represents a Zenodo published record.
type Record struct {
	ID          int       `json:"id"`
	ConceptID   string    `json:"conceptrecid,omitempty"`
	DOI         string    `json:"doi,omitempty"`
	ConceptDOI  string    `json:"conceptdoi,omitempty"`
	Title       string    `json:"title,omitempty"`
	Metadata    Metadata  `json:"metadata"`
	Stats       Stats     `json:"stats,omitempty"`
	Links       Links     `json:"links,omitempty"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Revision    int       `json:"revision"`
	State       string    `json:"state,omitempty"`
	Submitted   bool      `json:"submitted,omitempty"`
}

// Deposition represents a Zenodo deposition (draft or published).
type Deposition struct {
	ID        int       `json:"id"`
	ConceptID string    `json:"conceptrecid,omitempty"`
	DOI       string    `json:"doi,omitempty"`
	DOIURL    string    `json:"doi_url,omitempty"`
	Title     string    `json:"title,omitempty"`
	Metadata  Metadata  `json:"metadata"`
	Links     Links     `json:"links,omitempty"`
	State     string    `json:"state,omitempty"`
	Submitted bool      `json:"submitted"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
}

// Metadata contains the bibliographic metadata for a record or deposition.
type Metadata struct {
	Title               string              `json:"title,omitempty"`
	Description         string              `json:"description,omitempty"`
	UploadType          string              `json:"upload_type,omitempty"`
	PublicationType     string              `json:"publication_type,omitempty"`
	ImageType           string              `json:"image_type,omitempty"`
	PublicationDate     string              `json:"publication_date,omitempty"`
	AccessRight         string              `json:"access_right,omitempty"`
	License             json.RawMessage     `json:"license,omitempty"`
	ResourceType        *ResourceType       `json:"resource_type,omitempty"`
	EmbargoDate         string              `json:"embargo_date,omitempty"`
	AccessConditions    string              `json:"access_conditions,omitempty"`
	DOI                 string              `json:"doi,omitempty"`
	PrereserveDOI       *PrereserveDOI      `json:"prereserve_doi,omitempty"`
	Keywords            []string            `json:"keywords,omitempty"`
	Notes               string              `json:"notes,omitempty"`
	Version             string              `json:"version,omitempty"`
	Language            string              `json:"language,omitempty"`
	Creators            []Creator           `json:"creators,omitempty"`
	Contributors        []Contributor       `json:"contributors,omitempty"`
	RelatedIdentifiers  []RelatedIdentifier `json:"related_identifiers,omitempty"`
	Communities         []CommunityRef      `json:"communities,omitempty"`
	Grants              []Grant             `json:"grants,omitempty"`
	JournalTitle        string              `json:"journal_title,omitempty"`
	JournalVolume       string              `json:"journal_volume,omitempty"`
	JournalIssue        string              `json:"journal_issue,omitempty"`
	JournalPages        string              `json:"journal_pages,omitempty"`
	ConferenceTitle     string              `json:"conference_title,omitempty"`
	ConferenceAcronym   string              `json:"conference_acronym,omitempty"`
	ConferenceURL       string              `json:"conference_url,omitempty"`
	ConferenceDates     string              `json:"conference_dates,omitempty"`
	ConferencePlace     string              `json:"conference_place,omitempty"`
	ConferenceSession   string              `json:"conference_session,omitempty"`
	ConferenceSessionPart string            `json:"conference_session_part,omitempty"`
	ImprintISBN         string              `json:"imprint_isbn,omitempty"`
	ImprintPlace        string              `json:"imprint_place,omitempty"`
	ImprintPublisher    string              `json:"imprint_publisher,omitempty"`
	PartOfTitle         string              `json:"partof_title,omitempty"`
	PartOfPages         string              `json:"partof_pages,omitempty"`
	ThesisSupervisors   []Creator           `json:"thesis_supervisors,omitempty"`
	ThesisUniversity    string              `json:"thesis_university,omitempty"`
	Subjects            []Subject           `json:"subjects,omitempty"`
	References          []string            `json:"references,omitempty"`
}

// Creator represents a record creator.
type Creator struct {
	Name        string `json:"name"`
	Affiliation string `json:"affiliation,omitempty"`
	ORCID       string `json:"orcid,omitempty"`
	GND         string `json:"gnd,omitempty"`
}

// Contributor represents a record contributor.
type Contributor struct {
	Name        string `json:"name"`
	Affiliation string `json:"affiliation,omitempty"`
	ORCID       string `json:"orcid,omitempty"`
	GND         string `json:"gnd,omitempty"`
	Type        string `json:"type"`
}

// RelatedIdentifier links to related resources.
type RelatedIdentifier struct {
	Identifier string `json:"identifier"`
	Relation   string `json:"relation"`
	Scheme     string `json:"scheme,omitempty"`
}

// CommunityRef is a reference to a community.
// The API returns "identifier" (depositions) or "id" (records).
type CommunityRef struct {
	Identifier string `json:"identifier"`
	ID         string `json:"id"`
}

// Slug returns the community identifier, preferring "identifier" over "id".
func (c CommunityRef) Slug() string {
	if c.Identifier != "" {
		return c.Identifier
	}
	return c.ID
}

// Grant represents a funding grant.
type Grant struct {
	ID string `json:"id,omitempty"`
}

// PrereserveDOI holds a pre-reserved DOI.
type PrereserveDOI struct {
	DOI  string `json:"doi"`
	ID   interface{} `json:"recid"`
}

// ResourceType represents the resource type from the API.
type ResourceType struct {
	Title string `json:"title,omitempty"`
	Type  string `json:"type,omitempty"`
}

// LicenseString extracts a license identifier from the License field,
// which may be a plain string or an object like {"id": "cc-by-4.0"}.
func (m *Metadata) LicenseString() string {
	if m.License == nil {
		return ""
	}
	// Try plain string first.
	var s string
	if err := json.Unmarshal(m.License, &s); err == nil {
		return s
	}
	// Try object with "id" field.
	var obj struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(m.License, &obj); err == nil {
		return obj.ID
	}
	return string(m.License)
}

// Subject represents a subject classification.
type Subject struct {
	Term       string `json:"term"`
	Identifier string `json:"identifier"`
	Scheme     string `json:"scheme"`
}

// Links contains HATEOAS links from the API.
type Links struct {
	Self         string `json:"self,omitempty"`
	HTML         string `json:"html,omitempty"`
	DOI          string `json:"doi,omitempty"`
	Bucket       string `json:"bucket,omitempty"`
	Publish      string `json:"publish,omitempty"`
	Edit         string `json:"edit,omitempty"`
	Discard      string `json:"discard,omitempty"`
	NewVersion   string `json:"newversion,omitempty"`
	LatestDraft  string `json:"latest_draft,omitempty"`
	Latest       string `json:"latest,omitempty"`
	LatestHTML   string `json:"latest_html,omitempty"`
	Versions     string `json:"versions,omitempty"`
	AccessLinks  string `json:"access_links,omitempty"`
}

// RecordSearchResult is the paginated response from search endpoints.
type RecordSearchResult struct {
	Hits  RecordHits `json:"hits"`
	Links Links      `json:"links,omitempty"`
}

// RecordHits contains the search result hits and total count.
type RecordHits struct {
	Hits  []Record `json:"hits"`
	Total int      `json:"total"`
}

// VersionList is the response from the versions endpoint.
type VersionList struct {
	Hits []Record `json:"hits"`
}
