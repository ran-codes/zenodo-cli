# Zenodo CLI — Implementation Plan (v0.1)

## Tech Stack

| Component | Choice | Justification |
|-----------|--------|---------------|
| Language | Go 1.24+ | Single binary, cross-platform, notion-sync precedent |
| CLI framework | `spf13/cobra` | Nested subcommands, auto-completions, flag handling |
| Config | `spf13/viper` | Layered config: flags > env > file > defaults |
| HTTP client | `net/http` (stdlib) | JSON-only operations, no streaming needed |
| Credential storage | `zalando/go-keyring` | Cross-platform OS keyring, simple API |
| Table output | `olekukonez/tablewriter` | ASCII table rendering for terminal |
| Diff display | `r3labs/diff` | Structured JSON object diffing for metadata preview |
| Colored output | `fatih/color` | Colored terminal diff (red/green) |
| Logging | `log/slog` (stdlib) | Structured logging, Go 1.21+ |
| Testing | `net/http/httptest` (stdlib) | Mock HTTP server for API tests |

## Project Structure

```
zenodo-cli/
├── cmd/
│   └── zenodo/
│       └── main.go                 # Entry point
├── internal/
│   ├── cli/
│   │   ├── root.go                 # Root cobra command, global flags
│   │   ├── config.go               # config set/get/profiles commands
│   │   ├── records.go              # records list/search/get/versions commands
│   │   ├── deposit.go              # deposit edit/update/discard/publish commands
│   │   ├── communities.go          # communities list command
│   │   ├── licenses.go             # licenses search command
│   │   └── access.go               # access links list command
│   ├── api/
│   │   ├── client.go               # HTTP client with auth, rate limiting, base URL
│   │   ├── ratelimit.go            # Rate limiter (token bucket + header parsing)
│   │   ├── records.go              # Records API methods
│   │   ├── depositions.go          # Depositions API methods
│   │   ├── communities.go          # Communities API methods
│   │   ├── licenses.go             # Licenses API methods
│   │   └── access.go               # Access links API methods
│   ├── config/
│   │   ├── config.go               # Config loading (viper), profile management
│   │   ├── keyring.go              # Keyring get/set/delete/migrate
│   │   └── paths.go                # XDG-compliant config path resolution
│   ├── model/
│   │   ├── record.go               # Record, Deposition, Metadata structs
│   │   ├── community.go            # Community struct
│   │   ├── license.go              # License struct
│   │   ├── stats.go                # Stats struct
│   │   └── error.go                # API error response struct
│   ├── output/
│   │   ├── formatter.go            # Interface: Format(data, format, fields) string
│   │   ├── table.go                # Table formatter
│   │   ├── json.go                 # JSON formatter
│   │   ├── csv.go                  # CSV formatter
│   │   └── diff.go                 # Metadata diff display (colored)
│   └── validate/
│       └── metadata.go             # Client-side metadata validation
├── go.mod
├── go.sum
├── .goreleaser.yaml                # Cross-platform release builds
└── .claude/
    └── reference/
        └── v0.1/
            ├── RPD.md
            └── PLAN.md
```

## Implementation Phases

### Phase 1: Foundation

**Goal**: Runnable CLI skeleton with auth and config.

1. Initialize Go module (`github.com/ran-codes/zenodo-cli`)
2. Set up cobra root command with global flags:
   - `--token` (with warning to prefer env/keyring)
   - `--profile`
   - `--sandbox`
   - `--output json|table|csv`
   - `--fields`
   - `--verbose`
3. Implement `internal/config/`:
   - `config.go` — viper setup, profile loading, `LoadConfig()`, `SaveConfig()`
   - `keyring.go` — `GetToken()`, `SetToken()`, `DeleteToken()`, `MigrateToken()`
   - `paths.go` — `GetConfigDir()` with XDG support
4. Implement `config set` and `config get` commands
5. Implement `internal/api/client.go`:
   - `NewClient(baseURL, token)` constructor
   - Auth header injection
   - JSON request/response helpers: `Get()`, `Post()`, `Put()`, `Delete()`
   - Error response parsing into structured `model.APIError`
6. Implement `internal/api/ratelimit.go`:
   - Token bucket rate limiter (100 req/min general, 30 req/min search)
   - Parse `X-RateLimit-Remaining` / `X-RateLimit-Reset` headers
   - Proactive sleep when approaching limits
   - Logging when throttling occurs

**Tests**: Config loading/saving, keyring mock, client auth injection, rate limiter behavior.

### Phase 2: Read Operations

**Goal**: List, search, and get records — the core BI functionality.

1. Define model structs (`model/record.go`, `model/community.go`, etc.)
2. Implement `internal/output/formatter.go` — interface + auto-detect TTY for default format
3. Implement table, JSON, CSV formatters with `--fields` support
4. Implement API methods:
   - `api/records.go` — `ListUserRecords()`, `SearchRecords()`, `GetRecord()`, `ListVersions()`
   - `api/communities.go` — `SearchCommunities()`
   - `api/licenses.go` — `SearchLicenses()`
   - `api/access.go` — `ListAccessLinks()`
5. Implement CLI commands:
   - `records list` (with `--status`, `--community`, `--all` pagination)
   - `records search <query>` (with `--community`, `--all`)
   - `records get <id>` (with `--format json|bibtex|datacite|table`)
   - `records versions <id>`
   - `communities list [query]`
   - `licenses search [query]`
   - `access links list <id>`
