---
Title: Diary
Ticket: PMQ-1
Status: active
Topics:
    - backend
    - go
    - porting
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: markdown-quizz/cmd/markdown-quizz/main.go
      Note: Wire Cobra root + Glazed commands
    - Path: markdown-quizz/go.mod
      Note: Introduce dependencies for CLI skeleton
    - Path: markdown-quizz/go.sum
      Note: Lock dependencies after go mod tidy
    - Path: markdown-quizz/internal/cli/commands.go
      Note: Build Cobra subcommands via Glazed
    - Path: markdown-quizz/internal/cli/serve.go
      Note: Define and run initial serve command using schema/fields/values
    - Path: markdown-quizz/legacy-version/DSL_SPECIFICATION.md
      Note: Legacy quiz DSL schema and scoring semantics
    - Path: markdown-quizz/legacy-version/drizzle/schema.ts
      Note: Authoritative data model for users
    - Path: markdown-quizz/legacy-version/server/db.ts
      Note: Database access patterns and analytics logic
    - Path: markdown-quizz/legacy-version/server/routers.ts
      Note: Primary API surface and quiz scoring helpers
    - Path: uhoh/pkg/doc/topics/uhoh-dsl.md
      Note: Uhoh form DSL structure used for comparison
    - Path: uhoh/pkg/doc/topics/uhoh-wizard-dsl.md
      Note: Uhoh wizard DSL and simplified form schema for overlap analysis
ExternalSources: []
Summary: Research log for porting the legacy TypeScript server to Go.
LastUpdated: 2026-01-05T19:02:21.775971654-05:00
WhatFor: Capture analysis steps and decisions while porting markdown-quizz.
WhenToUse: Use during the port to track research and implementation decisions.
---









# Diary

## Goal

Capture the research and analysis steps for porting the markdown-quizz legacy server and aligning DSL behavior with Go and uhoh.

## Step 1: Inventory legacy server behavior and DSL overlap

I reviewed the legacy TypeScript server to understand the current API, auth flow, database model, and quiz scoring behavior. I also surveyed uhoh DSL documentation to compare form/wizard schemas and identify where adapters will be required in the Go port.

This step produced a structured design doc that inventories the current system and proposes a Go architecture that preserves legacy behavior while keeping a path open for uhoh integration.

### What I did
- Read `legacy-version/server` entrypoints, routers, and database access layers to map the API and dependencies.
- Captured quiz DSL parsing and scoring rules from `legacy-version/DSL_SPECIFICATION.md` and router helpers.
- Scanned uhoh DSL documentation and listed relevant docs/specs for future cross-reference.
- Wrote the design document with API inventory, systems map, and proposed Go package layout.

### Why
- Establish a reliable baseline before translating the system into Go and to reduce risk of behavior drift.

### What worked
- The router definitions and schema files provide a compact, complete view of the API and data model.
- Uhoh docs clearly describe form and wizard DSLs, making the gap analysis straightforward.

### What didn't work
- N/A.

### What I learned
- The legacy server already carries a set of Forge-based helpers (LLM, storage, image generation, voice transcription, maps) that are not wired into the routers, so the Go port can treat them as optional adapters rather than core dependencies.

### What was tricky to build
- N/A (research-only step).

### What warrants a second pair of eyes
- Confirm whether the frontend must keep tRPC wire compatibility, or if we can switch to explicit HTTP JSON endpoints.

### What should be done in the future
- Validate whether existing markdown content depends on any undocumented DSL behavior beyond the spec and router helpers.

### Code review instructions
- Start in `ttmp/2026/01/05/PMQ-1--port-markdown-quizz-from-typescript-to-go/design-doc/01-legacy-server-analysis-and-go-architecture.md`.
- Verify that API inventory and DSL comparison match `legacy-version/server/routers.ts` and `legacy-version/DSL_SPECIFICATION.md`.

### Technical details
- Commands used: `rg --files legacy-version/server`, `sed -n '1,240p' legacy-version/server/routers.ts`, `sed -n '1,260p' legacy-version/server/db.ts`, `rg --files -g '*.md' /home/manuel/workspaces/2026-01-05/port-markdown-quizz/uhoh`.

## Step 2: Record port decisions for tRPC, sqlite, and no auth

I captured the new porting constraints: keep `/api/trpc`, switch persistence to sqlite, and remove authentication. This updates the Go architecture assumptions and clarifies how we should simplify user context and database handling going forward.

