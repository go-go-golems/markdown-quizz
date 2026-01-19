---
Title: 'Backend gap analysis: legacy server vs Go backend'
Ticket: PMQ-1
Status: active
Topics:
    - backend
    - go
    - porting
    - legacy
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: markdown-quizz/internal/cli/serve.go
      Note: Go server entrypoint; sqlite path and /api/trpc mounting
    - Path: markdown-quizz/internal/documents/store.go
      Note: Go documents store; must match legacy behavior for CRUD + analytics/submissions.
    - Path: markdown-quizz/internal/quiz/store.go
      Note: Go quiz store; must match legacy behavior for submissions and scoring.
    - Path: markdown-quizz/internal/trpc/server.go
      Note: |-
        Go tRPC HTTP adapter; must preserve the procedures and envelopes used by the SPA.
        Go tRPC HTTP adapter; procedure mapping
    - Path: markdown-quizz/legacy-version/client/src/main.tsx
      Note: SPA bootstrap for tRPC client; should target Go backend via /api proxy.
    - Path: markdown-quizz/legacy-version/package.json
      Note: Defines dev scripts; legacy server is reference-only
    - Path: markdown-quizz/legacy-version/server/_core/sdk.ts
      Note: |-
        Legacy OAuth + session cookie verification; logs OAUTH_SERVER_URL errors when misconfigured.
        Source of the OAUTH_SERVER_URL warning and session cookie auth behavior
    - Path: markdown-quizz/legacy-version/server/db.ts
      Note: |-
        Legacy DB layer; throws \"Database not available\" when DATABASE_URL is unset.
        Source of 'Database not available' when DATABASE_URL is missing
    - Path: markdown-quizz/legacy-version/server/routers.ts
      Note: Canonical legacy tRPC router; defines procedures the UI expects.
    - Path: markdown-quizz/legacy-version/vite.config.ts
      Note: |-
        Vite dev proxy configuration that should point /api to the Go server.
        Proxy config that should send /api to the Go backend
ExternalSources: []
Summary: Compare legacy Express+tRPC server to the new Go backend; explain the \"Database not available\" and OAuth env errors; list remaining gaps and recommended improvements.
LastUpdated: 2026-01-19T00:00:00Z
WhatFor: Clarify what backend work is still needed (vs what is just legacy-server configuration), and provide a prioritized improvement plan.
WhenToUse: Use when debugging frontend API failures, planning parity work, or deciding what parts of the legacy server can be deleted.
---


# Backend gap analysis: legacy server vs Go backend

## Executive Summary

If you see `Database not available` or `[OAuth] ERROR: OAUTH_SERVER_URL is not configured`, you are hitting the *legacy Express server* (`legacy-version/server/`) rather than the Go backend. Those errors are expected in legacy mode unless you configure legacy-only env vars (MySQL + OAuth service).

The Go backend is intended to be the runtime backend now. It uses SQLite, does not require OAuth, and exposes `/api/trpc` with the subset of procedures the SPA uses. The work remaining is less about “missing DB wiring” and more about tightening parity, tests, and removing legacy-only dependencies (including the legacy server itself as a runtime).

## Problem Statement

We have two backends in the repo:

1. **Legacy server** (`legacy-version/server/`): Express + tRPC + Drizzle(MySQL) + Manus OAuth integration.
2. **Go backend** (`markdown-quizz/`): Glazed CLI + SQLite + Go tRPC adapter at `/api/trpc`.

The frontend can be run in at least two ways:

- **Correct for Go backend**: run the SPA dev server (`pnpm dev` or `pnpm dev:ui`) and proxy `/api` to the Go server.
- **Legacy mode**: run the legacy Express server (`pnpm dev:legacy-server`), which serves the SPA and handles `/api/trpc` itself.

The confusing failures happen when you accidentally run legacy mode without legacy env configuration.

## Symptom Analysis (Why you see these errors)

### “Database not available” on Save

Source: `legacy-version/server/db.ts`.

Legacy DB access is lazy and only initializes if `process.env.DATABASE_URL` is set. If it is missing, `getDb()` returns `null` and many functions throw `Error("Database not available")`. This is a legacy-server-only failure mode.

Implication: this does **not** indicate the Go backend is missing “DB plumbing”. It indicates the request is being handled by the legacy server without MySQL configured.

### “[OAuth] ERROR: OAUTH_SERVER_URL is not configured”

