---
Title: Diary
Ticket: PMQ-4
Status: active
Topics:
    - cli
    - go
    - api
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/cli/commands.go
      Note: Wire new documents commands into cobra parent (commit 5e2e1a6)
    - Path: internal/cli/documents_commands.go
      Note: Add documents update/delete/analytics/submissions CLI commands (commit 5e2e1a6)
    - Path: internal/cli/rest_api_helpers.go
      Note: Add trimOptionalStringPtr helper for PATCH semantics (commit 5e2e1a6)
    - Path: internal/restclient/client_test.go
      Note: Add path/method tests for new document endpoints (commit 5e2e1a6)
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-19T18:01:48.873606821-05:00
WhatFor: ""
WhenToUse: ""
---


# Diary

## Goal

Bring the Glazed CLI up to parity with the REST API surface for documents (update/delete/analytics/submissions), with tests and a reviewable commit history.

## Step 1: Add missing document commands

The REST API already exposed `PATCH /api/documents/:id`, `DELETE /api/documents/:id`, `GET /api/documents/:id/analytics`, and `GET /api/documents/:id/submissions`, but the Glazed CLI only supported `documents list|get|import`. This step adds the missing commands so the CLI covers the full documents API surface.

**Commit (code):** 5e2e1a63ee8d6d4f52450b5bd79cfd4518a7b260 — "cli: add documents update/delete/analytics/submissions"

### What I did
- Implemented `documents update|delete|analytics|submissions` Glazed commands and wired them into `documents` cobra parent.
- Added a small helper `trimOptionalStringPtr` for clean PATCH request construction.
- Added `internal/restclient` path/method tests for update/delete/analytics/submissions endpoints.
- Ran `GOWORK=off go test ./... -count=1`.

### Why
- CLI parity makes it possible to manage the full document lifecycle and inspect analytics without going through the web UI.

### What worked
- `documents update` uses `--publish/--unpublish` (mutually exclusive) to avoid the “bool default can’t express PATCH intent” pitfall.
- REST client tests catch accidental path/method regressions.

### What didn't work
- N/A

### What I learned
- For PATCH semantics, it’s safer to use “action flags” (`--publish/--unpublish`) than a single `--is-published` bool, because a default `false` would otherwise look like an explicit update.

### What was tricky to build
- Keeping PATCH semantics consistent with the backend (`no fields to update`) while still providing CLI ergonomics (content from `--content-file`).

### What warrants a second pair of eyes
- The exact “empty string means unset” behavior for string fields: currently the CLI trims and treats empty strings as “do not update”, which means you can’t set a field to an explicit empty string via CLI.

### What should be done in the future
- Add explicit `--clear-description` / `--clear-category` flags if we need to support setting nullable fields to NULL/empty via CLI.

### Code review instructions
- Start at `internal/cli/documents_commands.go` (new commands) and `internal/cli/commands.go` (wiring).
- Validate with `GOWORK=off go test ./... -count=1`, then run `go run ./cmd/markdown-quizz documents --help` and `go run ./cmd/markdown-quizz documents update --help`.

### Technical details
- Examples:
  - `go run ./cmd/markdown-quizz documents update --id 1 --title "New title"`
  - `go run ./cmd/markdown-quizz documents update --id 1 --content-file ./doc.md`
  - `go run ./cmd/markdown-quizz documents update --id 1 --publish`
  - `go run ./cmd/markdown-quizz documents delete --id 1`
  - `go run ./cmd/markdown-quizz documents analytics --id 1`
  - `go run ./cmd/markdown-quizz documents submissions --id 1`
