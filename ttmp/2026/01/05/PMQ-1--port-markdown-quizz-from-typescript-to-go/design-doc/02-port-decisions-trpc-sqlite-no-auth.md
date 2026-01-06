---
Title: 'Port decisions: tRPC, sqlite, no auth'
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
    - Path: legacy-version/server/routers.ts
      Note: tRPC endpoint behavior driving /api/trpc decision
ExternalSources: []
Summary: Decision log for keeping /api/trpc, switching to sqlite, and removing auth.
LastUpdated: 2026-01-05T19:09:21.409126108-05:00
WhatFor: Lock down porting constraints for the Go migration.
WhenToUse: Use when implementing the Go server or reviewing API parity.
---


# Port decisions: tRPC, sqlite, no auth

## Executive Summary

The Go port will retain the `/api/trpc` endpoint shape, switch persistence to sqlite, and remove authentication entirely. These decisions reduce frontend churn, simplify local deployments, and cut auth-related dependencies from the port.

## Problem Statement

We need to adapt the planned Go architecture to three new constraints: keep the tRPC endpoint path, use sqlite instead of MySQL, and remove all authentication. This requires revising API expectations, data access, and middleware behavior in the Go design.

## Proposed Solution

### Keep `/api/trpc`

- Preserve the endpoint path to avoid client changes.
- Implement a Go handler that accepts the subset of tRPC requests the frontend uses, or proxy/compat adapter if full tRPC semantics are required.
- Maintain response shapes and error codes that the existing frontend expects.

### Switch persistence to sqlite

- Replace MySQL/Drizzle with sqlite tables mirroring the legacy schema.
- Use a Go sqlite driver + query builder/sqlc for CRUD and analytics queries.
- Store JSON payloads for `quiz_forms.definition` and `quiz_submissions.responses` in sqlite JSON columns or TEXT with JSON encoding (define consistent serialization).

### Remove authentication

- Drop OAuth callback, session cookies, and JWT verification.
- Treat all endpoints as public and skip user context in request handling.
- For data that previously required a user ID (documents, submissions), define a simplified rule:
  - Use a fixed user ID (e.g., `1`) or allow nullable `userId` and update schema/queries accordingly.
  - Document the chosen approach so analytics and submissions remain consistent.

## Design Decisions

- **API path**: Keep `/api/trpc` and adapt the Go router to the current tRPC call pattern to avoid frontend changes.
- **Database**: Use sqlite as the primary store, with schema parity to legacy tables wherever possible.
- **Auth**: Remove all auth-related flows; treat endpoints as public and simplify data ownership logic.

## Alternatives Considered

- **Move to explicit REST endpoints**: would simplify Go but requires a frontend update.
- **Keep auth but stub out OAuth provider**: still introduces JWT/cookie complexity for little value.
- **Use an in-memory store**: useful for testing but not durable for real usage.

## Implementation Plan

1. Define sqlite schema equivalents for `users`, `documents`, `quiz_forms`, and `quiz_submissions`.
2. Implement sqlite access layer and update query logic to match legacy behavior.
3. Add a `/api/trpc` handler or adapter layer in Go that matches the frontend’s tRPC usage.
4. Remove auth middleware and refactor handlers to not depend on `ctx.user`.
5. Validate document/quiz flows and analytics against expected behavior.

## Open Questions

- Should `userId` become nullable everywhere or should we standardize on a single user row?
- What minimal subset of tRPC request/response semantics does the frontend rely on (batching, superjson, error format)?

## References

- `ttmp/2026/01/05/PMQ-1--port-markdown-quizz-from-typescript-to-go/design-doc/01-legacy-server-analysis-and-go-architecture.md`
- `legacy-version/server/routers.ts`
