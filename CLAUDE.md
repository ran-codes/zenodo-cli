## Rules

- Be concise. Iterate quickly. Only give verbose/detailed answers when explicitly asked.
- Ask before doing heavyweight work (subagents, large number of tool calls, broad searches).

## Developer Workflow

- Code changes go on a feature branch, not directly on main.
- Open a PR describing what changed and why.
- Ask for review before merging. Only merge after explicit approval.
- Include an ELI5 section in PR descriptions.
- Run `go build ./...` and `go test ./...` before pushing.
- See `.claude/reference/v0.1/DEVELOP.md` for project structure and build instructions.
