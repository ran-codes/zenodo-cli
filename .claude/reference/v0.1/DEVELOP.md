# Developer Guide — zenodo-cli v0.1

## Prerequisites

- Go 1.24+ (module uses 1.25.7, but 1.24+ should work)
- Git
- A Zenodo API token (production or sandbox)

## Clone & Build

```bash
git clone https://github.com/ran-codes/zenodo-cli.git
cd zenodo-cli
go build -o zenodo ./cmd/zenodo/
```

On Windows the binary will be `zenodo.exe`:

```bash
go build -o zenodo.exe ./cmd/zenodo/
```

## Run Tests

```bash
# Unit tests (no network, no token needed)
go test ./...

# Integration tests (requires ZENODO_SANDBOX_TOKEN)
ZENODO_SANDBOX_TOKEN=<token> go test ./test/integration/ -tags=integration -v -timeout 120s
```

## First-Time Setup

```bash
# Store your API token (goes to OS keyring)
./zenodo config set token <your-token>

# Verify it works
./zenodo records search "test"
```

### Multiple Profiles

```bash
# Set up a sandbox profile
./zenodo config set profiles.sandbox.base_url https://sandbox.zenodo.org/api
./zenodo config use sandbox
./zenodo config set token <sandbox-token>

# Switch back to production
./zenodo config use production

# List profiles (* = active)
./zenodo config profiles
```

## Project Structure

```
cmd/zenodo/main.go          Entry point, exit code mapping, JSON error output
internal/
  cli/
    root.go                 Root cobra command, PersistentPreRunE (config, auth, output)
    context.go              AppContext struct (shared runtime state)
    config.go               config set/get/profiles/use commands
    records.go              records list/search/get/versions commands
    communities.go          communities list command
    licenses.go             licenses search command
    access.go               access links list command (disabled v0.1)
    deposit.go              deposit edit/update/discard/publish commands
    completion.go           Shell completion generation
    version.go              Version command (ldflags-injected)
  api/
    client.go               HTTP client with auth, rate limiting, error parsing
    ratelimit.go            Dual token bucket rate limiter
    records.go              Records API methods
    communities.go          Communities API methods
    licenses.go             Licenses API methods
    access.go               Access links API methods (disabled v0.1)
    depositions.go          Depositions API methods (edit/update/publish/discard)
  config/
    config.go               Viper-backed config with profiles
    paths.go                XDG config paths (cross-platform)
    keyring.go              OS keyring for token storage
  model/
    record.go               Record, Deposition, Metadata structs
    community.go            Community structs
    license.go              License structs
    access.go               Access link structs (unused v0.1)
    stats.go                Stats struct
    error.go                APIError with hints
  output/
    formatter.go            Output dispatcher (json/table/csv), field selection
    json.go                 JSON formatter
    table.go                Table formatter
    csv.go                  CSV formatter
    diff.go                 Metadata diff display (colored)
  validate/
    metadata.go             Metadata validation rules
test/
  integration/
    records_test.go         Sandbox integration tests
```

## Key Design Decisions

- **Token resolution chain**: `--token` flag > `ZENODO_TOKEN` env var > OS keyring > config file
- **Output auto-detection**: TTY gets `table`, piped output gets `json`
- **Rate limiting**: Dual token bucket (100 req/min general, 30 req/min search) with header-based updates
- **Metadata updates**: GET-merge-PUT pattern — never blind PUT. Fetches current, merges changes, shows diff, confirms
- **Exit codes**: 0=success, 1=API error, 2=auth error, 3=validation, 4=rate limit, 5=user cancelled

## Adding a New Command

1. Create `internal/cli/<command>.go` with cobra command
2. Add API method in `internal/api/<resource>.go` if needed
3. Add model structs in `internal/model/` if needed
4. Register command in `init()` function
5. Run `go build ./...` and `go test ./...`

## Release Process

Releases are automated via goreleaser. To create a release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

This triggers `.github/workflows/release.yml` which builds cross-platform binaries (linux/darwin/windows x amd64/arm64) and creates a GitHub release.

## Dependencies

| Package | Purpose |
|---------|---------|
| `spf13/cobra` | CLI framework |
| `spf13/viper` | Config management |
| `zalando/go-keyring` | OS keyring for token storage |
| `olekukonko/tablewriter` v0.0.5 | Table output formatting |
| `mattn/go-isatty` | TTY detection for output auto-switching |
| `r3labs/diff/v3` | Structural metadata diffing |
| `fatih/color` | Colored terminal output |
