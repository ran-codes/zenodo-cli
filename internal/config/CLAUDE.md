# internal/config

Configuration management with Viper, OS keyring, and XDG paths.

- `config.go` — Viper-backed config with `Load/Save`, profile management (`DefaultProfile`, `ProfileNames`, `ProfileBaseURL`, `SetProfileValue`), token resolution.
- `paths.go` — `GetConfigDir()` follows XDG on Linux/macOS, `%APPDATA%` on Windows. `GetConfigFilePath()` returns the YAML config path.
- `keyring.go` — `KeyringProvider` interface with `OSKeyring` implementation. `Keyring` struct wraps it with probe/get/set/delete/migrate. `ResolveTokenFull()` implements the token resolution chain: flag > env var > keyring > config file.
- Token resolution chain: `--token` flag > `ZENODO_TOKEN` env > OS keyring > config file fallback.