The decisions are documented in a new design doc so the implementation plan can align with the frontend’s expectations and the reduced auth surface area.

### What I did
- Added a design doc documenting the tRPC, sqlite, and no-auth decisions.
- Outlined the impact on API routing, data model, and user handling assumptions.

### Why
- Ensure the Go port aligns with updated product constraints and avoids drifting into unused auth infrastructure.

### What worked
- The decisions map cleanly onto the existing legacy API inventory and make a smaller, more direct backend scope.

### What didn't work
- N/A.

### What I learned
- Keeping `/api/trpc` while removing auth requires explicit guidance on how to represent `userId` in documents and submissions.

### What was tricky to build
- N/A (documentation-only step).

### What warrants a second pair of eyes
- Confirm the frontend’s exact tRPC transport assumptions (batching, superjson, error envelopes) before implementing the adapter in Go.

### What should be done in the future
- Decide on the canonical `userId` strategy (nullable vs fixed row) and update schema/queries accordingly before implementation.

### Code review instructions
- Start in `ttmp/2026/01/05/PMQ-1--port-markdown-quizz-from-typescript-to-go/design-doc/02-port-decisions-trpc-sqlite-no-auth.md`.
- Check that the decisions are consistent with the earlier legacy inventory doc.

### Technical details
- Document created: `ttmp/2026/01/05/PMQ-1--port-markdown-quizz-from-typescript-to-go/design-doc/02-port-decisions-trpc-sqlite-no-auth.md`.

## Step 3: Expand the full-port design doc to include new constraints

I revised the main design document to cover the complete Go port with the newly confirmed constraints: keep `/api/trpc`, move to sqlite, and remove authentication. This keeps a single end-to-end design reference aligned with current scope and avoids splitting critical architecture details across multiple docs.

The updated doc now presents the full architecture for the Go program, including revised API assumptions and data model notes tied to the no-auth decision.

### What I did
- Rewrote the primary design doc to incorporate the new constraints and present a complete port plan.
- Consolidated the architecture narrative so implementation has one canonical source.

### Why
- The main design doc should reflect the entire Go program with current constraints, not just the legacy inventory.

### What worked
- The existing legacy inventory sections slotted cleanly into the revised, full-port structure.

### What didn't work
- N/A.

### What I learned
- The most sensitive risk is the tRPC adapter behavior; documenting the expected subset upfront will prevent surprises during implementation.

### What was tricky to build
- N/A (documentation-only step).

### What warrants a second pair of eyes
- Confirm that the revised design doc fully supersedes any older assumptions about auth and database choice.

### What should be done in the future
- Decide on the `userId` strategy (nullable vs fixed) before schema implementation.

### Code review instructions
- Start in `ttmp/2026/01/05/PMQ-1--port-markdown-quizz-from-typescript-to-go/design-doc/01-legacy-server-analysis-and-go-architecture.md`.
- Check that the final architecture description aligns with the new constraints and references.

### Technical details
- Document updated: `ttmp/2026/01/05/PMQ-1--port-markdown-quizz-from-typescript-to-go/design-doc/01-legacy-server-analysis-and-go-architecture.md`.

## Step 6: Implement initial Glazed-based `serve` CLI skeleton

I added the first functional CLI entrypoint for the Go port, using Glazed’s current `schema/fields/values` API for flag/config decoding. The goal is to lock in the CLI parsing pattern early so future server/bootstrap work can focus on HTTP behavior and persistence, not argument parsing details.

This step wires a `serve` command that starts a minimal HTTP server with a `/healthz` endpoint and a placeholder `/api/trpc` handler, while already decoding the intended server settings (host/port/sqlite path/static dir/log level) from Glazed.

### What I did
- Implemented `cmd/markdown-quizz/main.go` root command and wired Glazed-built subcommands.
- Added `internal/cli/serve.go` defining the `serve` command schema via `schema.NewSection` + `fields.New(...)` (no `schema.NewGlazedSchema()` for this bare server command) and decoding values with `values.DecodeSectionInto`.
- Added a minimal `net/http` server with graceful shutdown via `signal.NotifyContext` and `errgroup`.
- Ran `go mod tidy`, `gofmt`, and `go test ./...` in the `markdown-quizz` module.

### Why
- We need a stable, repeatable CLI/config surface to support the upcoming server implementation tasks (db path, port binding, static assets, logging).

### What worked
- The `refactor-new-packages` pattern (`schema/fields/values`) maps cleanly to typed config structs for the server.
- The Glazed Cobra parser config with `AppName: "markdown_quizz"` enables env/config layering without custom glue code.

