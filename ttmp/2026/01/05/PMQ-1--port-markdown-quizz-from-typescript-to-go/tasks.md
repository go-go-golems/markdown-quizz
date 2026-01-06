# Tasks

## TODO

- [ ] Add tasks here

- [x] Inventory legacy API endpoints and tRPC procedure names from legacy-version/server/routers.ts
- [x] Record tRPC transport expectations (batching, superjson, error envelope) from the frontend client
- [ ] Decide userId strategy for no-auth mode (nullable vs fixed user row) and document in design doc
- [x] Define sqlite schema (users/documents/quiz_forms/quiz_submissions) with JSON storage strategy
- [x] Choose sqlite driver and migration approach (e.g., modernc/sqlite + goose or pure schema.sql)
- [x] Implement sqlite connection/init module with connection lifecycle and migrations
- [x] Implement quiz DSL extraction: find <form id=...> blocks and parse YAML into JSON
- [x] Implement quiz scoring logic (exact checkbox match, fields in definition.fields or definition.form.fields)
- [x] Implement document CRUD queries (create/update/delete/get/list) with sqlite and keep slug generation parity
- [x] Implement quiz_forms persistence and re-extraction on document updates
- [x] Implement quiz submission creation (single + multiple) and scoring persistence
- [x] Implement analytics aggregation (total submissions, avg/high/low score) for a document
- [x] Implement submissions listing (by document and by user) and single submission fetch with form definition
- [x] Implement /api/trpc adapter: procedure routing, input decoding, and response envelope formatting
- [x] Implement system.health endpoint in Go handler (public)
- [ ] Decide fate of system.notifyOwner in no-auth mode (keep/remove) and implement if needed
- [ ] Implement CLI entrypoint with Glazed schema/fields/values for server config (port, sqlite path, static dir, log level)
- [ ] Implement HTTP server bootstrap with router wiring, middleware, and graceful shutdown
- [x] Integrate static file serving for SPA (configurable asset dir)
- [ ] Port frontend API client (if needed) to match Go tRPC adapter assumptions
- [ ] Add unit tests for quiz DSL parsing and scoring (checkbox exact match, radio/text correctness)
- [ ] Add unit tests for document CRUD (create/update/delete/list/get) and quiz form persistence
- [ ] Add integration tests for /api/trpc adapter (documents.create/update, quiz.submitMultiple, analytics)
- [ ] Document deployment/run instructions (CLI flags, sqlite file location, static assets)
