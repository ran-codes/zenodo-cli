# internal/api

HTTP client and Zenodo API methods.

- `client.go` — `Client` struct with `Get/Post/Put/Delete/GetRaw`. Handles auth header, JSON encode/decode, error parsing, rate limiter integration.
- `ratelimit.go` — Dual token bucket: 100 req/min general, 30 req/min for search endpoints. Updates from `X-RateLimit-*` response headers.
- API method files (`records.go`, `communities.go`, etc.) take typed params and return model structs.
- Paths are relative to base URL (e.g. `/records`, `/communities`). Base URL already includes `/api`. No trailing slashes.
- The real Zenodo API returns some fields as objects where you might expect strings (e.g. `license`, `resource_type`, community `title`). See model package.
