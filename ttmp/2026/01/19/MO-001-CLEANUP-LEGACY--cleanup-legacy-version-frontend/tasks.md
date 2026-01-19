# Tasks

## TODO

- [x] Remove getLoginUrl() OAuth portal integration (VITE_OAUTH_PORTAL_URL/VITE_APP_ID) from the legacy SPA
- [x] Remove redirect-to-login behavior from client/src/main.tsx (React Query cache subscribers)
- [x] Remove useAuth hook usage and all 'Sign In Required' gates/links in pages/layout
- [x] Remove any Manus-specific login UI components (e.g., ManusDialog) if unused after auth removal
- [x] Validate: pnpm dev:ui renders; pnpm build:ui succeeds; key routes load against Go backend (note: build+tsc pass; dev server bind fails here with EPERM)
