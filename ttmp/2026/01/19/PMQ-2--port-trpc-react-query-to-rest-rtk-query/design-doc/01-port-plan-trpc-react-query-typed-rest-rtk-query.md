---
Title: 'Port Plan: tRPC + React Query → Typed REST + RTK Query'
Ticket: PMQ-2
Status: active
Topics:
    - frontend
    - backend
    - api
    - porting
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: markdown-quizz/internal/rest/server.go
      Note: REST /api handler implementation (documents + quiz)
    - Path: markdown-quizz/legacy-version/client/src/store/api.ts
      Note: RTK Query API slice (REST contract + tags)
    - Path: markdown-quizz/legacy-version/client/src/components/MarkdownRenderer.tsx
      Note: quiz.submitMultiple usage
    - Path: markdown-quizz/legacy-version/client/src/components/QuizForm.tsx
      Note: quiz.submit usage
    - Path: markdown-quizz/legacy-version/client/src/pages/Admin.tsx
      Note: documents.myDocuments + documents.delete usage
    - Path: markdown-quizz/legacy-version/client/src/pages/Analytics.tsx
      Note: documents.getById + analytics + submissions usage
    - Path: markdown-quizz/legacy-version/client/src/pages/DocumentEditor.tsx
      Note: documents.getById + create/update usage
    - Path: markdown-quizz/legacy-version/client/src/pages/DocumentView.tsx
      Note: documents.getBySlug usage
    - Path: markdown-quizz/legacy-version/client/src/pages/Home.tsx
      Note: documents.list usage
    - Path: markdown-quizz/legacy-version/client/src/pages/MySubmissions.tsx
      Note: quiz.mySubmissions usage
    - Path: markdown-quizz/legacy-version/client/src/pages/SubmissionReview.tsx
      Note: quiz.getSubmission usage
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-19T11:07:10.171082481-05:00
WhatFor: ""
WhenToUse: ""
---


# Port Plan: tRPC + React Query → Typed REST + RTK Query

## Executive Summary

We will replace the frontend’s tRPC + React Query stack with a typed HTTP REST API consumed via RTK Query, and we explicitly do **not** care about backwards-compatibility or migration compatibility. We can change frontend + backend together and rip out tRPC in one cutover.

This ticket is intentionally systematic: we treat the current tRPC API as a safety-critical interface surface. No endpoint is replaced until we can answer, in writing: “Who calls it?”, “What does it accept/return?”, “What are its invariants?”, and “What is the RTK Query caching/invalidations story?”

## Problem Statement

The current frontend is coupled to:

- **tRPC** (procedure paths and tRPC request/response envelopes)
- **React Query** (cache, invalidation patterns, mutation side-effects)
- **superjson** (runtime serialization behavior and implicit date handling)

This coupling has costs:

1. **Opaque interface surface**: the real “API contract” is distributed across TS types, router definitions, and runtime envelopes.
2. **Harder debugging**: batching/superjson/tRPC error codes make network traces non-obvious.
3. **Porting friction**: the Go backend currently implements a compatibility tRPC adapter; long-term we want plain HTTP endpoints that are easier to document, test, and evolve.

The goal is not to change product behavior. The goal is to replace the interface mechanism with an explicit, typed, auditable REST interface and RTK Query data fetching. We will do this as a **big-bang cutover** (no compatibility window), because we control both sides of the interface.

## Proposed Solution

### 1) Define the target REST API surface (typed)

Create a REST API under `/api` that covers all currently-used features:

- Documents
  - list (`scope=all` vs `scope=mine`)
  - get by slug
  - get by id
  - create/update/delete
  - analytics + submissions by document
- Quiz
  - submit (single)
  - submit batch (submitMultiple)
  - list my submissions
  - get submission detail (with `formDefinition`)

Every endpoint gets:

- method + path
- request JSON schema (even for `GET`, define params)
- response JSON schema
- error schema + status codes

### 2) Implement REST endpoints in the Go backend (final API)

Implement `/api/*` REST endpoints that call the same underlying stores (`internal/documents`, `internal/quiz`). The REST API becomes the only supported API for the frontend, and tRPC is removed as part of the cutover.

### 3) Replace the frontend data layer with RTK Query

Create a single RTK Query API slice:

- `baseUrl: '/api'`
- `credentials: 'include'` (safe default; can be relaxed later)
- explicit `tagTypes` and invalidations to mirror the old `trpc.useUtils().invalidate()` usage.

Migrate page-by-page:

- Home: document list
- Admin: myDocuments + delete
- DocumentEditor: getById + create/update
- DocumentView: getBySlug
- Analytics: getById + analytics + submissions
- QuizForm/MarkdownRenderer: submit / submitMultiple
- MySubmissions / SubmissionReview