6. Implement `--all` auto-pagination:
   - Loop through pages until exhausted or 10k ceiling
   - Rate-limit-aware delays between pages
   - Warn when results are truncated at 10k

**Tests**: Mock API responses for each endpoint, pagination logic, output formatting, field selection.

### Phase 3: Metadata Editing

**Goal**: Safe metadata updates with diff preview.

1. Implement `api/depositions.go`:
   - `GetDeposition()`, `UpdateDeposition()`, `EditDeposition()`, `PublishDeposition()`, `DiscardDeposition()`
2. Implement `internal/output/diff.go`:
   - Accept old and new metadata structs
   - Produce field-by-field colored diff output
   - Handle nested objects (creators, related_identifiers)
3. Implement `internal/validate/metadata.go`:
   - Validate required fields based on `upload_type` and `access_right`
   - Return clear error messages for missing/invalid fields
4. Implement GET-merge-PUT logic in `deposit update`:
   - GET current metadata
   - Parse user's changes (from flags, `--file`, or `--stdin`)
   - Merge into current metadata (only overwrite specified fields)
   - Validate merged result
   - Display diff
   - Prompt for confirmation (unless `--yes`)
   - PUT merged metadata
   - If not `--dry-run`, call publish action
5. Implement CLI commands:
   - `deposit edit <id>` — unlock for editing
   - `deposit update <id>` — the full update flow with diff
   - `deposit discard <id>` — discard changes
   - `deposit publish <id>` — re-publish (with diff + confirmation)

**Tests**: Merge logic (partial updates, nested fields), diff output, validation rules, dry-run behavior.

### Phase 4: Polish & Release

**Goal**: Production-ready quality.

1. Shell completions (cobra generates these):
   - `zenodo completion bash|zsh|fish|powershell`
2. `--version` flag with build info (git commit, build date)
3. `.goreleaser.yaml` for cross-platform builds
4. Error handling audit:
   - All API errors surfaced with status code + message
   - Network errors distinguished from API errors
   - Auth errors (401/403) suggest checking token/scopes
5. Integration tests against Zenodo sandbox
6. `config profiles` command for listing/switching profiles
7. Man page / help text polish

---

## Key Design Decisions

### D-1: GET-Merge-PUT for Metadata Updates

Zenodo's PUT endpoint is **full replacement**. The CLI must:
1. `GET /deposit/depositions/:id` to fetch current metadata
2. Deep-merge user's partial changes into the full metadata object
3. `PUT` the merged result

This prevents accidental field deletion. The merge logic must handle:
- Top-level scalar fields (title, description) — simple overwrite
- Array fields (creators, keywords) — replace entire array if user specifies it
- Nested objects — merge recursively

### D-2: Output Format Auto-Detection

```go
func DefaultFormat() string {
    if isatty.IsTerminal(os.Stdout.Fd()) {
        return "table"
    }
    return "json"
}
```

When piped to another command or agent, output defaults to JSON. Interactive terminal gets tables.

### D-3: Rate Limiting Architecture

Rate limiting lives in the HTTP client layer, not in individual commands. Every API call passes through the rate limiter automatically.

```
Command → api.Client.Get() → rateLimiter.Wait() → http.Client.Do() → parse response headers → update rateLimiter state
```

Two buckets:
- General: 100 req/min
- Search: 30 req/min (for `/records/`, `/communities/`, `/licenses/` endpoints)

### D-4: Profile Storage

```yaml
# ~/.config/zenodo-cli/config.yaml
default_profile: production

profiles:
  production:
    base_url: https://zenodo.org/api
    # token stored in keyring as "zenodo-cli:production"
  sandbox:
    base_url: https://sandbox.zenodo.org/api
    # token stored in keyring as "zenodo-cli:sandbox"
  work:
    base_url: https://zenodo.org/api
    # token stored in keyring as "zenodo-cli:work"
```

Each profile gets its own keyring entry. `--profile` flag or `ZENODO_PROFILE` env var selects the active profile.

### D-5: Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | API error (4xx/5xx from Zenodo) |
| 2 | Authentication error (no token, invalid token, insufficient scopes) |
| 3 | Validation error (invalid metadata, missing required fields) |
| 4 | Rate limit exceeded (after retries exhausted) |
| 5 | User cancelled (declined confirmation prompt) |

---

## Dependencies (Final List)

```
github.com/spf13/cobra          # CLI framework
github.com/spf13/viper          # Configuration
github.com/zalando/go-keyring   # OS keyring
github.com/olekukonez/tablewriter # Table output
github.com/r3labs/diff/v3       # Structured object diff
github.com/fatih/color          # Colored terminal output
github.com/mattn/go-isatty      # TTY detection for output format auto-detect
```

7 dependencies total. All well-maintained, widely used in the Go ecosystem.

---

## Risk Mitigations

| Risk | Mitigation |
|------|------------|
| Metadata PUT wipes fields | GET-merge-PUT with diff preview (D-1) |
| Search result truncation at 10k | Warn user, suggest narrower query |
| Rate limiting on batch operations | Client-layer rate limiter with proactive backoff (D-3) |
| Keyring unavailable (headless/CI) | Graceful fallback to config file with warning |
| Zenodo API changes (InvenioRDM migration) | Target stable legacy endpoints first; add InvenioRDM endpoints as they stabilize |
| Token leaked in logs/output | Mask to `****<last4>` in all log/error paths |
