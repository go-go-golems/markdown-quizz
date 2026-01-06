---
Title: Legacy server analysis and Go architecture
Ticket: PMQ-1
Status: active
Topics:
    - backend
    - go
    - porting
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: glazed/cmd/examples/refactor-new-packages/main.go
      Note: Canonical example for schema/fields/values usage
    - Path: glazed/pkg/doc/tutorials/05-build-first-command.md
      Note: Glazed CLI patterns and layer parsing tutorial
    - Path: markdown-quizz/internal/cli/serve.go
      Note: First implementation of the planned Glazed-based CLI entrypoint
ExternalSources: []
Summary: 'Complete port plan: legacy inventory, API parity, and Go architecture.'
LastUpdated: 2026-01-05T19:02:28.844589884-05:00
WhatFor: Port the legacy TypeScript server to Go while preserving user-facing behavior.
WhenToUse: Use when planning, implementing, or reviewing the Go port end-to-end.
---


# Legacy server analysis and Go architecture

## Executive Summary

The legacy server is an Express + tRPC app that parses quiz forms embedded in markdown, stores documents and submissions in a relational database, and exposes a limited API surface for documents, quiz submissions, and system utilities. The Go port will keep the `/api/trpc` endpoint path for frontend compatibility, move persistence to sqlite, and remove authentication entirely.

The Go architecture should preserve the data model and quiz scoring semantics, keep the tRPC call shape expected by the client, and provide a modular layout for parsing, scoring, and data access. The result is a single Go program that serves the API and (optionally) the static SPA assets.

## Problem Statement

We need a complete Go replacement for the legacy TypeScript server in `legacy-version/server`. The replacement must keep the `/api/trpc` endpoint path, use sqlite instead of MySQL, and operate without any authentication or user session management. The system must preserve the quiz DSL parsing and scoring semantics while remaining compatible with the existing frontend.

## Proposed Solution

### Legacy server inventory (current behavior)

**Service topology**
- Runtime: `legacy-version/server/_core/index.ts` creates an Express app, mounts tRPC at `/api/trpc`, and serves the SPA (Vite in dev; static in prod).
- Context/auth: `createContext` resolves optional user via `sdk.authenticateRequest` (this is removed in the Go port).

**API surface (tRPC routers)**
- `system.health` (public, GET/query): `{ timestamp } -> { ok: true }`
- `system.notifyOwner` (admin, mutation): `{ title, content } -> { success }`
- `auth.me` (public, query): returns `ctx.user`
- `auth.logout` (public, mutation): clears `COOKIE_NAME`
- `documents.list` (public, query): admin gets all, others get published-only
- `documents.myDocuments` (protected, query): list by `authorId`
- `documents.getBySlug` (public, query): slug lookup + access check + forms
- `documents.getById` (protected, query): admin or author only
- `documents.create` (protected, mutation): create doc + extract forms
- `documents.update` (protected, mutation): update doc + re-extract forms
- `documents.delete` (protected, mutation): delete doc + cascade forms/submissions
- `documents.analytics` (protected, query): analytics summary
- `documents.submissions` (protected, query): list submissions + user names
- `quiz.submitMultiple` (protected, mutation): submit multiple forms + scoring
- `quiz.submit` (protected, mutation): submit one form + scoring
- `quiz.mySubmissions` (protected, query): list submissions by user
- `quiz.getSubmission` (protected, query): submission + form definition

**Quiz DSL + parsing**
- Forms embedded in markdown as `<form id="..."> ...yaml... </form>`.
- YAML is parsed and stored in `quiz_forms.definition` as JSON.
- Scoring checks fields under `definition.fields` or `definition.form.fields`:
  - Only fields with `correct` participate in scoring.
  - Checkbox scoring requires an exact match to the full correct set (no partial credit).

**Database schema (legacy)**
- `users`: `openId`, `name`, `email`, `loginMethod`, `role`, `lastSignedIn`.
- `documents`: `title`, `slug`, `content`, `description`, `category`, `isPublished`, `authorId`, timestamps.
- `quiz_forms`: `documentId`, `formId`, `definition` (JSON), timestamps.
- `quiz_submissions`: `userId`, `documentId`, `formId`, `responses` (JSON), `score`, `maxScore`, `submittedAt`.

**Manus services + infra utilities**
- Notification, storage, data API, LLM, image generation, voice transcription, maps (not wired into routers today).

### Decisions applied to the Go port

