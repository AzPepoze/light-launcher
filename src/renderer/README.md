# Renderer

The Svelte UI of Light Launcher. It runs in a browser page served by the Go launcher and talks to it over HTTP. Back to the [main README](../../README.md).

## How it talks to Go

| Part          | Path               | Use                                                         |
| ------------- | ------------------ | ----------------------------------------------------------- |
| Call a method | `POST /api/invoke` | Every page action, such as `GetAllGames`                    |
| Live events   | `GET /api/events`  | Progress and session updates, as a server-sent event stream |
| Ready ping    | `POST /api/ready`  | Sent once the page has loaded                               |

Every request sends the token in the `X-LL-Token` header. The token comes from the URL fragment (`#t=...`). The page reads it once and removes it from the address bar.

All calls go through `src/lib/transport.ts`. Pages import from `src/lib/api/index.ts`.

## Develop

From the repository root:

```sh
bun run dev
```

| Step   | What happens                                          |
| ------ | ----------------------------------------------------- |
| Server | Runs the launcher on port 9246 with `--dev-token dev` |
| UI     | Starts the Vite dev server on port 9245               |

`bun run dev` opens the native window on the Vite page: GTK on Linux, WebView2 on Windows. Use `bun run dev:linux` or `bun run dev:windows` to pick one.

For a browser, run `bun run dev:browser` and open `http://localhost:9245/#t=dev`. Vite forwards `/api` to the Go server.

The launcher serves the API and reads its UI from `src/renderer`; Vite serves the page itself.

## Scripts

| Script                   | What it does                                 |
| ------------------------ | -------------------------------------------- |
| `bun run build:renderer` | Build only the UI into `src/renderer/dist`   |
| `bun run dev`            | Detect the OS and open the native dev window |
| `bun run dev:browser`    | Start the API and Vite in the browser        |
| `bun run dev:linux`      | Build and open the GTK window                |
| `bun run dev:windows`    | Build and open the WebView2 window           |
| `bun run format:check`   | Check formatting with oxfmt                  |
| `bun run lint:ui`        | Run oxlint                                   |
| `bun run test`           | Run the UI and Go tests                      |
| `bun run test:ui`        | Run the unit tests with vitest               |
| `bun run typecheck`      | Check TypeScript and Svelte types            |
