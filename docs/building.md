[Back to README](../README.md)

# Building

This page explains how to build the Linux and Windows files from source.

## Tools

| Tool                             | Needed for        | Notes                                                                     |
| :------------------------------- | :---------------- | :------------------------------------------------------------------------ |
| Go                               | Everything        | Version from `go.mod`.                                                    |
| Bun                              | The UI            | Installs the UI packages and builds the UI.                               |
| Git                              | Version number    | `git describe` reads the tags.                                            |
| GTK4 and WebKitGTK 6.0 dev files | Linux shell       | Arch: `gtk4 webkitgtk-6.0`. Ubuntu: `libgtk-4-dev libwebkitgtk-6.0-dev`.  |
| zip, tar                         | Packing           | Usually installed already.                                                |
| NSIS (`makensis`)                | Windows installer | Optional. Without it, the installer is skipped.                           |

## Steps

1. Install the packages with Bun.

   ```sh
   bun install
   ```

2. Build the Linux files.

   ```sh
   bash scripts/build-linux.sh
   ```

3. Build the Windows files. This works on Linux, because Go cross-compiles the program.

   ```sh
   bash scripts/build-windows.sh
   ```

> [!NOTE]
> The launcher is built from `core/cmd/launcher` with the `embedui` tag. The UI is copied there while the build runs, and the copy is removed at the end. `bun run build:launcher` uses the same script as the release builds, so the UI is embedded in `bin/light-launcher` too.

## Development build

```sh
bun run build
```

This writes `bin/light-launcher` and `bin/light-launcher-instance`. The launcher has the UI built in, so it runs without `--ui`. On Linux, `bun run build:linux` also builds the GTK window. `bun run dev` starts the Vite dev server and the launcher API instead.

> [!NOTE]
> The local build scripts use Bash. On Windows, run them from Git Bash or WSL, or use `bun run dev`.

## Output

Both scripts write to the `dist/` folder.

| File                                             | Made by            | What it is                                    |
| :----------------------------------------------- | :----------------- | :-------------------------------------------- |
| `light-launcher-<version>-linux-x64.tar.gz`      | `build-linux.sh`   | Linux files in one archive.                   |
| `light-launcher-<version>-windows-x64.zip`       | `build-windows.sh` | Windows files, no install needed.             |
| `light-launcher-<version>-windows-x64-setup.exe` | `build-windows.sh` | Windows installer.                            |
| `SHA256SUMS`                                     | Both scripts       | Checksums for the archives and the installer. |

The `<version>` comes from `git describe --tags --always`. A leading `v` is removed.

| Folder          | What is in it                     |
| :-------------- | :-------------------------------- |
| `dist/linux/`   | The Linux files before packing.   |
| `dist/windows/` | The Windows files before packing. |

## Build one part

| Part                     | Command                                                                                                                |
| :----------------------- | :--------------------------------------------------------------------------------------------------------------------- |
| Launcher with UI         | `bash scripts/build-launcher.sh bin/light-launcher`                                                                    |
| GTK shell (Linux)        | `cd shell/gtk && go build -ldflags="-s -w" -o light-launcher-shell .`                                                  |
| WebView2 shell (Windows) | `cd shell/webview2 && GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H=windowsgui" -o light-launcher-shell.exe .` |
| Instance                 | `go build -ldflags="-s -w" -o light-launcher-instance ./core/cmd/instance`                                             |

The stripped GTK shell is about 3 MB. The GTK and WebKit libraries are not in the file. They come from your system packages.

## Settings

| Name           | Default               | Effect                                         |
| :------------- | :-------------------- | :--------------------------------------------- |
| `LAUNCHER_PKG` | `./core/cmd/launcher` | The package the scripts build as the launcher. |

## More

| Topic          | Page                         |
| :------------- | :--------------------------- |
| Linux packages | [Linux](linux.md)            |
| Windows files  | [Windows](windows.md)        |
| Shell options  | [Shells](../shell/README.md) |
| Publishing     | [Releasing](releasing.md)    |
