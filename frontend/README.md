# webhook-middleware dashboard

A Next.js (App Router) + shadcn/ui analytics dashboard for the
`webhook-middleware` Go backend. Every menu talks directly to the backend's
`/dashboard/webhooks*` REST API — there is no separate database or server
component here, this is a pure client that renders what the backend reports.

## Stack

- **Next.js 14** (App Router, TypeScript)
- **Tailwind CSS** + hand-rolled **shadcn/ui** primitives (`components/ui/*`,
  new-york style, CSS-variable theme, light/dark mode via `next-themes`)
- **TanStack Query** for data fetching/caching/mutations
- **Recharts** for the analytics charts
- **sonner** for toasts, **lucide-react** for icons

## Pages / menu → backend mapping

| Menu       | Route         | Backend endpoints used                                                   |
|------------|---------------|----------------------------------------------------------------------------|
| Overview   | `/`           | `GET /dashboard/webhooks/stats`, `GET /dashboard/webhooks` (recent), `POST /dashboard/webhooks/retry-failed` |
| Webhooks   | `/webhooks`   | `GET /dashboard/webhooks` (filters+pagination), `GET /dashboard/webhooks/:id`, `POST /dashboard/webhooks/:id/replay`, `POST /dashboard/webhooks/retry-failed`, `GET /dashboard/webhooks/sources` (filter dropdown) |
| Sources    | `/sources`    | `GET /dashboard/webhooks/sources`                                          |
| Settings   | `/settings`   | `GET /healthz` (connection test)                                           |

Every endpoint reply is wrapped in the backend's standard envelope
(`{ success, message, data, meta }`) — see `lib/api.ts` / `lib/types.ts`.

## Getting started

```bash
cd frontend
cp .env.local.example .env.local   # adjust if your backend isn't on :8090
npm install
npm run dev
```

Open http://localhost:3000. The dashboard defaults to
`http://localhost:8090` for the API base URL; change it any time from the
**Settings** page (stored in `localStorage`, no rebuild needed) — handy for
pointing the same static build at staging/production backends.

If the backend has `DASHBOARD_API_KEY` set, put the same value in
Settings → `X-API-Key` (or `NEXT_PUBLIC_API_KEY` in `.env.local`); it's sent
as the `X-API-Key` header on every `/dashboard/*` request.

> The backend already enables `middleware.CORS()` for all origins, so no
> proxy/rewrite is needed to call it from `localhost:3000`.

## Scripts

- `npm run dev` — start the dev server
- `npm run build && npm run start` — production build
- `npm run lint` — ESLint (next/core-web-vitals)

## Structure

```
app/(dashboard)/        - layout (sidebar+topbar) + the 4 menu pages
components/ui/           - shadcn/ui primitives (button, card, table, sheet, …)
components/charts/       - Recharts wrappers themed to the design tokens
components/webhooks/     - filters, table, detail side-panel
hooks/use-webhooks.ts    - TanStack Query hooks for every backend endpoint
lib/api.ts               - typed fetch client + envelope handling
lib/types.ts              - TS mirrors of the Go domain/repository JSON shapes
lib/settings.ts           - localStorage-backed API base URL / API key
```
