# Integrated Testing Gap

## ELI5

We can write tests that check "if the API returns X, does the CLI do Y?" But we can't check "does the API *actually* return X?" In this conversation, the mock was wrong because the assumption about the API response shape was wrong. `ListUserRecords` returns `[]Deposition` from `/deposit/depositions`, not a `RecordSearchResult` from `/records`. The unit test passed with a fake response that didn't match reality.

Unit tests with mocks are like practicing with a cardboard cutout of your opponent. Useful, but if the cutout doesn't look like the real opponent, you're practicing wrong.

## Options

1. **Unit tests with mocks** — Fast, no credentials needed, but only as good as our assumptions about the API. Won't catch "wrong endpoint" or "wrong response shape" bugs. This is what we have today.

2. **Sandbox integration tests** — Real API calls against `sandbox.zenodo.org`. Catches real bugs but needs a token and is slower/flakier (e.g. 502 errors when Zenodo is down).

3. **Recorded fixtures** — Record real API responses once, replay them in tests. Best of both worlds but fixtures go stale if the API changes.

## Proposal

Use **sandbox integration tests** gated behind an env var (`ZENODO_SANDBOX_TOKEN`). When the token is set, tests hit the real sandbox API. When it's not, they're skipped.

- CI can run them if the token is added as a GitHub secret
- Developers can run them locally by storing a sandbox token
- They skip gracefully when no token is available
- Pair with existing unit tests (mocks) for fast feedback on logic

### Example pattern

```go
func TestListUserRecords_Integration(t *testing.T) {
    token := os.Getenv("ZENODO_SANDBOX_TOKEN")
    if token == "" {
        t.Skip("ZENODO_SANDBOX_TOKEN not set, skipping integration test")
    }

    client := api.NewClient("https://sandbox.zenodo.org/api", token)
    result, err := client.ListUserRecords(api.RecordListParams{})
    if err != nil {
        t.Fatalf("ListUserRecords() error: %v", err)
    }
    // Verify shape, not specific data
    t.Logf("Got %d depositions", len(result))
}
```

### Setup

1. Create account at https://sandbox.zenodo.org
2. Generate a personal access token with `deposit:write` and `deposit:actions` scopes
3. Set `ZENODO_SANDBOX_TOKEN=<token>` in your environment
4. Run `go test ./... -tags=integration`
