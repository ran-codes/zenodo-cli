# internal/cli

Cobra command definitions. Each file is one command group (records, deposit, config, etc.).

- `root.go` — Root command with `PersistentPreRunE` that resolves config, profile, token, base URL, and output format into `appCtx`.
- `context.go` — `AppContext` struct shared across all commands.
- Commands create an `api.Client` from `appCtx.BaseURL` and `appCtx.Token`, call API methods, then pipe results through `output.Format()`.
- Deposit commands use GET-merge-PUT pattern: fetch current metadata, merge changes, show diff, confirm before writing.
- New commands: create file, add cobra command, register in `init()`, use `appCtx` for resolved state.
