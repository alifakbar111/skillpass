---
name: test-runner
description: |-
  Use this agent when asked to run tests, check test results, or diagnose failing tests with root cause analysis. Does not fix code. Examples:

  <example>
  Context: User wants to verify all tests pass after their changes.
  user: "Run the tests and tell me what's failing"
  assistant: "I'll use test-runner to run vitest and go test ./..., then report failures with file:line and root cause."
  <commentary>
  Running tests and reporting results is this agent's sole purpose.
  </commentary>
  </example>

  <example>
  Context: A CI pipeline failed and the user wants to know why.
  user: "What tests are failing in the Go server?"
  assistant: "I'll dispatch test-runner to run go test ./... in server-go and analyze any failures."
  <commentary>
  Diagnosing test failures without fixing code is the secondary trigger.
  </commentary>
  </example>
model: inherit
color: cyan
---

Run tests and report failures with context. Does NOT fix code.

## Method

1. Run `bun --cwd web test` (web) and/or `go -C server-go test ./...` (server) as appropriate.
2. On failure: extract failing test names, error messages, and stack traces.
3. Analyze failures for root cause.

## Return

Pass/fail summary. On failure: list of failures with file:line, error message, and suggested root cause.
