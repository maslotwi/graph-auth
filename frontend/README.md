# Graph Auth Frontend

React 19 + TypeScript + Vite + shadcn/ui, managed with Bun.

## Development

```bash
bun install
bun run dev
```

The Vite dev server proxies `/api/*` to the Go backend at `http://localhost:8080`.

## Environment variables

| Variable | File | Purpose |
|----------|------|---------|
| `VITE_API_BASE_URL` | `.env.development` | Leave empty to use same-origin requests via the dev proxy |
| `VITE_USE_MOCKS` | `.env.development` | `true` enables in-app fetch mocks; `false` uses the real Go backend |
| `VITE_API_BASE_URL` | `.env.production` | Set to the production API origin when deployed separately |

Use `apiUrl("/api/...")` from `src/lib/api.ts` or the modules in `src/api/` for typed requests.

## MVP flow (mock or real API)

1. `/register` — email-only signup
2. `/check-email` — confirmation screen
3. `/verify?token=...` — email link handler
4. `/setup/root` — create root graph node with permissions
5. `/` — authenticated dashboard

## Roadmap

| Phase | Feature | Key deps |
|-------|---------|----------|
| 2 | Graph visualization | `@neo4j-nvl/react`, `GET /api/graph` |
| 3 | QR add child node | QR scanner, `POST /api/nodes/pair` |
| 4 | QR SSO to external apps | QR display + auth polling |
| 5 | Permissions editor + logout cascade | Predecessor permission caps, `POST /api/nodes/:id/logout` |

## Adding components

To add components to your app, run the following command:

```bash
bunx --bun shadcn@latest add button
```

This will place the ui components in the `src/components` directory.

## Using components

To use the components in your app, import them as follows:

```tsx
import { Button } from "@/components/ui/button"
```
