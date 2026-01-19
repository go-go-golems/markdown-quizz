---
Title: Diary
Ticket: PMQ-2
Status: active
Topics:
    - frontend
    - backend
    - api
    - porting
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go.work
      Note: Workspace-level go test baseline (root ./... failure)
    - Path: markdown-quizz/legacy-version/server/db.ts
      Note: Legacy TS server tests fail with 'Database not available'
    - Path: markdown-quizz/legacy-version/server/quiz.submitMultiple.test.ts
      Note: Failing baseline tests (Database not available)
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-19T15:45:35.60951032-05:00
WhatFor: ""
WhenToUse: ""
---



# Diary

## Goal

Track implementation steps for PMQ-2 (tRPC + React Query → REST + RTK Query) with frequent, copy/paste-ready notes, exact commands, failures, and validation instructions.

## Step 1: Baseline (pre-change)

Established a pre-change baseline for Go + legacy frontend tooling so we can tell when regressions are introduced during the cutover. The Go modules are green when tested from within each module directory, but the workspace root `go test ./...` currently fails due to how `go.work` is being invoked.

The legacy TypeScript server test suite is currently not fully runnable in this environment because a subset of tests expect a database to be available and fail early with `Database not available`. This is useful context when deciding what “green” means during the migration (Go contract tests will become the authoritative backend validation).

### What I did
- Tried to run workspace-level tests: `go test ./...` (failed; see “What didn't work”)
- Ran Go module tests:
  - `cd markdown-quizz && make test`
  - `cd glazed && go test ./...`
  - `cd uhoh && go test ./...`
- Ran legacy frontend typecheck: `pnpm -C markdown-quizz/legacy-version check`
- Ran legacy frontend tests: `pnpm -C markdown-quizz/legacy-version test` (failed; see “What didn't work”)

### Why
- Capture a “known-good” point for Go before introducing a new `/api/*` surface and contract tests.
- Identify existing failing tests so we don’t misattribute failures introduced by the migration.

### What worked
- `markdown-quizz` Go tests pass via `make test` (which uses `GOWORK=off`).
- `glazed` and `uhoh` Go tests pass when run from their module roots.
- Legacy frontend typecheck (`tsc --noEmit`) passes.

### What didn't work
- `go test ./...` from the workspace root failed:
  - Error: `pattern ./...: directory prefix . does not contain modules listed in go.work or their selected dependencies`
- `pnpm -C markdown-quizz/legacy-version test` failed:
  - `server/quiz.submitMultiple.test.ts` (2 tests) → `TRPCError: Database not available`
  - Stack points at `markdown-quizz/legacy-version/server/db.ts:96` throwing `new Error("Database not available")`

### What I learned
- For now, “Go backend green” should be measured per-module, not via `go test ./...` at the workspace root.
- The legacy TypeScript server tests are not a reliable signal in this environment; PMQ-2 needs Go-side REST contract tests to become the primary backend correctness guardrail.

### What was tricky to build
- N/A (baseline only)

### What warrants a second pair of eyes
- Whether the workspace root `go test ./...` failure is expected for this repo layout or indicates a misconfigured `go.work` / tooling assumption.

### What should be done in the future
- Decide whether to fix workspace-root Go test invocation (optional).
- Decide whether to disable or rework legacy TS server tests once Go REST becomes the only backend.

### Code review instructions
- Validate Go baseline:
  - `cd markdown-quizz && make test`
  - `cd glazed && go test ./...`
  - `cd uhoh && go test ./...`
- Validate legacy frontend typecheck: `pnpm -C markdown-quizz/legacy-version check`

### Technical details
- Workspace-root `go test ./...` failure reproduced from `/home/manuel/workspaces/2026-01-05/port-markdown-quizz`.
