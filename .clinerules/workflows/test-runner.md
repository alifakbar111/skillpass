# Test Runner (Cline Workflow)

Invoke via: `/test-runner <target>`

Run tests and report failures with context. Does NOT fix code.

## Method

1. Run `bun --cwd web test` (web) and/or `go -C server-go test ./...` (server) as appropriate.
2. On failure: extract failing test names, error messages, and stack traces.
3. Analyze failures for root cause.

## Return

Pass/fail summary. On failure: list of failures with file:line, error message, and suggested root cause.