**Keep `/api/trpc`**
- Preserve the endpoint path to avoid client changes.
- Implement a Go handler that accepts the subset of tRPC requests the frontend uses.
- Maintain response shapes and error codes that the existing frontend expects.

**Use sqlite**
- Replace MySQL/Drizzle with sqlite tables mirroring the legacy schema.
- Store JSON payloads for `quiz_forms.definition` and `quiz_submissions.responses` in sqlite JSON columns or TEXT with JSON encoding.

**Remove authentication**
- Drop OAuth callback, session cookies, and JWT verification.
- Treat all endpoints as public and skip user context in request handling.
- Define a consistent `userId` strategy for writes (fixed ID or nullable values) to keep analytics predictable.

### Go architecture (complete program)

**Service shape**
- `cmd/markdown-quizz`: bootstraps config, database, router, and HTTP server.
- Single binary that serves API and (optionally) static assets.

**Core packages**
- `internal/config`: env parsing (PORT, SQLITE_PATH, etc.).
- `internal/db`: sqlite connection and migrations.
- `internal/trpc`: handler that maps tRPC request envelopes to Go procedures.
- `internal/quizdsl`: markdown form extraction + YAML parsing.
- `internal/scoring`: score calculation for form definitions and response maps.
- `internal/documents`: document CRUD + quiz form persistence.
- `internal/quiz`: submission creation and analytics.
- `internal/system`: health + optional notifications (if still needed).

**API contract strategy**
- Keep `/api/trpc` as the external endpoint, using a Go adapter to match request/response shapes the frontend relies on.
- Provide clear error mapping (unauthorized/forbidden become generic errors or are removed since auth is gone).

**Data model parity**
- Keep schema fields aligned with legacy for minimal frontend changes.
- If `userId` is fixed or nullable, document the decision and update queries accordingly.

**Static hosting**
- Serve SPA assets from the Go binary (optional) or run behind a reverse proxy in production.

### Glazed-based CLI config parsing (new constraint)

Use the Glazed framework to parse CLI flags and config for the Go program, not HTTP request bodies. The port should adopt the new schema/fields/values API (as in `glazed/cmd/examples/refactor-new-packages/main.go`) to define CLI sections and decode CLI settings consistently:

- Define command schemas using `schema.NewSchema` and `schema.NewSection` (add `schema.NewGlazedSchema()` only for commands that emit Glazed rows).
- Declare parameters via `fields.New(...)` with defaults/help/choices.
- Resolve values via `values.Values` and decode into typed structs with `values.DecodeSectionInto`.
- Follow the guidance from `glazed/pkg/doc/tutorials/05-build-first-command.md` for parameter initialization and structured output.

For the port, this means the CLI entrypoint (e.g., `cmd/markdown-quizz`) uses Glazed to parse flags/config (server port, sqlite path, logging, etc.). HTTP request parsing remains normal JSON decoding on the server side.

### Key files and symbols (legacy + references)

**Legacy server entrypoints**
- `legacy-version/server/_core/index.ts`: `startServer`, `setupVite`, `serveStatic`.
- `legacy-version/server/routers.ts`: `appRouter`, `documents.*`, `quiz.*`, `auth.*`, `system.*`.
- `legacy-version/server/db.ts`: document CRUD, `getQuizAnalytics`, submission queries.

**Legacy DSL + schema**
- `legacy-version/DSL_SPECIFICATION.md`: quiz DSL schema and scoring rules.
- `legacy-version/drizzle/schema.ts`: `users`, `documents`, `quiz_forms`, `quiz_submissions`.

**Glazed references**
- `glazed/pkg/doc/tutorials/05-build-first-command.md`: CLI patterns + parameter parsing guidance (layers + `InitializeStruct`).
- `glazed/cmd/examples/refactor-new-packages/main.go`: current wrapper API example (`schema/fields/values`, `values.DecodeSectionInto`).

## Legacy tRPC Contract (what we must match)

The legacy server uses tRPC v10 with `httpBatchLink` and `superjson`. The Go port should implement enough of the HTTP adapter behavior to keep the existing SPA working against `/api/trpc`.

### Procedure inventory (from `legacy-version/server/routers.ts`)

**system**
- `system.health` (public query): input `{ timestamp: number }` → output `{ ok: true }`
- `system.notifyOwner` (admin mutation): input `{ title: string, content: string }` → output `{ success: boolean }`

