# Shells

[Back to README](../README.md)

A shell is the native window that shows the launcher UI. The Go core serves the UI on `127.0.0.1` and starts the shell with the page address. The shell draws the page and handles native parts, such as file dialogs.

See [Architecture](../docs/architecture.md) for the full picture.

## The two shells

| Shell | Binary | Platform | Web engine | Status |
| :--- | :--- | :--- | :--- | :--- |
| Qt | `light-launcher-shell` | Linux | Qt6 WebEngine (Chromium) | Builds. Normal and assist windows open on Hyprland. |
| WebView2 | `light-launcher-shell.exe` | Windows | Microsoft Edge WebView2 | Cross-compiled only. Not run on Windows yet. |

## Command line

The shell has a normal mode, which shows the launcher UI, and an assist mode.

```text
light-launcher-shell [--transparent] <url>
light-launcher-shell --assist <url> --profile <dir> [--title <text>] [--message <text>]
```

| Part | Meaning |
| :--- | :--- |
| `<url>` | A `http://` or `https://` address with a host. Required. |
| `--transparent` | Transparent page background. Qt only. WebView2 logs a warning and ignores it. |
| `--assist <url>` | Assist mode. Show this page for a person and report the download it starts. |
| `--profile <dir>` | Persistent browser profile folder. Created if missing. Put it under the user config folder, never inside the repository. Cookies and logins survive between runs. |
| `--title <text>` | Window title in assist mode. Defaults to `LightLauncher`. |
| `--message <text>` | Short banner above the page in assist mode. |

Assist mode takes the page only from `--assist` and ignores `LL_URL`.

Exit codes:

| Code | Meaning |
| :--- | :--- |
| `0` | Normal end: the window closed, or assist mode captured a download. |
| `1` | The shell could not start, or assist mode could not create its profile folder. On Windows this includes a missing WebView2 runtime. |
| `2` | Bad arguments. |

## Assist mode

Both shells have an assist mode. It opens a page for a human to use, then reports the download the page starts. The launcher downloads the file itself.

```text
light-launcher-shell --assist <url> --profile <dir> [--title <text>] [--message <text>]
```

| Part | Meaning |
| :--- | :--- |
| `--assist` | Selects assist mode. Must be the first argument. |
| `<url>` | A `http://` or `https://` address with a host. Required. |
| `--profile <dir>` | Browser profile folder. Created if missing. Choose a user data folder, not one inside the repository. Cookies and logins survive between runs. |
| `--title <text>` | Window title. Defaults to `LightLauncher`. |
| `--message <text>` | Short banner above the page, such as "Finish the page; the download is captured automatically". |

> [!IMPORTANT]
> Assist mode is for a human. The window is a normal embedded browser. The user logs in, waits for timers, presses the download button and solves any challenge by hand. The shell does not click, type, fill forms, solve CAPTCHAs, spoof fingerprints or hide that it is an embedded browser. There are no stealth flags.

When the page starts a download, the shell cancels it, prints one `download` line and exits. Nothing is written to disk. `blob:` and `data:` downloads are canceled without a line because there is no address to download. New windows open in the same window, and only `http` and `https` navigation is allowed. Cookies restored from the profile folder are included in the `Cookie` header.

Assist mode writes NDJSON to `stdout`, one object per line, and flushes every line. Logs go to `stderr`.

| Event | Fields | Meaning |
| :--- | :--- | :--- |
| `ready` | | The window is shown and the page is loading. |
| `download` | `url`, `filename`, `size`, `headers` | The page started a download. The shell cancels it and exits `0`. |
| `closed` | | The user closed the window without a download. Exit `0`. |
| `error` | `message` | Fatal. Exit `1`. |

`size` is `-1` when the server did not send a length. `headers` holds only the values the shell knows: `Cookie` for the download address, `Referer` from the current page and `User-Agent` from the browser settings.

## Native functions in the page

In normal mode the shell adds `window.llShell` to every page it loads. Assist mode does not add it.

| Function | Returns |
| :--- | :--- |
| `llShell.pickFolder()` | A promise with the folder path, or `null` if the user cancels. |
| `llShell.pickFile()` | A promise with the file path, or `null` if the user cancels. |

## Window behavior

| Item | Qt (Linux) | WebView2 (Windows) |
| :--- | :--- | :--- |
| Title | LightLauncher | LightLauncher. Assist mode uses `--title`. |
| Minimum size | 800 x 600 | 800 x 600. Assist mode has none. |
| Default size | 1100 x 720 | 1100 x 720. Assist mode is 1000 x 720. |
| Remembers position and size | Yes. Saved in the user config folder. | No. |
| Links to other sites | Open in the default browser. | Open in the default browser. Assist mode keeps them in the window. |
| Wayland or X11 | Qt picks it automatically. | Not applicable. |

## Limits

| Item | Detail |
| :--- | :--- |
| Qt bridge | Pages talk to the shell through console messages. Only the launcher's own address can use it. |
| Qt page origin | The launcher origin is the scheme and host of the address you pass in. Other origins get no answers. |
| WebView2 | Not run on a real Windows machine yet. See the build page. |
| WebView2 assist output | Cross-compiled and unit-tested only. Not run on real Windows or Wine yet. |
| WebView2 download capture | The `go-webview2` package has no download event or cookie manager, so the shell adds a small COM wrapper for `ICoreWebView2_4`, the cookie manager and the user agent setting. |
| Assist banner | A small non-interactive overlay in the page document. It does not take clicks. |

## More

| Topic | Page |
| :--- | :--- |
| Build both shells | [Building](../docs/building.md) |
| Windows behavior | [Windows](../docs/windows.md) |
| Linux setup | [Linux](../docs/linux.md) |
