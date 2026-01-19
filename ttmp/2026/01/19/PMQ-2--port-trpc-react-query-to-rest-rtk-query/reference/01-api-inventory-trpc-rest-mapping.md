---
Title: API Inventory (tRPC → REST mapping)
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
    - Path: markdown-quizz/internal/documents/store.go
      Note: Documents + quiz form persistence that REST endpoints will call
    - Path: markdown-quizz/internal/quiz/store.go
      Note: Quiz submission/analytics persistence that REST endpoints will call
    - Path: markdown-quizz/internal/rest/server.go
      Note: Final Go REST /api surface (documents + quiz)
    - Path: markdown-quizz/legacy-version/client/src/store/api.ts
      Note: RTK Query endpoints + tag invalidations
    - Path: markdown-quizz/legacy-version/client/src/main.tsx
      Note: Redux Provider wiring for RTK Query
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-19T11:07:10.235688791-05:00
WhatFor: ""
WhenToUse: ""
---


# API Inventory (tRPC → REST mapping)

## Goal

Provide a complete, auditable inventory of every tRPC procedure used by the frontend, how it is implemented in the Go backend today, and the proposed typed REST replacement (HTTP method, path, request/response schema, and RTK Query caching/tagging plan).

## Context

The legacy frontend used `@trpc/react-query` hooks and `@tanstack/react-query` caching/invalidation patterns. The Go backend now exposes the final REST API under `/api/*` (implemented in `internal/rest/server.go`), consumed by the SPA via RTK Query.

This document treats the frontend+backend as a regulated system: each “interface surface” must be enumerated, typed, and testable. The intent is to replace tRPC+React Query with plain HTTP JSON REST endpoints and RTK Query (`@reduxjs/toolkit/query`) without changing product behavior.

Important constraint: **no migration/backwards-compatibility requirements**. We can change frontend + backend together and rip out tRPC in one cutover.

**Scope of inventory (frontend calls):**
- `documents.*` procedures used by pages: Home, Admin, DocumentEditor, DocumentView, Analytics
- `quiz.*` procedures used by pages/components: QuizForm, MarkdownRenderer, MySubmissions, SubmissionReview

**Explicitly out of scope:**
- Demo-only references in `ComponentShowcase.tsx` (e.g. `ai.chat`) unless that becomes a real feature.

## Quick Reference

### Source of truth for “what the UI calls”

Frontend call sites (extracted from `legacy-version/client/src`):

- `documents.list` (query)
- `documents.myDocuments` (query)
- `documents.getBySlug` (query)
- `documents.getById` (query)
- `documents.create` (mutation)
- `documents.update` (mutation)
- `documents.delete` (mutation)
- `documents.analytics` (query)
- `documents.submissions` (query)
- `quiz.submit` (mutation)
- `quiz.submitMultiple` (mutation)
- `quiz.mySubmissions` (query)
- `quiz.getSubmission` (query)

### Data model snapshots (current UI expectations)

These are the shapes the UI currently reads (field presence matters more than “class vs plain object”):

- **Document**
  - `id: number`
  - `title: string`
  - `slug: string`
  - `content: string`
  - `description: string | null`
  - `category: string | null`
  - `isPublished: boolean`
  - `authorId: number`
  - `createdAt: string` (rendered via `new Date(...)`)
  - `updatedAt: string` (rendered via `new Date(...)`)

- **QuizForm**
  - `formId: string`
  - `definition: unknown` (YAML-derived structure; must round-trip as JSON)

- **QuizSubmission**
  - `id: number`
  - `userId: number`
  - `documentId: number`
  - `formId: string`
  - `responses: Record<string, unknown>`
  - `score: number | null`
  - `maxScore: number | null`
  - `submittedAt: string` (rendered via `new Date(...)`)

- **DocumentAnalytics**
  - `totalSubmissions: number`
  - `averageScore: number` (percentage, 0–100)
  - `highestScore: number` (percentage, 0–100)
  - `lowestScore: number` (percentage, 0–100)

### REST conventions (proposal)

**Base URL**: `/api`

**Dates**: ISO 8601 strings in UTC (e.g. `"2026-01-19T16:07:00Z"`). If the backend stores strings, normalize on output.

**Error envelope** (uniform across endpoints):

```json
{
  "error": {
    "code": "bad_request",
    "message": "title is required",
    "details": { "field": "title" }
  }
}
```

### Inventory table (tRPC → REST → RTK Query)

For each procedure: what the UI used historically (tRPC), and the REST+RTK replacement that becomes the only supported interface after cutover.

