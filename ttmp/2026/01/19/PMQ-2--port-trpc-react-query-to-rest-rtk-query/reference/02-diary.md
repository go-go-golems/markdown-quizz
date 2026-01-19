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
    - Path: markdown-quizz/internal/cli/serve.go
      Note: Mount /api REST handler alongside /api/trpc
    - Path: markdown-quizz/internal/rest/server.go
      Note: REST /api handler implementation (documents + quiz)
    - Path: markdown-quizz/internal/rest/server_test.go
      Note: REST contract tests (sqlite-backed)
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

**Commit (docs):** 9865e87 — "PMQ-2: add docs and baseline"

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

## Step 2: Implement `/api/*` REST handlers + contract tests

Implemented the initial REST surface area under `/api/*` (documents + quiz submissions) to mirror the current tRPC procedures used by the SPA. This gives us a concrete, debuggable HTTP interface with explicit status codes and a uniform error envelope, while still keeping `/api/trpc` mounted for the moment so we can cut over the frontend in a single later step.

Added Go-side integration-style tests that drive the REST handler against a real sqlite DB (migrations + stores) to lock in request/response shapes and ensure quiz scoring + analytics behavior stays correct through the migration.

**Commit (code):** 35f8f75 — "api: add REST /api endpoints"

### What I did
- Added `internal/rest` with a `Server` that routes `/api/documents/*` and `/api/quiz/*`
- Mounted the REST handler in `internal/cli/serve.go` at `mux.Handle("/api/", restServer)` (after `/api/trpc`)
- Added `internal/rest/server_test.go` to exercise document creation + quiz submissions + analytics

### Why
- Make the backend interface explicit and testable without tRPC envelopes/batching/superjson.
- Provide contract tests that become the backend “green” signal before the frontend cutover.

### What worked
- `GOWORK=off go test ./... -count=1` passes, including the new REST tests.
- The REST endpoints cover the full inventory table for documents + quiz submissions.

### What didn't work
- N/A in this step.

### What I learned
- The sqlite-backed stores are a good foundation for handler-level contract tests: we can validate quiz scoring and analytics without spinning up a server process.

### What was tricky to build
- Path routing without a router library (ensuring `/api/trpc` remains unaffected while `/api/*` is mounted broadly).

### What warrants a second pair of eyes
- Status code choices (`201` for quiz submit vs `200`) and strict JSON decoding (`DisallowUnknownFields`)—these are contract decisions that the frontend will now need to match exactly.

### What should be done in the future
- Add more negative tests (bad IDs, missing fields, not-found cases) once the frontend cutover is underway.

### Code review instructions
- Start at `markdown-quizz/internal/rest/server.go` (routing + request/response shapes).
- Validate with `cd markdown-quizz && GOWORK=off go test ./... -count=1`.

### Technical details
- Error envelope: `{ "error": { "code": "...", "message": "...", "details": ... } }`
