# APIHub Frontend

React + TypeScript + Vite dashboard for APIHub.

## Run

```bash
npm install
npm run dev
```

PowerShell may block `npm.ps1` depending on local execution policy. Use this form if needed:

```powershell
npm.cmd run dev
```

The dev server runs at `http://localhost:5173` and proxies `/api` to the Go backend at `http://localhost:8080`.

## Build

```bash
npm.cmd run build
```

Current status on 2026-05-19: build passes, with a Vite chunk-size warning.

## Lint

```bash
npm.cmd run lint
```

Current status on 2026-05-19: lint fails. Known errors are tracked in the root [ROADMAP.md](../ROADMAP.md).

## Notes

- The root Dockerfile builds `frontend/dist`, but the Go backend does not yet serve it.
- Use the Vite dev server for the full UI until static serving is added.
- Auth state is stored client-side through `src/lib/auth.ts`; backend auth is controlled by `APIHUB_AUTH_ENABLED`.
