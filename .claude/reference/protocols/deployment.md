# Deployment

## Matrix Build vs Goreleaser

**Matrix build** (used by notion-sync): You write explicit `go build` commands in a GitHub Actions matrix — one job per OS/arch combo. Simple but verbose. Outputs standalone binaries.

**Goreleaser** (used by zenodo-cli): A declarative tool that reads `.goreleaser.yaml` and handles building, packaging, checksums, and changelogs in one step. Less CI code, supports multi-binary projects natively. By default produces archives (tar.gz/zip), but can output raw binaries with `format: binary`.

zenodo-cli uses goreleaser with `format: binary` so we get the simplicity of goreleaser config with the same standalone-binary UX as notion-sync.

## Feature Comparison

| Feature | notion-sync (matrix) | zenodo-cli (goreleaser) | ELI5 |
|---|---|---|---|
| Build system | Manual `go build` in CI matrix | Goreleaser (declarative yaml) | How the binaries get compiled in CI |
| Release artifacts | Standalone binaries (`name-os-arch`) | Standalone binaries (via `format: binary`) | What users actually download |
| macOS/Linux install | `curl \| bash` install script | `curl \| bash` install script | One-liner to install on Mac/Linux |
| Windows install | Scoop (points at `.exe`) | Scoop (points at `.exe`) | One-liner to install on Windows |
| Checksums | `sha256sum` in CI step | Goreleaser generates `checksums.txt` | Verifies download wasn't corrupted |
| Version injection | `-X main.version=` in CI | `-X internal/cli.version=` via ldflags | Embeds version number into the binary at build time |
| Scoop manifest update | CI step with `jq` | CI step with `jq` | Auto-updates the Scoop package after release |
| Multi-binary support | Separate build line per binary | Multiple `builds:` entries in yaml | How we ship both `zenodo` and `zenodo-mcp` |
| Changelog | GitHub auto-generated | Goreleaser with commit filters | Release notes on the GitHub release page |

## Checklist

### 1. Pre-Release — Feature Readiness

- [ ] All features intended for this release are merged to `main`
- [ ] `go build ./...` and `go test ./...` pass on `main`
- [ ] Rebuild binaries and smoke-test key commands

### 2. Pre-Release — CI/CD Readiness

- [ ] `.github/workflows/release.yml` exists and triggers on `v*` tags
- [ ] `.goreleaser.yaml` is configured (platforms, ldflags, format: binary)
- [ ] `bucket/zenodo-cli.json` Scoop manifest exists
- [ ] Release workflow includes Scoop manifest auto-update step
- [ ] `scripts/install.sh` exists for macOS/Linux
- [ ] Version injection works (`internal/cli/version.go` has ldflags vars)

### 3. Tag & Release

- [ ] Ensure you're on `main` with a clean working tree
- [ ] Tag: `git tag v<VERSION>` (e.g. `git tag v0.1.0`)
- [ ] Push tag: `git push origin v<VERSION>`
- [ ] Monitor the GitHub Actions release workflow for success

### 4. Post-Release Verification

- [ ] GitHub release page has binaries for all platforms (linux/darwin/windows × amd64/arm64)
- [ ] `checksums.txt` is included in the release assets
- [ ] `bucket/zenodo-cli.json` was auto-updated with new version and hash
- [ ] Scoop install works: `scoop bucket add zenodo-cli https://github.com/ran-codes/zenodo-cli && scoop install zenodo-cli`
- [ ] macOS install works: `curl -fsSL ... | bash`
- [ ] `zenodo version` shows correct version, commit, and date
