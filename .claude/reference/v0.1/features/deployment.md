# Deployment — CI/CD Design

## Problem

zenodo-cli ships two binaries (`zenodo` + `zenodo-mcp`) and needs to support three install methods:

1. **macOS/Linux**: `curl | bash` one-liner
2. **Windows**: Scoop package manager
3. **Manual**: Download from GitHub Releases

The original goreleaser config produced tar.gz/zip archives. This made `curl | bash` impractical (download archive → extract → find binary → move to PATH) and diverged from the pattern established in notion-sync.

## Options Considered

### Option A: Matrix build (notion-sync pattern)

Manual `go build` commands in a GitHub Actions matrix — one job per OS/arch.

```yaml
strategy:
  matrix:
    include:
      - goos: darwin
        goarch: arm64
      - goos: linux
        goarch: amd64
      # ...
steps:
  - run: go build -o zenodo-${{ matrix.goos }}-${{ matrix.goarch }} ./cmd/zenodo
  - run: go build -o zenodo-mcp-${{ matrix.goos }}-${{ matrix.goarch }} ./cmd/zenodo-mcp
```

**Pros**: Simple, explicit, identical to notion-sync.
**Cons**: Verbose CI config. Every new binary = more lines. No auto-changelog. Must manually handle checksums, ldflags, etc.

### Option B: Goreleaser with archives (original)

Goreleaser bundles both binaries into tar.gz (Unix) / zip (Windows) per platform.

**Pros**: Minimal config. Multi-binary is one yaml entry.
**Cons**: Archives break `curl | bash`. Users must extract before using. Diverges from notion-sync UX.

### Option C: Goreleaser with `format: binary` (chosen)

Goreleaser outputs standalone binaries (no archive wrapping). Names follow `{name}-{os}-{arch}` pattern.

**Pros**: Declarative config. Multi-binary support. Auto-changelog. Checksums. Same standalone-binary UX as notion-sync.
**Cons**: Slightly different from notion-sync's CI internals (goreleaser vs matrix), but identical end result.

## Implementation

### Goreleaser config (`.goreleaser.yaml`)

Two `builds:` entries — one for `zenodo`, one for `zenodo-mcp`. Both use `binary: {name}-{{ .Os }}-{{ .Arch }}` naming. Archives section set to `format: binary` so no tar.gz/zip wrapping.

### Release workflow (`.github/workflows/release.yml`)

Triggered on `v*` tags. Runs goreleaser, then updates the Scoop manifest with checksums for both `.exe` files and commits to main.

### Install script (`scripts/install.sh`)

Detects OS/arch, downloads both `zenodo` and `zenodo-mcp` binaries from the latest GitHub release, installs to `/usr/local/bin`. Uses `sudo` only if needed.

### Scoop manifest (`bucket/zenodo-cli.json`)

Downloads both `.exe` files. Uses an `installer` script to rename from `zenodo-windows-amd64.exe` → `zenodo.exe` (and same for mcp). Both listed in `bin` array.

## Release artifacts

Each release produces:

| File | Description |
|------|-------------|
| `zenodo-darwin-amd64` | CLI — macOS Intel |
| `zenodo-darwin-arm64` | CLI — macOS Apple Silicon |
| `zenodo-linux-amd64` | CLI — Linux x64 |
| `zenodo-linux-arm64` | CLI — Linux ARM |
| `zenodo-windows-amd64.exe` | CLI — Windows x64 |
| `zenodo-windows-arm64.exe` | CLI — Windows ARM |
| `zenodo-mcp-darwin-amd64` | MCP server — macOS Intel |
| `zenodo-mcp-darwin-arm64` | MCP server — macOS Apple Silicon |
| `zenodo-mcp-linux-amd64` | MCP server — Linux x64 |
| `zenodo-mcp-linux-arm64` | MCP server — Linux ARM |
| `zenodo-mcp-windows-amd64.exe` | MCP server — Windows x64 |
| `zenodo-mcp-windows-arm64.exe` | MCP server — Windows ARM |
| `checksums.txt` | SHA-256 checksums for all files |
