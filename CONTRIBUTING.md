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
| `shell/gtk`         | The Linux window program (WebKitGTK 6.0 and GTK4).                    |
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
| Bash                             | Local build scripts           | Git Bash or WSL on Windows.               |
| GTK4 and WebKitGTK 6.0 dev files | Linux window program          | See [Building](docs/building.md).         |
| umu-launcher                     | Running Proton games on Linux | Not needed for tests.                     |
| NSIS, zip                        | Windows release files         | Only for `scripts/build-windows.sh`.      |

## Development setup

```sh
git clone https://github.com/AzPepoze/light-launcher.git
cd light-launcher
bun install
```

### Run the UI in development

```sh
bun run dev
```

Open `http://localhost:9245/#t=dev`. Vite serves the page with live reload and proxies API calls to the launcher on port 9246.

### Build locally

```sh
bun run build
./bin/light-launcher
```

`bun run build` builds the launcher with the UI embedded and the game runner. On Linux, `bun run build:linux` also builds the GTK window. To run the UI from a different folder, pass `--ui` to the launcher. See [Usage](docs/usage.md#light-launcher).

> [!NOTE]
> The local build scripts use Bash. On Windows, run them from Git Bash or WSL, or use `bun run dev` for development.

## Scripts

| Command                  | What it does                                                  |
| :----------------------- | :------------------------------------------------------------ |
| `bun run build`          | Builds the launcher with the embedded UI and the game runner. |
| `bun run build:instance` | Builds the game runner.                                       |
| `bun run build:launcher` | Builds the launcher with the embedded UI.                     |
| `bun run build:linux`    | Builds the local Linux app, including the GTK window.         |
| `bun run build:renderer` | Builds the UI into `src/renderer/dist`.                       |
| `bun run build:shell`    | Builds the GTK window (Linux).                                |
| `bun run clean`          | Removes `bin`, `dist` and the UI build folder.                |
| `bun run dev`            | Starts the launcher API and the Vite UI.                      |
| `bun run dev:renderer`   | Starts only the Vite UI.                                      |
| `bun run dev:server`     | Starts only the launcher API.                                 |
| `bun run e2e`            | Runs the launcher API end-to-end test.                        |
| `bun run format`         | Formats the code with `oxfmt`.                                |
| `bun run format:check`   | Checks formatting without changing files.                     |
| `bun run lint`           | Runs `oxlint` and `go vet`.                                   |
| `bun run lint:fix`       | Fixes lint findings with `oxlint`.                            |
| `bun run lint:go`        | Runs `go vet` on the Go code.                                 |
| `bun run lint:ui`        | Runs `oxlint`.                                                |
| `bun run test`           | Runs the UI and Go tests.                                     |
| `bun run test:coverage`  | Runs the UI tests with coverage.                              |
| `bun run test:go`        | Runs the Go tests.                                            |
| `bun run test:ui`        | Runs the Vitest tests.                                        |
| `bun run test:watch`     | Runs the Vitest tests in watch mode.                          |
| `bun run typecheck`      | Runs `tsc` and `svelte-check`.                                |

## Tests and checks

CI runs these checks. Run them before you open a pull request.

```sh
# Go and UI
bun run lint
bun run format:check
bun run typecheck
bun run test
bun run e2e

# Window programs
(cd shell/logx && go test ./...)
(cd shell/gtk && go vet ./... && go test ./... && go build -o ../../bin/light-launcher-shell .)
(cd shell/webview2 && GOOS=windows GOARCH=amd64 go vet ./...)
```

> [!NOTE]
> The GTK build needs the GTK4 and WebKitGTK 6.0 dev packages. The WebView2 check only compiles for Windows. It cannot run on Linux.

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

Release versions follow these commit types. See [Releasing](docs/releasing.md#version-bumps).

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