### 4) Rip out tRPC (same cutover)

In the same change set / cutover:

- Remove `@trpc/*`, `@tanstack/react-query`, `superjson` from the SPA.
- Remove the Go `/api/trpc` adapter and any tRPC-specific request/response handling.

## Design Decisions

### JSON over superjson

We will use plain JSON on the wire. Dates must be ISO 8601 strings. The UI must parse dates explicitly or treat them as strings.

Rationale: eliminates hidden runtime serialization behavior.

### Stable REST shapes and status codes

Standardize on:

- `200` for successful reads
- `201` for creates
- `200` for deletes (response `{success:true}`)
- `400` for validation errors
- `404` for missing resources
- `500` for unexpected errors

Rationale: predictable status codes make debugging and client behavior robust.

### Explicit cache invalidation via tags

We will not rely on implicit “invalidate everything” behavior. Each mutation must state what it invalidates.

Rationale: predictable caching is the difference between a UI that “usually works” and one that is operationally reliable.

### No compatibility window

We will not maintain a period where both `/api/trpc` and `/api/*` are supported for the same UI. We will implement REST endpoints, update the UI to RTK Query, and delete the tRPC adapter and dependencies in one coordinated cutover.

Rationale: we own both ends; a compatibility window increases complexity and doubles the number of behaviors to keep correct.

### Locked contract decisions (implemented)

- **Base URL**: `/api`
- **Transport**: JSON only (no superjson); dates are ISO-8601 strings
- **Error envelope**:

```json
{
  "error": { "code": "bad_request", "message": "title is required", "details": { "field": "title" } }
}
```

- **Deletes**: `200` with `{ "success": true }`
- **Strict JSON decoding**: request bodies reject unknown fields (`DisallowUnknownFields`)

### Typing strategy (chosen)

Handwritten TypeScript types co-located with RTK Query endpoint definitions in `legacy-version/client/src/store/api.ts` (no OpenAPI generation in this repo at this time). Contract tests in Go backstop drift.

## Alternatives Considered

### A1) Keep tRPC, but change only the client library

Not aligned with the goal: we want to remove procedural envelopes and tRPC semantics from the wire.

### A1.5) Phased migration with a compatibility window (rejected)

Maintain `/api/trpc` while introducing `/api/*`, then migrate pages gradually.

Rejected because it is explicitly out of scope: we do not care about backwards compatibility, and a compatibility window increases moving parts (two APIs, two client stacks, duplicated test surface).

### A2) REST without a typed contract (handwritten TS types only)

Feasible, but risky: types drift over time and backend changes become “silent breakages”.

If we choose this, we must compensate with:

- strict server-side validation
- contract tests
- a policy that every endpoint change updates both server and client types in the same PR

### A3) OpenAPI-first (recommended for “typed REST”)

Define OpenAPI for `/api/*`, then generate:

- TS client types
- (optionally) RTK Query endpoint definitions (via generator or small template)

Pros:

- strongest contract story
- best for long-term maintainability

Cons:

- higher upfront work
- requires generator/tooling decisions

### A4) JSON Schema + zod in frontend

Possible, but adds its own complexity. OpenAPI tends to integrate better with common HTTP tooling and docs.

## Implementation Plan

1. Finalize REST endpoint mapping (see `reference/01-api-inventory-trpc-rest-mapping.md`).
2. Decide the typing strategy:
   - OpenAPI-first vs handwritten TS types
3. Implement REST endpoints in Go (`/api/*`) and add contract tests.
4. Add RTK Query slice and migrate UI call sites (in the same branch).
5. Delete tRPC adapter + tRPC dependencies and verify behavior.

Operational checklist (OSHA-style):

- Every endpoint has request/response schema + error schema.
- Every endpoint has at least one integration test.
- Every mutation has a documented invalidation policy (RTK tags).
- All dates are ISO strings; no implicit date decoding.

## Open Questions

1. Do we want `/api/documents?scope=mine` or separate `/api/my/documents`? (Resolved: keep `scope=mine` query param.)
2. Should we preserve `submitMultiple` as a single endpoint, or push the UI toward multiple `submit` calls? (Resolved: preserve as `POST /api/quiz/submissions/batch`.)
3. Should we introduce pagination now (documents list, submissions list), or defer until needed? (Resolved: defer.)
4. How strictly do we want to validate `responses` payloads (currently `map[string]any` / JSON blob)? (Resolved: treat as JSON blob; validate only outer envelope.)

## References

- API inventory + mapping table: `ttmp/2026/01/19/PMQ-2--port-trpc-react-query-to-rest-rtk-query/reference/01-api-inventory-trpc-rest-mapping.md`
- Go REST handler: `internal/rest/server.go`
- Frontend RTK Query layer: `legacy-version/client/src/store/api.ts`