**auth**
- `auth.me` (public query): output `User | null`
- `auth.logout` (public mutation): output `{ success: true }`

**documents**
- `documents.list` (public query): output `Document[]` (admin gets all; otherwise published-only)
- `documents.myDocuments` (protected query): output `Document[]` (author-only)
- `documents.getBySlug` (public query): input `{ slug: string }` → output `Document & { forms: QuizForm[] }` (with access checks)
- `documents.getById` (protected query): input `{ id: number }` → output `Document`
- `documents.create` (protected mutation): input `{ title, content, description?, category?, isPublished? }` → output `{ id: number, slug: string }`
- `documents.update` (protected mutation): input `{ id, title?, content?, description?, category?, isPublished? }` → output `{ success: true }`
- `documents.delete` (protected mutation): input `{ id: number }` → output `{ success: true }`
- `documents.analytics` (protected query): input `{ id: number }` → output `{ totalSubmissions, averageScore, highestScore, lowestScore }`
- `documents.submissions` (protected query): input `{ id: number }` → output submissions list for that document

**quiz**
- `quiz.submitMultiple` (protected mutation): input `{ documentId: number, submissions: { formId: string, responses: Record<string, any> }[] }` → output `{ results: { formId: string, score: number, maxScore: number }[] }`
- `quiz.submit` (protected mutation): input `{ documentId: number, formId: string, responses: Record<string, any> }` → output `{ id: number, score: number | null, maxScore: number | null }`
- `quiz.mySubmissions` (protected query): output submissions list for current user
- `quiz.getSubmission` (protected query): input `{ id: number }` → output submission + `formDefinition` and access checks

### Transport expectations (from `legacy-version/client/src/main.tsx`)

- Client uses `@trpc/client` `httpBatchLink({ url: "/api/trpc", transformer: superjson })`.
- Client always sends `credentials: "include"` on fetch (cookie auth in legacy; ignored in no-auth mode).
- Client considers `TRPCClientError.message === "Please login (10001)"` as “unauthorized” and redirects to login; the Go port should avoid emitting this legacy auth error string.
- Because batching is enabled, the Go adapter should support batched requests (`?batch=1`) in addition to single-procedure calls.

## Design Decisions

- Preserve `/api/trpc` endpoint path for frontend compatibility.
- Use sqlite for persistence with schema parity to legacy.
- Remove auth entirely and simplify request handling.

## Alternatives Considered

- **Move to explicit REST endpoints**: simpler server but frontend changes required.
- **Keep auth but stub out OAuth provider**: adds complexity without value.
- **Use an in-memory store**: insufficient for durable quiz results.

## Implementation Plan

1. Define sqlite schema equivalents for `users`, `documents`, `quiz_forms`, and `quiz_submissions` in `internal/db/migrations` (or `internal/db/schema.sql`), including JSON storage strategy.
2. Implement sqlite access layer in `internal/db` with explicit query functions matching `legacy-version/server/db.ts` semantics.
3. Build the `/api/trpc` adapter layer in `internal/trpc`, mapping procedure names to handlers with the expected tRPC response envelope.
4. Implement document CRUD + DSL extraction in `internal/documents` and `internal/quizdsl` (form extraction + YAML parse).
5. Implement quiz submissions + scoring + analytics in `internal/quiz` and `internal/scoring`.
6. Implement CLI entrypoint in `cmd/markdown-quizz` using Glazed `schema/fields/values` to parse flags/config (port, sqlite path, static asset dir).
7. Optionally integrate static file serving (e.g., `internal/http/static`) and wire into the main server.

## Open Questions

- Should `userId` be nullable everywhere or fixed to a single row?
  - Recommendation for no-auth mode: create a single synthetic user row (`id=1`, role `admin`) and treat all requests as that user. This keeps `authorId`/`userId` non-null, simplifies queries, and preserves “my*” endpoints as “everything”.
- What exact subset of tRPC semantics does the frontend rely on (batching, superjson, error envelope)?

## References

- `legacy-version/server/routers.ts`
- `legacy-version/server/db.ts`
- `legacy-version/drizzle/schema.ts`
- `legacy-version/DSL_SPECIFICATION.md`
- `uhoh/pkg/doc/topics/uhoh-dsl.md`
- `uhoh/pkg/doc/topics/uhoh-wizard-dsl.md`
- `glazed/pkg/doc/tutorials/05-build-first-command.md`
- `glazed/cmd/examples/refactor-new-packages/main.go`
