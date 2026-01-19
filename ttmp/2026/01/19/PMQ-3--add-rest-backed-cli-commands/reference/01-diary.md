---
Title: Diary
Ticket: PMQ-3
Status: active
Topics:
    - backend
    - go
    - api
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: markdown-quizz/internal/cli/commands.go
      Note: Register documents/quiz/submissions command groups
    - Path: markdown-quizz/internal/cli/documents_commands.go
      Note: documents list/get/import commands
    - Path: markdown-quizz/internal/cli/quiz_commands.go
      Note: quiz submit and submit-batch commands
    - Path: markdown-quizz/internal/cli/rest_api_helpers.go
      Note: Shared API flags + JSON file helpers
    - Path: markdown-quizz/internal/cli/submissions_commands.go
      Note: submissions mine/by-document/get commands
    - Path: markdown-quizz/internal/restclient/client.go
      Note: HTTP+JSON client for /api/* (used by CLI)
    - Path: markdown-quizz/internal/restclient/client_test.go
      Note: Path/query/error envelope tests
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-19T17:19:23.81611608-05:00
WhatFor: ""
WhenToUse: ""
---


# Diary

## Goal

Track implementation steps for adding a Glazed-based CLI that calls the Go REST API (`/api/*`) to import markdown documents (extracting `<form>` quizzes), submit quiz responses (single + batch), and list/get submissions.

## Context

The backend exposes a REST API under `/api/*` (no `/api/trpc`). This ticket adds CLI subcommands that speak HTTP to that API (rather than touching sqlite directly), using Glazed so output can be rendered as `table/json/yaml/csv/...`.

## Quick Reference

### Commands (examples)

```bash
# Documents
markdown-quizz documents list --scope all
markdown-quizz documents import --file ./doc.md --title "My Doc" --output json
markdown-quizz documents get --slug my-doc-slug --output json

# Quiz submit (single) from JSON file (responses object)
markdown-quizz quiz submit --document-id 123 --form-id f1 --responses-file ./responses.json --output json

# Quiz submit (batch) from JSON file (array of {formId,responses})
markdown-quizz quiz submit-batch --document-id 123 --submissions-file ./subs.json --output json

# Submissions
markdown-quizz submissions mine --output table
markdown-quizz submissions by-document --document-id 123 --output table
markdown-quizz submissions get --id 456 --output json
```

## Usage Examples

### File formats

`responses.json` (single-submit):

```json
{ "q1": "a", "q2": ["x","y"] }
```

`subs.json` (batch-submit):

```json
[
  { "formId": "f1", "responses": { "q1": "a" } },
  { "formId": "f2", "responses": { "q2": ["x","y"] } }
]
```

### Running against a local server

```bash
# Terminal A
markdown-quizz serve

# Terminal B
markdown-quizz documents list --scope all --output table
```

## Related

## Step 1: Add a small REST client package

Created a dedicated Go client for the markdown-quizz REST API so CLI commands can be implemented cleanly without duplicating HTTP request/response handling. The client handles base URL normalization, JSON marshalling/unmarshalling, query params, timeouts, and the standard error envelope.

This is the “plumbing” layer that keeps Glazed commands thin: commands call `restclient.*` methods and just emit rows.

**Commit (code):** a8c5b93 — "cli: add REST API client"

### What I did
- Added `internal/restclient` with:
  - `Client` + `New(Options{BaseURL,Timeout})`
  - typed request/response structs for documents + quiz/submissions
  - helpers to decode error envelope into a structured `APIError`

### Why
- Centralize REST call logic and error handling for all CLI commands.

### What worked
- `GOWORK=off go test ./... -count=1` stayed green.

### What didn't work
- N/A

### What I learned
- Using an explicit `APIError` type makes CLI error output much clearer than raw body strings.

### What was tricky to build
- Base URL path prefix handling (supporting `http://host/prefix` and still hitting `/prefix/api/...`).

### What warrants a second pair of eyes
- Error parsing behavior for non-JSON error bodies (we fall back to raw text).

### What should be done in the future
- Add more endpoint coverage methods as the REST API expands.

### Code review instructions
- Start at `markdown-quizz/internal/restclient/client.go`.
- Validate with `cd markdown-quizz && GOWORK=off go test ./... -count=1`.

### Technical details
- All CLI-facing methods use `doJSON(ctx, method, "api/...", query, in, out)`.

## Step 2: Add Glazed CLI commands that call REST

Implemented `documents/*`, `quiz/*`, and `submissions/*` command groups, all using Glaze mode so output formatting is automatic. The CLI supports importing markdown to create a document (which triggers quiz extraction server-side), submitting responses for a single form, submitting a batch (submitMultiple replacement), and listing/getting submissions.

**Commit (code):** 4c08cdb — "cli: add REST-backed commands (documents/quiz/submissions)"

### What I did
- Added CLI commands under `internal/cli/*_commands.go` and registered them in `internal/cli/commands.go`
- Implemented flags:
  - `--base-url` and `--timeout-seconds` (API settings)
  - `documents import --file ...`
  - `quiz submit --responses-json ...` and `--responses-file ...`
  - `quiz submit-batch --submissions-json ...` and `--submissions-file ...`

### Why
- Provide a scriptable interface for populating the DB from markdown and exercising the REST API without needing the SPA.

### What worked
- `go run ./cmd/markdown-quizz --help` shows the new command groups.
- `go run ./cmd/markdown-quizz documents list --help` shows Glazed output flags automatically.

### What didn't work
- Initial schema wiring mistake: treated `schema.Section` as a pointer type and used a non-existent `fields.TypeDuration`. Fixed by using `timeout-seconds` as an int.

### What I learned
- In this repo’s Glazed APIs, `schema.NewSection(...)` returns a concrete section that should be passed by value into `schema.WithSections(...)`.

### What was tricky to build
- Keep the Glazed parsing model consistent (always decode from `parsedLayers` rather than reading cobra flags).

### What warrants a second pair of eyes
- Command UX: flag naming (`timeout-seconds`, JSON inputs) and whether we want friendlier YAML support later.

### What should be done in the future
- Add `documents update` (PATCH) and `documents delete` commands if we want full parity from CLI.

### Code review instructions
- Start at `markdown-quizz/internal/cli/commands.go` and then `markdown-quizz/internal/cli/*_commands.go`.
- Validate with `cd markdown-quizz && GOWORK=off go test ./... -count=1`.

### Technical details
- JSON file flags read raw JSON and validate shapes client-side before sending.

## Step 3: Add targeted tests for the REST client

Added lightweight httptest-backed tests that lock in URL joining semantics, query param handling, and error envelope decoding. This ensures the CLI plumbing is stable even if we refactor command code.

**Commit (code):** 8c20755 — "test: restclient"

### What I did
- Added `internal/restclient/client_test.go` using `httptest.NewServer`

### Why
- Prevent regressions in request path formation and error parsing (most common sources of CLI breakage).

### What worked
- Tests validate:
  - `/api/documents` path + `scope=mine`
  - `/api/quiz/submissions/batch` path
  - error envelope → `APIError`
  - base URL prefix handling

### What didn't work
- N/A

### What I learned
- Testing the low-level client gives better signal than trying to test cobra/glazed output.

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- Whether we should assert request bodies for submit endpoints (currently we only assert path/method).

### What should be done in the future
- Add body assertions for submit endpoints if we see regressions.

### Code review instructions
- Start at `markdown-quizz/internal/restclient/client_test.go`.
- Run: `cd markdown-quizz && GOWORK=off go test ./... -count=1`.

### Technical details
- Uses `httptest.NewServer` and decodes JSON directly in handlers.

## Step 4: Improve CLI output by de-pointering nullable fields

Adjusted CLI output rows to render nullable fields and nullable score values as actual values (or null) instead of Go pointer addresses in table mode.

**Commit (code):** 54554d8 — "cli: render nullable fields"

### What I did
- Added helper functions to emit nullable `*string/*int/*float64` as values for Glazed output rows.

### Why
- Table output should be human-readable; pointer addresses are not actionable.

### What worked
- `markdown-quizz submissions mine --output table` now shows numeric scores instead of `0xc000...`.

### What didn't work
- N/A

### What I learned
- Glazed table output will print pointer addresses unless we dereference before emitting.

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- Whether we want to normalize “missing score/maxScore” to `null` or `0` (currently `null`).

### What should be done in the future
- Add a `--no-nulls` option later if we want to coerce nulls to empty strings in tables.

### Code review instructions
- Start at `markdown-quizz/internal/cli/rest_api_helpers.go` and the `*_commands.go` row emitters.

### Technical details
- CLI derives `scorePct` when both `score` and `maxScore` exist and `maxScore > 0`.
