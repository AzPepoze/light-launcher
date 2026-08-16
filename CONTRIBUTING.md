# Contributing to LightLauncher

Thank you for contributing to **LightLauncher**! Bug reports, feature requests, documentation improvements, and pull requests are welcome.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Prerequisites](#prerequisites)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Available Scripts](#available-scripts)
- [Testing & Quality Checks](#testing--quality-checks)
- [Coding Guidelines](#coding-guidelines)
- [Submitting a Pull Request](#submitting-a-pull-request)

---

## Architecture Overview

LightLauncher utilizes a dual-engine architecture designed for maximum performance and isolation on Linux:

1. **Frontend & Controller (Electron + Svelte 5 + TypeScript)**:
   - Svelte 5 UI with modern styling and responsive layouts.
   - Electron main process coordinates system state, wineprefix indexing, configuration persistence, and system tray lifecycle.
2. **Instance Daemon (`core/cmd/instance` in Go)**:
   - Super-lightweight Go binary that supervises individual game process lifecycle.
   - Executes games via `umu-run` with isolated per-game tray, logs, Gamescope, and LSFG integration.

---

## Prerequisites

Ensure you have the following installed locally:

- **[Bun](https://bun.sh/)** (or Node.js >= 18)
- **[Go](https://go.dev/)** (>= 1.22)
- **[umu-launcher](https://github.com/Open-Wine-Components/umu-launcher)** (for runtime execution)
- **Steam** and compatibility tools (e.g. Proton, GE-Proton)

---

## Development Setup

```bash
# 1. Clone the repository
git clone https://github.com/AzPepoze/light-launcher.git
cd light-launcher

# 2. Install Node dependencies
bun install

# 3. Start development server (Live reload for Svelte + Electron + Go instance compilation)
bun run dev
```

---

## Project Structure

```text
light-launcher/
├── core/                       # Go instance daemon source
│   ├── cmd/instance/           # Main Go executable entrypoint
│   └── internal/               # Adapters, builder, logger, config types
├── src/
│   ├── main/                   # Electron main process (Services & IPC)
│   ├── preload/                # Electron preload bridge
│   ├── renderer/               # Svelte 5 frontend (Pages, Components, Stores)
│   └── shared/                 # Shared types, constants, logger utilities
├── bin/                        # Compiled binary outputs
├── dist/                       # Packaged web & electron outputs
└── package.json                # Scripts & project dependencies
```

---

## Available Scripts

| Command                | Description                                                               |
| :--------------------- | :------------------------------------------------------------------------ |
| `bun run dev`          | Runs Go instance build and concurrently starts Vite dev server & Electron |
| `bun run dev:go`       | Compiles the Go instance binary to `./bin/light-launcher-instance`        |
| `bun run dev:renderer` | Starts the Vite dev server for the Svelte frontend                        |
| `bun run dev:electron` | Compiles TypeScript and starts Electron                                   |
| `bun run build`        | Builds Go binary, production Vite renderer, and TypeScript main process   |
| `bun run test`         | Runs the Vitest test suite                                                |
| `bun run typecheck`    | Runs TypeScript (`tsc`) and Svelte (`svelte-check`) type checking         |
| `bun run lint`         | Runs `oxlint` for fast code linting                                       |
| `bun run format`       | Formats the codebase using `oxfmt`                                        |

---

## Testing & Quality Checks

Before submitting your pull request, ensure all checks pass:

```bash
# 1. Run unit tests
bun run test

# 2. Verify TypeScript & Svelte types
bun run typecheck

# 3. Check code formatting & lint
bun run lint
bun run format:check
```

---

## Coding Guidelines

- **Clean Architecture & Separation of Concerns**: Keep controllers thin; delegate domain logic to specialized services.
- **Type Safety**: Avoid `any` types wherever possible; maintain synchronization between Go structs and TypeScript interfaces in `src/shared/types/`.
- **Logging**: Use `LoggerService` rather than raw `console.log`/`console.error` in the main process.
- **Formatting**: Run `bun run format` (`oxfmt`) before committing.

---

## Submitting a Pull Request

1. Fork the repository and create a feature branch (`git checkout -b feature/my-feature`).
2. Make your changes following the coding guidelines.
3. Commit with clear, descriptive commit messages.
4. Push to your branch and open a Pull Request explaining the changes made and how they were tested.
