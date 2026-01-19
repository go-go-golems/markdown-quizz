# Changelog

## 2026-01-19

- Initial workspace created


## 2026-01-19

Add systematic report for porting tRPC+React Query to typed REST+RTK Query (API inventory + migration plan + tasks).

### Related Files

- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/ttmp/2026/01/19/PMQ-2--port-trpc-react-query-to-rest-rtk-query/design-doc/01-port-plan-trpc-react-query-typed-rest-rtk-query.md — Detailed plan and alternatives
- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/ttmp/2026/01/19/PMQ-2--port-trpc-react-query-to-rest-rtk-query/reference/01-api-inventory-trpc-rest-mapping.md — Endpoint-by-endpoint mapping table
- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/ttmp/2026/01/19/PMQ-2--port-trpc-react-query-to-rest-rtk-query/tasks.md — Migration task checklist


## 2026-01-19

Clarify that PMQ-2 is a big-bang cutover: no migration/backwards-compatibility; REST+RTK replaces tRPC+React Query in one rip-out.

### Related Files

- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/ttmp/2026/01/19/PMQ-2--port-trpc-react-query-to-rest-rtk-query/design-doc/01-port-plan-trpc-react-query-typed-rest-rtk-query.md — Updated to explicit no-compat cutover
- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/ttmp/2026/01/19/PMQ-2--port-trpc-react-query-to-rest-rtk-query/reference/01-api-inventory-trpc-rest-mapping.md — Updated to reflect big-bang cutover
- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/ttmp/2026/01/19/PMQ-2--port-trpc-react-query-to-rest-rtk-query/tasks.md — Updated task language to no-compat rip-out


## 2026-01-19

Expand tasks.md into an intern-ready, step-by-step OSHA-style checklist for the big-bang REST+RTK cutover.

### Related Files

- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/ttmp/2026/01/19/PMQ-2--port-trpc-react-query-to-rest-rtk-query/tasks.md — Detailed actionable task breakdown


## 2026-01-19

Step 1: establish baseline (Go per-module tests; legacy TS tests fail due to DB not available) (commit 9865e87)

### Related Files

- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/ttmp/2026/01/19/PMQ-2--port-trpc-react-query-to-rest-rtk-query/reference/02-diary.md — Record baseline commands + failures

## 2026-01-19

Step 2: add Go REST /api handlers + contract tests (commit 35f8f75)

### Related Files

- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/internal/cli/serve.go — Mount /api REST handler
- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/internal/rest/server.go — REST endpoint routing + JSON shapes
- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/internal/rest/server_test.go — Contract tests for REST endpoints