Source: `legacy-version/server/_core/sdk.ts`.

The legacy server constructs an OAuth client during startup; it logs an error if `process.env.OAUTH_SERVER_URL` is unset. Again, this is legacy-server-only.

Implication: you are running legacy mode (or importing legacy OAuth codepaths) even though the current direction is “no auth”.

## Legacy Server: What it provides (and what it depends on)

### Runtime architecture

- Express server sets up:
  - OAuth callback route under `/api/oauth/callback`
  - tRPC adapter under `/api/trpc`
  - Vite middleware (dev) or static assets (prod)
- tRPC context optionally authenticates requests via `sdk.authenticateRequest()` which verifies a session cookie.

### Data layer

- MySQL via Drizzle.
- Schema includes `users`, `documents`, `quiz_forms`, `quiz_submissions` (plus roles/admin logic).
- Requires `DATABASE_URL` to function at all.

### Auth

- OAuth portal integration and JWT session cookies.
- Requires `OAUTH_SERVER_URL`, `VITE_APP_ID`, and cookie/JWT configuration.

### Non-core extras (may be unused by the current pages)

The legacy server contains modules for things like storage, maps, image generation, and LLM helpers. These should be treated as optional until proven required by the UI.

## Go Backend: What it provides today

### Runtime architecture

- Glazed CLI `serve` runs an HTTP server on `127.0.0.1:9092` by default.
- SQLite file created/migrated automatically (default `markdown-quizz.sqlite`).
- `/api/trpc` implemented by a Go adapter that matches the frontend’s tRPC v11 calling pattern (batch, `input=` on GET, superjson wrapper unwrapping).
- Optional static SPA serving when `--static-dir` is set.

### Data layer

- SQLite schema and migrations in `internal/db/migrations`.
- Stores documents, quiz forms (definitions), and submissions; computes analytics.

### Auth

- Deliberately removed for the Go port. The Go server currently behaves as a synthetic admin user for the tRPC endpoints it implements.

## Gap & Improvement Assessment (Go backend vs legacy)

### A. Critical to correct runtime usage (highest priority)

1. **Make it hard to accidentally run legacy mode**
   - If the SPA is intended to be driven by the Go server, default commands should not start the legacy server.
   - Recommendation: keep legacy server behind an explicit command (`pnpm dev:legacy-server`) and use Vite for `pnpm dev`.

2. **Add a “which backend am I hitting?” debug affordance**
   - Recommendation: add a small health endpoint and/or banner in the SPA showing API base and server signature (e.g. response header or `/healthz` call) to reduce confusion.

### B. Parity correctness & safety (important)

1. **tRPC envelope + error consistency**
   - Ensure the Go adapter returns errors the SPA can handle (codes/messages).
   - Recommendation: add integration tests that send real tRPC requests from recorded frontend patterns.

2. **Data/ownership invariants**
   - Legacy has users/roles and access checks; Go currently simplifies to a single user.
   - Recommendation: explicitly document which invariants are intentionally dropped and what the future auth model would be (if any).

3. **system.notifyOwner**
   - Legacy implements actual notification delivery; Go currently stubs.
   - Recommendation: decide whether this is required for this product; implement or remove from UI.

### C. Developer experience (nice-to-have)

1. **Embed the SPA in the Go binary**
   - Recommendation: switch from `--static-dir` to `go:embed` build pipeline (optional), so deployment is single-binary.

2. **Reduce legacy surface area**
   - Recommendation: move the shared `AppRouter` type out of `legacy-version/server/` so we can delete the legacy server runtime code.

## Practical Guidance: How to run “the Go version” correctly

1. Start the Go backend:
   - `go run ./cmd/markdown-quizz serve --static-dir legacy-version/dist/public`
2. Start the SPA dev server (no legacy server):
   - `pnpm -C legacy-version dev` (or `pnpm -C legacy-version dev:ui`)

If you instead run `pnpm -C legacy-version dev:legacy-server`, you will hit the legacy server and must configure `DATABASE_URL`, `OAUTH_SERVER_URL`, etc. That mode is for reference only.

## Open Questions

- Do we want *any* auth in the Go backend for non-dev deployments, or is “single-user local tool” the target?
- Should the Go backend preserve legacy access rules (`publishedOnly`, admin-only analytics/submissions), or simplify them permanently?
- Should we remove legacy-only code immediately, or keep it until the Go parity is fully covered by integration tests?
