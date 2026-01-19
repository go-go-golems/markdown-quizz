# Changelog

## 2026-01-19

- Initial workspace created


## 2026-01-19

Remove legacy Manus OAuth/login flow from the SPA; remove login redirects and auth gates; keep UI runnable in no-auth mode (build+tsc pass; vite dev server bind fails in this environment with EPERM).

### Related Files

- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/legacy-version/client/src/components/DashboardLayout.tsx — Removed login gate
- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/legacy-version/client/src/main.tsx — Removed redirect-to-login logic
- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/legacy-version/client/src/pages/Admin.tsx — Removed sign-in gate
- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/legacy-version/client/src/pages/Home.tsx — Removed login UI
- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/ttmp/2026/01/19/MO-001-CLEANUP-LEGACY--cleanup-legacy-version-frontend/reference/01-diary.md — Detailed diary of investigation and changes


## 2026-01-19

Hygiene: ignore local sqlite artifacts; seed docmgr vocabulary for ticket metadata; mark validation task complete (dev server bind EPERM noted).

### Related Files

- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/.gitignore — Ignore sqlite artifacts
- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/ttmp/2026/01/19/MO-001-CLEANUP-LEGACY--cleanup-legacy-version-frontend/reference/01-diary.md — Diary step recording validation status
- /home/manuel/workspaces/2026-01-05/port-markdown-quizz/markdown-quizz/ttmp/vocabulary.yaml — Seed vocabulary entries