| ID | tRPC procedure | UI usage (file) | Input | Output | Go implementation (final) | REST replacement | RTK Query endpoint + tags |
|---:|---|---|---|---|---|---|---|
| 1 | `documents.list` | `client/src/pages/Home.tsx` | none | `Document[]` | `internal/rest/server.go` | `GET /api/documents?scope=all` | `useListDocumentsQuery` (`providesTags: ['Documents']`) |
| 2 | `documents.myDocuments` | `client/src/pages/Admin.tsx` | none | `Document[]` | `internal/rest/server.go` | `GET /api/documents?scope=mine` | `useMyDocumentsQuery` (`providesTags: ['MyDocuments']`) |
| 3 | `documents.getBySlug` | `client/src/pages/DocumentView.tsx` | `{ slug }` | `Document & { forms: QuizForm[] }` | `internal/rest/server.go` | `GET /api/documents/by-slug/{slug}` | `useDocumentBySlugQuery` (`providesTags: (doc) => [{type:'Document',id:doc.id}]`) |
| 4 | `documents.getById` | `client/src/pages/DocumentEditor.tsx`, `client/src/pages/Analytics.tsx` | `{ id }` | `Document` | `internal/rest/server.go` | `GET /api/documents/{id}` | `useDocumentByIdQuery` (`providesTags: (doc)=>[{type:'Document',id:doc.id}]`) |
| 5 | `documents.create` | `client/src/pages/DocumentEditor.tsx` | `{ title, content, description?, category?, isPublished }` | `{ id, slug }` | `internal/rest/server.go` | `POST /api/documents` | `useCreateDocumentMutation` (`invalidatesTags: ['Documents','MyDocuments']`) |
| 6 | `documents.update` | `client/src/pages/DocumentEditor.tsx` | `{ id, title?, content?, description?, category?, isPublished? }` | `{ success: true }` | `internal/rest/server.go` | `PATCH /api/documents/{id}` | `useUpdateDocumentMutation` (`invalidatesTags: [{type:'Document',id},'Documents','MyDocuments']`) |
| 7 | `documents.delete` | `client/src/pages/Admin.tsx` | `{ id }` | `{ success: true }` | `internal/rest/server.go` | `DELETE /api/documents/{id}` | `useDeleteDocumentMutation` (`invalidatesTags: ['Documents','MyDocuments']`) |
| 8 | `documents.analytics` | `client/src/pages/Analytics.tsx` | `{ id }` | `DocumentAnalytics` | `internal/rest/server.go` | `GET /api/documents/{id}/analytics` | `useDocumentAnalyticsQuery` (`providesTags: [{type:'DocAnalytics',id}]`) |
| 9 | `documents.submissions` | `client/src/pages/Analytics.tsx` | `{ id }` | `Array<{submission: QuizSubmission, userName: string | null}>` | `internal/rest/server.go` | `GET /api/documents/{id}/submissions` | `useDocumentSubmissionsQuery` (`providesTags: [{type:'DocSubmissions',id}]`) |
| 10 | `quiz.submit` | `client/src/components/QuizForm.tsx` | `{ documentId, formId, responses }` | `{ id, score, maxScore }` | `internal/rest/server.go` | `POST /api/quiz/submissions` | `useSubmitQuizMutation` (`invalidatesTags: ['MySubmissions',{type:'DocAnalytics',id:documentId},{type:'DocSubmissions',id:documentId}]`) |
| 11 | `quiz.submitMultiple` | `client/src/components/MarkdownRenderer.tsx` | `{ documentId, submissions: [{formId,responses}...] }` | `{ results: [{formId,score,maxScore}...] }` | `internal/rest/server.go` | `POST /api/quiz/submissions/batch` | `useSubmitQuizBatchMutation` (same invalidation as #10) |
| 12 | `quiz.mySubmissions` | `client/src/pages/MySubmissions.tsx` | none | `Array<{submission: QuizSubmission, documentTitle: string, documentSlug: string}>` | `internal/rest/server.go` | `GET /api/quiz/submissions?scope=mine` | `useMySubmissionsQuery` (`providesTags: ['MySubmissions']`) |
| 13 | `quiz.getSubmission` | `client/src/pages/SubmissionReview.tsx` | `{ id }` | `{ submission, documentTitle, documentSlug, formDefinition }` | `internal/rest/server.go` | `GET /api/quiz/submissions/{id}` | `useSubmissionByIdQuery` (`providesTags: (r)=>[{type:'Submission',id:r.submission.id}]`) |

### Migration invariants checklist (non-negotiable)

- Every endpoint must have a written request/response schema (even if enforced in code later).
- Every endpoint must have explicit HTTP status codes.
- Date fields must be consistent across endpoints (no “sometimes Date object” vs “sometimes string”).
- Cache invalidation must be explicit via RTK tags, not “magic”.
- Batch behavior must be either preserved (`submitMultiple`) or intentionally removed with a compensating design (but the UI currently calls it).
- The cutover deletes tRPC: once REST+RTK is in place, remove `/api/trpc` and remove `@trpc/*` and React Query from the SPA. (Done.)

## Usage Examples

### Example RTK Query skeleton (proposed)

```ts
export const api = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({ baseUrl: '/api', credentials: 'include' }),
  tagTypes: ['Documents', 'MyDocuments', 'Document', 'MySubmissions', 'DocAnalytics', 'DocSubmissions', 'Submission'],
  endpoints: (build) => ({
    listDocuments: build.query<Document[], void>({
      query: () => ({ url: 'documents', params: { scope: 'all' } }),
      providesTags: ['Documents'],
    }),
    createDocument: build.mutation<{id:number;slug:string}, CreateDocumentRequest>({
      query: (body) => ({ url: 'documents', method: 'POST', body }),
      invalidatesTags: ['Documents','MyDocuments'],
    }),
  }),
})
```

## Related

- Port plan (design doc): `ttmp/2026/01/19/PMQ-2--port-trpc-react-query-to-rest-rtk-query/design-doc/01-port-plan-trpc-react-query-typed-rest-rtk-query.md`
