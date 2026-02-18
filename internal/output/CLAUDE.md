# internal/output

Output formatters: JSON, table, CSV.

- `formatter.go` — `Format()` dispatches to json/table/csv based on output mode. Handles field selection (`--fields`), nested field resolution (dot notation), and conversion of structs to row maps.
- `json.go`, `table.go`, `csv.go` — Individual formatters.
- `diff.go` — `DiffMetadata()` shows colored before/after diff of metadata changes (green=added, red=removed, yellow=changed). Used by deposit update command.
- Table formatter uses `olekukonko/tablewriter` v0.0.5 (not v1.x which has breaking API changes).
