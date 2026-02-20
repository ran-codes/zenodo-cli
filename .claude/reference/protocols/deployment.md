# Deployment Checklist

## 1. Pre-Release — Feature Readiness

- [ ] All features intended for this release are merged to `main`
- [ ] Any commands not ready are commented out (e.g. deposit write, access links in v0.1)
- [ ] `go build ./...` and `go test ./...` pass on `main`
- [ ] Rebuild binary (`go build -o zenodo.exe ./cmd/zenodo`) and smoke-test key commands

## 2. Pre-Release — CI/CD Readiness

- [ ] `.github/workflows/release.yml` exists and triggers on `v*` tags
- [ ] `.goreleaser.yaml` is configured (platforms, ldflags, archives, checksums)
- [ ] `bucket/zenodo.json` Scoop manifest exists
- [ ] Release workflow includes Scoop manifest auto-update step
- [ ] Version injection works (`internal/cli/version.go` has ldflags vars)

## 3. Tag & Release

- [ ] Ensure you're on `main` with a clean working tree
- [ ] Tag: `git tag v<VERSION>` (e.g. `git tag v0.1.0`)
- [ ] Push tag: `git push origin v<VERSION>`
- [ ] Monitor the GitHub Actions release workflow for success

## 4. Post-Release Verification

- [ ] GitHub release page has binaries for all platforms (linux/darwin/windows × amd64/arm64)
- [ ] `checksums.txt` is included in the release assets
- [ ] `bucket/zenodo.json` was auto-updated with new version and hash (check the commit on `main`)
- [ ] Scoop install works: `scoop bucket add zenodo https://github.com/ran-codes/zenodo-cli && scoop install zenodo`
- [ ] `zenodo version` shows correct version, commit, and date
