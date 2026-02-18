# internal/model

Data structs matching Zenodo API responses.

- `record.go` — `Record`, `Deposition`, `Metadata`, `Creator`, `Contributor`, `Links`, search result wrappers.
- `License` field in `Metadata` is `json.RawMessage` because the API returns either a string or `{"id": "cc-by-4.0"}`. Use `LicenseString()` to extract.
- `community.go` — `Community` has nested `CommunityMetadata` (title is inside `metadata`, not top-level).
- `license.go` — `License.Title` is `map[string]string` (localized, e.g. `{"en": "Creative Commons..."}`). Use `TitleString()` to extract.
- `error.go` — `APIError` with `Hint()` for user-friendly messages on 401/403/404/429.
- `conceptrecid` fields are strings, not ints.
