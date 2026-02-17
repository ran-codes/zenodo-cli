# Zenodo CLI — Requirements & Product Document (v0.1)

## Overview

A Go CLI tool that wraps the Zenodo REST API for metadata management, asset inventory, and BI/reporting across Zenodo communities (organizations). Designed for both human interactive use and agent/automation consumption.

## Goals

1. Manage metadata on existing Zenodo records via CLI
2. List and search assets across multiple communities/organizations
3. Provide BI-oriented queries (counts, filters, aggregate stats)
4. Output machine-readable data (JSON, CSV) for agent consumption
5. Secure credential management following `notion-sync` patterns

## Non-Goals (v0.1)

- File uploads/downloads
- Creating new depositions from scratch
- OAI-PMH harvesting
- Bulk metadata export (`/exporter`)
- Time-series statistics (Zenodo API does not expose per-period stats)
- Funder/grant management (API still in testing)

## Target Users

- **Primary**: The developer (ran-codes) interacting via CLI and AI agents
- **Secondary**: Scripts, cron jobs, CI pipelines for reporting

---

## Functional Requirements

### FR-1: Authentication & Configuration

| ID | Requirement |
|----|-------------|
| FR-1.1 | 3-tier credential resolution: env var (`ZENODO_TOKEN`) > OS keyring > config file |
| FR-1.2 | `config set token <value>` stores token in OS keyring, removes from config file if present |
| FR-1.3 | Auto-migrate plaintext tokens from config file to keyring on startup |
| FR-1.4 | Config file at `$XDG_CONFIG_HOME/zenodo-cli/config.yaml` with `0600` permissions |
| FR-1.5 | Multi-profile support: `--profile <name>` to switch between sandbox/production and multiple accounts |
| FR-1.6 | Default to production Zenodo (`https://zenodo.org/api`). Sandbox via `--sandbox` flag or profile config |
| FR-1.7 | Token masking in all log/error output (show only last 4 chars) |

### FR-2: Records & Search

| ID | Requirement |
|----|-------------|
| FR-2.1 | `records list` — list authenticated user's records and drafts (`GET /api/user/records`) |
| FR-2.2 | `records search <query>` — search all published records with Elasticsearch query syntax |
| FR-2.3 | `records get <id>` — retrieve full record details |
| FR-2.4 | `records versions <id>` — list all versions of a record |
| FR-2.5 | Support `--community <id>` filter on list/search |
| FR-2.6 | Support `--status draft|published` filter on list |
| FR-2.7 | Support `--query` for date range filters (e.g., `publication_date:[2024-01-01 TO 2024-12-31]`) |
| FR-2.8 | Auto-paginate with `--all` flag (respecting rate limits, warn at 10k ceiling) |
| FR-2.9 | Support `--format json|bibtex|datacite|table` on `records get` via Accept headers |

### FR-3: Metadata Editing

| ID | Requirement |
|----|-------------|
| FR-3.1 | `deposit edit <id>` — unlock a published record for editing |
| FR-3.2 | `deposit update <id>` — update metadata via inline flags (`--title`, `--description`, etc.) or `--file metadata.json` or `--stdin` |
| FR-3.3 | Internally always GET current metadata, merge changes, then PUT (never blind PUT) |
| FR-3.4 | Display colored field-by-field diff of changes before confirming |
| FR-3.5 | Prompt for confirmation before re-publishing. `--yes` skips prompt but still prints diff |
| FR-3.6 | `--dry-run` prints diff and exits without touching API |
| FR-3.7 | `deposit discard <id>` — discard unpublished changes |
| FR-3.8 | `deposit publish <id>` — re-publish after metadata edit (with diff + confirmation) |
| FR-3.9 | Client-side metadata validation before sending PUT (required fields based on `upload_type` and `access_right`) |

### FR-4: Communities

| ID | Requirement |
|----|-------------|
| FR-4.1 | `communities list [query]` — search/list communities |
| FR-4.2 | Support `--output json|table|csv` |

### FR-5: Access Links

| ID | Requirement |
|----|-------------|
| FR-5.1 | `access links list <id>` — list share links for a record |
| FR-5.2 | Support `--output json|table|csv` |

### FR-6: Licenses

| ID | Requirement |
|----|-------------|
| FR-6.1 | `licenses search [query]` — search available licenses |
| FR-6.2 | Support `--output json|table|csv` |

---

## Non-Functional Requirements

### NFR-1: Output Formats

| ID | Requirement |
|----|-------------|
| NFR-1.1 | Every command supports `--output json|table|csv` |
| NFR-1.2 | Default to `table` when stdout is a TTY, `json` when piped |
| NFR-1.3 | `--fields` flag to select specific columns (e.g., `--fields id,title,doi,created`) |
| NFR-1.4 | Errors output as structured JSON to stderr when `--output json` is set |
| NFR-1.5 | Consistent exit codes: 0=success, 1=API error, 2=auth error, 3=validation error |

