# IPv4 vs IPv6 Connection Bug

## ELI5

Your computer tries to call Zenodo using two phone lines: IPv6 (the new one) and IPv4 (the old one). The new line is advertised but broken on your network. Go's HTTP client tries the new line first, waits 30 seconds for it to fail, then tries the old line. We told it to skip the broken line and use IPv4 directly.

## Problem

Go's default `http.Transport` uses dual-stack dialing — it tries both IPv6 and IPv4 when connecting. On networks where IPv6 is advertised but not functional (common on Windows, MSYS/Git Bash, some corporate networks), every HTTP request to `zenodo.org` would:

1. Attempt IPv6 connection
2. Hang for the full dial timeout (~30s)
3. Fall back to IPv4
4. Finally succeed

This made every API call take 30+ seconds, and with the default 30s client timeout, most requests just failed with `context deadline exceeded`.

## Symptoms

- `./zenodo.exe records list` hangs for 30 seconds then errors
- Error: `context deadline exceeded (Client.Timeout exceeded while awaiting headers)`
- Only happens on certain networks/machines — works fine where IPv6 is functional

## Solution

Force IPv4 in the HTTP transport in `internal/api/client.go`:

```go
httpClient: &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
            d := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
            return d.DialContext(ctx, "tcp4", addr)
        },
        ForceAttemptHTTP2: true,
    },
},
```

## Trade-off

This disables IPv6 entirely. If a user is on an IPv6-only network, the CLI won't work. This is unlikely for Zenodo users today, but if it becomes an issue, a better fix would be Happy Eyeballs (RFC 6555) — try both in parallel, use whichever connects first. Go's `net.Dialer` supports this natively but the default timeouts are too generous for broken networks.
