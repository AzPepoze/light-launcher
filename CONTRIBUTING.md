# Contributing to LightLauncher

Bug reports, ideas, documentation fixes and pull requests are welcome.

> [!NOTE]
> For the user docs, start at the [README](README.md). This page is for people who change the code.

## Project layout

| Folder              | What is in it                                                         |
| :------------------ | :-------------------------------------------------------------------- |
| `core/cmd/launcher` | The launcher program. It serves the UI and opens the window.          |
| `core/cmd/instance` | The game runner. It starts one game and shows the tray icon.          |
| `core/internal`     | Shared Go code: library, downloads, plugins, API, config and logging. |
| `shell/qt`          | The Linux window program (Qt6 WebEngine).                             |
| `shell/webview2`    | The Windows window program (WebView2).                                |
| `shell/logx`        | A small logger used by the shells.                                    |
| `src/renderer`      | The Svelte 5 UI.                                                      |
| `src/shared`        | TypeScript types and helpers shared by the UI.                        |
| `scripts`           | Build scripts for the Linux and Windows release files.                |
| `install`           | Arch packaging (`PKGBUILD`) and the Windows installer script.         |
| `docs`              | Topic pages and images.                                               |

Read [Architecture](docs/architecture.md) for how the parts connect.

## Prerequisites

| Tool                             | Needed for                    | Notes                                     |
| :------------------------------- | :---------------------------- | :---------------------------------------- |
| Go                               | Everything                    | Version from `go.mod` (currently 1.25.6). |
| Bun                              | UI, scripts, checks           | All `bun run` commands need it.           |
| Qt6 base and WebEngine dev files | Linux window program          | See [Building](docs/building.md).         |
| umu-launcher                     | Running Proton games on Linux | Not needed for tests.                     |
| NSIS, zip                        | Windows release files         | Only for `scripts/build-windows.sh`.      |

## Development setup

```sh
git clone https://github.com/AzPepoze/light-launcher.git
cd light-launcher
bun install
```

### Build the launcher and serve the UI from disk

```sh
make build
./bin/light-launcher --ui src/renderer/dist
```

`make build` builds the launcher, the game runner and the UI. `--ui` serves the UI from the built folder. See [Usage](docs/usage.md#light-launcher).

> [!WARNING]
> `bun run dev` and `make dev` do not work yet. The `dev:server` script starts the launcher with `--dev-token` and `--port`. The launcher does not accept these flags, so it exits with code 2. The Vite proxy in `src/renderer/vite.config.ts` also expects port 9246.

## Scripts

| Command                  | What it does                                                   |
| :----------------------- | :------------------------------------------------------------- |
| `bun run build`          | Builds the Go programs, then the UI.                           |
| `bun run build:go`       | Builds `bin/light-launcher` and `bin/light-launcher-instance`. |
| `bun run build:renderer` | Builds the UI into `src/renderer/dist`.                        |
| `bun run dev:renderer`   | Starts the Vite dev server for the UI.                         |
| `bun run dev:go`         | Builds the launcher for development.                           |
| `bun run test`           | Runs the Vitest tests in `src/`.                               |
| `bun run typecheck`      | Runs `tsc` and `svelte-check`.                                 |
| `bun run lint`           | Runs `oxlint`.                                                 |
| `bun run format`         | Formats the code with `oxfmt`.                                 |
| `bun run format:check`   | Checks formatting without changing files.                      |
| `make build`             | Same as the Go and UI builds, through `Makefile`.              |
| `make clean`             | Removes `bin`, `dist` and the UI build folder.                 |

## Tests and checks

CI runs these checks. Run them before you open a pull request.

```sh
# Go
go vet ./...
go test ./...
(cd shell/logx && go test ./...)

# UI and TypeScript
bun run lint
bun run format:check
bun run typecheck
bun run test

# Window programs
(cd shell/qt && go vet ./... && go build -o ../../bin/light-launcher-shell .)
(cd shell/webview2 && GOOS=windows GOARCH=amd64 go vet ./...)
```

> [!NOTE]
> The Qt build needs the Qt6 dev packages. The WebView2 check only compiles for Windows. It cannot run on Linux.

## Coding rules

- Keep controllers and handlers thin. Put the logic in services and packages.
- Avoid `any` in TypeScript. Keep the Go structs and the types in `src/shared/types` in step.
- Log through the project loggers: `core/internal/logger` in Go, and `src/renderer/src/lib/logger.ts` in the UI. Do not call `console` directly.
- Run `bun run format` before you commit.

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```text
type(optional-scope): short summary
```

| Type       | Use it for                              |
| :--------- | :-------------------------------------- |
| `feat`     | A new feature.                          |
| `fix`      | A bug fix.                              |
| `perf`     | A speed or memory change.               |
| `refactor` | A change that does not change behavior. |
| `docs`     | Documentation only.                     |
| `test`     | Tests only.                             |
| `build`    | Build files or dependencies.            |
| `ci`       | CI workflows.                           |
| `chore`    | Other maintenance.                      |
| `style`    | Formatting only.                        |
| `revert`   | Undo an earlier commit.                 |

Rules:

- Use lowercase, and start the summary with a verb in the imperative mood.
- No period at the end. Keep the line under about 72 characters.
- Add `!` after the type, or a `BREAKING CHANGE:` footer, for a breaking change.
- Make one commit per logical change.
- Do not add `Co-Authored-By` or other attribution lines.

Release versions come from these commit types. See [Releasing](docs/releasing.md#version-bumps).

## Pull requests

1. Fork the repository and create a branch.
2. Make your change and run the checks above.
3. Open a pull request. Keep the body short:

   ```text
   ## Summary
   One or two sentences on what changes and why.

   ## Changes
   - Main change
   - Main change

   ## Checked
   - What you ran, such as go test ./... and bun run test
   ```

Do not add attribution lines, such as `Generated with` or `Co-Authored-By`, to the pull request body.