### What didn't work
- N/A.

### What I learned
- `values.Values` is a thin alias over `layers.ParsedLayers`, so decoding helpers can be used without changing the underlying Glazed command interfaces.

### What was tricky to build
- Choosing how to structure the schema so future commands can share config (single `server` section now, but may need refactoring into shared sections later).

### What warrants a second pair of eyes
- Confirm the desired flag names and section prefixing strategy before we build more commands (to avoid churn in env/config variable naming).

### What should be done in the future
- Wire the decoded `sqlite-path`, `static-dir`, and `log-level` into the real server bootstrap once those components exist.

### Code review instructions
- Start in `markdown-quizz/internal/cli/serve.go` and verify the schema/decoding pattern matches `glazed/cmd/examples/refactor-new-packages/main.go`.
- Validate behavior by running `go test ./... -count=1` and `go run ./cmd/markdown-quizz serve --help`.

### Technical details
- Commands used: `go mod tidy`, `gofmt -w cmd/markdown-quizz/main.go internal/cli/commands.go internal/cli/serve.go`, `go test ./... -count=1`.

## Step 4: Add Glazed CLI parsing guidance to the full-port design

I corrected the design doc to use Glazed for CLI flag/config parsing only (not HTTP requests), based on the build-first-command guidance. This keeps Glazed scoped to the CLI entrypoint while leaving request parsing to standard HTTP JSON decoding.

The design now points to the relevant Glazed references and aligns the usage patterns with the new schema/fields/values API.

### What I did
- Updated the primary design doc with a Glazed parsing section tied to the tRPC adapter.
- Referenced Glazed guidance from `glazed/pkg/doc/tutorials/05-build-first-command.md` and the new wrapper API usage from `glazed/cmd/examples/refactor-new-packages/main.go`.

### Why
- The port should align with the current Glazed framework patterns and avoid legacy parsing paths.

### What worked
- The Glazed API maps cleanly onto the requirement to parse tRPC envelopes into typed request structs.

### What didn't work
- N/A.

### What I learned
- Glazed should be used for CLI flags/config parsing, not for HTTP request bodies.

### What was tricky to build
- N/A (documentation-only step).

### What warrants a second pair of eyes
- Confirm the CLI settings we want to expose via Glazed before implementing the CLI entrypoint.

### What should be done in the future
- Finalize the CLI parameter list (port, sqlite path, logging, static assets).

### Code review instructions
- Start in `ttmp/2026/01/05/PMQ-1--port-markdown-quizz-from-typescript-to-go/design-doc/01-legacy-server-analysis-and-go-architecture.md`.
- Check the Glazed CLI parsing section for alignment with the `refactor-new-packages` example.

### Technical details
- References reviewed: `glazed/pkg/doc/tutorials/05-build-first-command.md`, `glazed/cmd/examples/refactor-new-packages/main.go`.

## Step 5: Add precise file/symbol references and tighten the plan

I expanded the design doc with explicit file paths, symbol pointers, and a more concrete implementation plan. This makes the document actionable for the full port, with clear entrypoints into the legacy code and the Glazed API references.

The plan now enumerates target files and packages for each milestone so implementation can proceed without guessing where to place new code.

### What I did
- Added a “Key files and symbols” section with legacy entrypoints and Glazed references.
- Tightened the implementation plan with explicit package/file targets.
- Extended references to include the Glazed help and refactor example.

### Why
- The design doc should be immediately actionable and point to concrete locations and symbols.

### What worked
- The legacy server structure maps cleanly to the Go package layout with explicit file targets.

### What didn't work
- N/A.

### What I learned
- The refactor example is the clearest reference for the new Glazed schema/values API.

### What was tricky to build
- N/A (documentation-only step).

### What warrants a second pair of eyes
- Confirm the tRPC response envelope expectations before implementing `internal/trpc`.

### What should be done in the future
- Decide where migrations will live (`internal/db/migrations` vs a single schema file).

### Code review instructions
- Start in `ttmp/2026/01/05/PMQ-1--port-markdown-quizz-from-typescript-to-go/design-doc/01-legacy-server-analysis-and-go-architecture.md`.
- Review the “Key files and symbols” and “Implementation Plan” sections for completeness.

### Technical details
- Document updated: `ttmp/2026/01/05/PMQ-1--port-markdown-quizz-from-typescript-to-go/design-doc/01-legacy-server-analysis-and-go-architecture.md`.
