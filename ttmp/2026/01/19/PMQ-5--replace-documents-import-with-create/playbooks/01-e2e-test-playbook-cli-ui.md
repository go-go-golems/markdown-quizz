---
Title: E2E test playbook (CLI + UI)
Ticket: PMQ-5
Status: complete
Topics:
    - cli
    - go
    - api
DocType: playbooks
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/cli/commands.go
      Note: Lists CLI surface used in the playbook
    - Path: internal/cli/documents_commands.go
      Note: documents create/update/delete/get/list commands
    - Path: internal/cli/quiz_commands.go
      Note: quiz submit/submit-batch commands
    - Path: internal/cli/submissions_commands.go
      Note: submissions mine/by-document/get commands
    - Path: legacy-version/client/src/store/api.ts
      Note: UI RTK Query endpoints exercised
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-19T20:50:26.459401037-05:00
WhatFor: ""
WhenToUse: ""
---


# E2E test playbook (CLI + UI)

## Goal

Verify end-to-end behavior of the Markdown-Quizz app: documents CRUD, quiz extraction, submissions, analytics — via both the CLI and the web UI.

## Preconditions

- You have two terminals.
- Backend server is running (default `http://127.0.0.1:9092`).
- UI dev server is running (see repo scripts; typically `pnpm -C legacy-version dev`).
- CLI uses default base URL unless overridden: `--base-url http://127.0.0.1:9092`.

## Test Data

Create a small Markdown file `./e2e-doc.md`:

```md
# E2E Document

This is an E2E test document.

<form id="f1">
id: f1
title: E2E Quiz
questions:
  - id: q1
    type: single
    prompt: Pick A
    options: [A, B]
    correct: A
  - id: q2
    type: short
    prompt: Say hello
</form>
```

Create responses JSON `./responses.json`:

```json
{"q1":"A","q2":"hello"}
```

Create batch submissions JSON `./subs.json`:

```json
[{"formId":"f1","responses":{"q1":"A","q2":"hello"}}]
```

## CLI checklist

### Documents: create/list/get/update/delete

1) Create document from file
- `go run ./cmd/markdown-quizz documents create --content-file ./e2e-doc.md --output json`
- Record `documentId` from output.

2) List documents (confirm it appears)
- `go run ./cmd/markdown-quizz documents list --scope all --output table`

3) Get document by ID
- `go run ./cmd/markdown-quizz documents get --id <documentId> --include-content --output json`

4) Get document by slug and confirm forms extraction
- From list output, capture `slug`.
- `go run ./cmd/markdown-quizz documents get --slug <slug> --output json`
- Expect `formsCount >= 1` and form id `f1` present.

5) Update document title and publish state
- `go run ./cmd/markdown-quizz documents update --id <documentId> --title \"E2E Document (Updated)\" --publish --output json`

6) Validate update
- `go run ./cmd/markdown-quizz documents get --id <documentId> --output json`
- Expect title updated, `isPublished=true`.

### Quiz submissions: single and batch

7) Submit a single quiz response from file
- `go run ./cmd/markdown-quizz quiz submit --document-id <documentId> --form-id f1 --responses-file ./responses.json --output json`
- Record `submissionId` from output.

8) Submit batch from file
- `go run ./cmd/markdown-quizz quiz submit-batch --document-id <documentId> --submissions-file ./subs.json --output json`

### Submissions: mine/by-document/get

9) List my submissions
- `go run ./cmd/markdown-quizz submissions mine --output table`
- Expect at least one row for `<documentId>` / `f1`.

10) List submissions by document (two ways)
- `go run ./cmd/markdown-quizz submissions by-document --document-id <documentId> --output table`
- `go run ./cmd/markdown-quizz documents submissions --id <documentId> --output table`
- Expect both to show the same submission IDs.

11) Get submission detail
- `go run ./cmd/markdown-quizz submissions get --id <submissionId> --output json`
- Expect `formDefinition` present and `responsesCount` matches `responses`.

### Analytics

12) Get document analytics
- `go run ./cmd/markdown-quizz documents analytics --id <documentId> --output json`
- Expect `totalSubmissions >= 1` and sensible score stats.

### Delete

13) Delete document
- `go run ./cmd/markdown-quizz documents delete --id <documentId> --output json`

14) Validate deletion
- `go run ./cmd/markdown-quizz documents get --id <documentId> --output json` should return a REST error.

## UI checklist

Note: UI routes/components may evolve; use this as a feature checklist.

1) Documents list page
- Confirms the created document appears after CLI create (or after refresh).
- Confirms updated title/published state is reflected after CLI update (or after refresh).

2) Document detail view (by slug)
- Navigates to the document page.
- Confirms extracted quizzes/forms are visible (at least `f1`).

3) Quiz submission flow
- Submits responses for `f1`.
- Confirms success state (submission id and/or score if shown).

4) My submissions view
- Shows the new submission(s) created via CLI and/or UI.
- Clicking a submission shows detail (including responses and form definition, if UI exposes it).

5) Analytics view (if present)
- Shows totals/stats consistent with CLI `documents analytics`.

6) Delete behavior (if UI supports delete)
- After deleting via CLI, UI no longer shows the document (after refresh).

## Troubleshooting quick hits

- CLI points at wrong server:
  - Add `--base-url http://127.0.0.1:9092` to the command.
- Server not running / port conflict:
  - Use `lsof-who -p 9092 -k` and restart.
- Empty forms in `documents get --slug`:
  - Ensure your markdown uses `<form id="..."> ... </form>` and the YAML is valid.
