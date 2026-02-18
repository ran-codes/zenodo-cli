# internal/validate

Metadata validation before submitting to the API.

- `metadata.go` — `Metadata()` validates required fields (title, description, upload_type, publication_date, access_right, creators). Conditional validation: license required for open/embargoed, embargo_date for embargoed, access_conditions for restricted.
- Uses `LicenseString()` from model package since license field is `json.RawMessage`.
