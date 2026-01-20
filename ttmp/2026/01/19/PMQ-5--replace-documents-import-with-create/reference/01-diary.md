---
Title: Diary
Ticket: PMQ-5
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
      Note: Wire documents create and remove import (commit f5c7870)
    - Path: internal/cli/documents_commands.go
      Note: Replace import with create; support --content/--content-file/- stdin (commit f5c7870)
    - Path: internal/cli/rest_api_helpers.go
      Note: Add readFileOrStdin helper for stdin reads (commit f5c7870)
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-19T19:21:32.17294209-05:00
WhatFor: ""
WhenToUse: ""
---


# Diary

## Goal

Replace `documents import` with a more general `documents create` command, while preserving the common “create from file” workflow and adding stdin support.

## Step 1: Implement `documents create` and remove `documents import`

The backend already supports `POST /api/documents` for creation. Previously the CLI exposed this as `documents import` (file-centric). This step renames/reframes it to `documents create`, supports direct content input (`--content`), file input (`--content-file`), and stdin (`--content-file -`), and removes the old `import` command to avoid duplicate/overlapping UX.

**Commit (code):** f5c7870cbdb96ba84b8f7aea4c7bfb79cfd4f8d3 — "cli: replace documents import with create"

### What I did
- Replaced the `documents import` command implementation with `documents create` (same REST call, more flexible inputs).
- Added `readFileOrStdin` helper for `--content-file -` stdin support.
- Updated command wiring so `markdown-quizz documents` shows `create` and no longer shows `import`.
- Ran `GOWORK=off go test ./... -count=1`.

### Why
- “Import” implied file-only semantics; “create” is a better mental model for REST `POST /api/documents`.
- Supporting stdin enables scripted workflows (pipes) without temp files.

### What worked
- `go run ./cmd/markdown-quizz documents --help` shows `create` and does not show `import`.
- `documents create --content-file ./doc.md` still defaults the title from the filename if `--title` is omitted.

### What didn't work
- N/A

### What I learned
- A “create” command benefits from explicitly supporting three content sources (flag string, file, stdin); otherwise users keep reinventing wrappers around the file-only workflow.

### What was tricky to build
- Getting defaults correct without accidental empty creates: content must be non-empty, and title is required unless it can be derived from a real file path.

### What warrants a second pair of eyes
- Whether we want to keep accepting `--content-file` as the main file-path mechanism vs reintroducing a dedicated `--file` flag for clarity.

### What should be done in the future
- If we want smoother migration, add a `documents import` alias that prints a deprecation warning (not requested here).

### Code review instructions
- Start at `internal/cli/documents_commands.go` (create command) and `internal/cli/commands.go` (wiring).
- Validate with `GOWORK=off go test ./... -count=1` and `go run ./cmd/markdown-quizz documents --help`.

### Technical details
- Examples:
  - `go run ./cmd/markdown-quizz documents create --title \"My Doc\" --content \"# Hello\"`
  - `go run ./cmd/markdown-quizz documents create --content-file ./doc.md`
  - `cat ./doc.md | go run ./cmd/markdown-quizz documents create --title \"My Doc\" --content-file -`