### NFR-2: Rate Limiting

| ID | Requirement |
|----|-------------|
| NFR-2.1 | Built-in rate limiter at HTTP client layer (not per-command) |
| NFR-2.2 | Respect `X-RateLimit-Remaining` and `X-RateLimit-Reset` headers |
| NFR-2.3 | Proactive backoff before hitting limits |
| NFR-2.4 | Log when throttling occurs |
| NFR-2.5 | Search endpoint limit: max 30 req/min |
| NFR-2.6 | General limit: 100 req/min authenticated |

### NFR-3: Security

| ID | Requirement |
|----|-------------|
| NFR-3.1 | Tokens never logged or echoed in full (mask: `****abcd`) |
| NFR-3.2 | Config file written with `0600` permissions; warn if broader |
| NFR-3.3 | TLS certificate validation enforced (never skip verify) |
| NFR-3.4 | Warn users if they pass `--token` flag to prefer env var or keyring instead |

### NFR-4: Cross-Platform

| ID | Requirement |
|----|-------------|
| NFR-4.1 | Single binary for Windows, macOS, Linux |
| NFR-4.2 | OS keyring: Windows Credential Manager, macOS Keychain, Linux Secret Service |
| NFR-4.3 | Graceful fallback to config file if keyring unavailable (with warning) |

---

## Zenodo API Reference (Relevant Subset)

### Base URLs

| Environment | URL |
|-------------|-----|
| Production | `https://zenodo.org/api` |
| Sandbox | `https://sandbox.zenodo.org/api` |

### Authentication

- Bearer token via `Authorization: Bearer <TOKEN>` header
- Scopes needed: `deposit:write` (update metadata), `deposit:actions` (publish/edit/discard)

### Rate Limits

| Tier | Per Minute | Per Hour |
|------|-----------|----------|
| Authenticated (general) | 100 | 5,000 |
| Search endpoints | 30 | — |
| Max results per page | 100 (authenticated) | — |
| Max search depth | 10,000 results | — |

### Key API Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `GET` | `/api/user/records` | List user's own records + drafts |
| `GET` | `/deposit/depositions` | List depositions |
| `GET` | `/deposit/depositions/:id` | Get deposition detail |
| `PUT` | `/deposit/depositions/:id` | Update metadata (full replacement) |
| `POST` | `/deposit/depositions/:id/actions/edit` | Unlock published record |
| `POST` | `/deposit/depositions/:id/actions/publish` | Re-publish after edit |
| `POST` | `/deposit/depositions/:id/actions/discard` | Discard draft changes |
| `GET` | `/records/` | Search published records |
| `GET` | `/records/:id` | Get published record |
| `GET` | `/api/records/:id/versions` | List versions |
| `GET` | `/communities/` | Search communities |
| `GET` | `/licenses/` | Search licenses |
| `GET` | `/api/records/:id/access/links` | List access links |

### Record Stats (Available per Record)

```json
{
  "stats": {
    "downloads": 1234,
    "views": 5678,
    "unique_downloads": 890,
    "unique_views": 3456,
    "version_downloads": 1500,
    "version_views": 7000
  }
}
```

Note: Lifetime totals only. No per-period breakdowns available from API.

### Metadata Fields (Required)

- `upload_type` (publication, poster, presentation, dataset, image, video, software, lesson, physicalobject, other)
- `title`
- `description`
- `creators` (array: `{name, affiliation?, orcid?, gnd?}`)
- `publication_date`
- `access_right` (open, embargoed, restricted, closed)
- `license` (required if access_right is open or embargoed)

### Known Constraints

1. Published records cannot be deleted (permanent DOI)
2. Only one unpublished version at a time
3. Metadata PUT is full replacement (must send all fields)
4. New version ID differs from original record ID (extract from `latest_draft` link)
5. Elasticsearch special characters need escaping in search queries

---

## BI Query Examples

```bash
# How many communities do I manage?
zenodo communities list --output json | jq 'length'

# How many assets in community X?
zenodo records search --community my-org --output json | jq '.total'

# Assets published in 2025 for community X?
zenodo records search --community my-org --query "publication_date:[2025-01-01 TO 2025-12-31]" --output json | jq '.total'

# Download count for asset X?
zenodo records get 12345 --output json | jq '.stats.downloads'

# All assets with download stats across an org (CSV for spreadsheet)
zenodo records search --community my-org --all --fields id,title,stats.downloads,stats.views --output csv
```

---

## Future Considerations (Post v0.1)

- File upload/download support
- Snapshot-based time-series stats (periodic polling to build local history)
- `zenodo report` command that generates summary dashboards
- Community management (if API adds write support)
- Webhook/notification support (if API adds it)
