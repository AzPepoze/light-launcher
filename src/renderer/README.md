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

| Step   | What happens                                     |
| ------ | ------------------------------------------------ |
| Build  | Builds the Go launcher into `bin/light-launcher` |
| Server | Starts it with `--port 9246 --dev-token dev`     |
| UI     | Starts the Vite dev server on port 9245          |

Open `http://localhost:9245/#t=dev` in a browser. Vite forwards `/api` to the Go server.

> [!IMPORTANT]
> The Go launcher must accept `--port` and `--dev-token`. Until it does, `bun run dev` fails at the server step.

## Scripts

| Script                   | What it does                                            |
| ------------------------ | ------------------------------------------------------- |
| `bun run dev`            | Build the Go launcher, then start the UI and the server |
| `bun run build`          | Build the Go binaries and the UI                        |
| `bun run build:renderer` | Build only the UI into `src/renderer/dist`              |
| `bun run typecheck`      | Check TypeScript and Svelte types                       |
| `bun run lint`           | Run oxlint                                              |
| `bun run format:check`   | Check formatting with oxfmt                             |
| `bun run test`           | Run the unit tests with vitest                          |
