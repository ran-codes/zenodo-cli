package model

// Stats holds download and view statistics for a record.
type Stats struct {
	Downloads       int `json:"downloads"`
	Views           int `json:"views"`
	UniqueDownloads int `json:"unique_downloads"`
	UniqueViews     int `json:"unique_views"`
	VersionDownloads int `json:"version_downloads"`
	VersionViews    int `json:"version_views"`
}
