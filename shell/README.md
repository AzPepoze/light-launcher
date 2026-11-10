# Shells

[Back to README](../README.md)

A shell is the native window that shows the launcher UI. The Go core serves the UI on `127.0.0.1` and starts the shell with the page address. The shell draws the page and handles native parts, such as file dialogs.

See [Architecture](../docs/architecture.md) for the full picture.

## The two shells

| Shell | Binary | Platform | Web engine | Status |
| :--- | :--- | :--- | :--- | :--- |
| Qt | `light-launcher-shell` | Linux | Qt6 WebEngine (Chromium) | Builds. Window opens on Hyprland. |
| WebView2 | `light-launcher-shell.exe` | Windows | Microsoft Edge WebView2 | Cross-compiled only. Not run on Windows yet. |

## Command line

```text
light-launcher-shell [--transparent] <url>
```

| Part | Meaning |
| :--- | :--- |
| `<url>` | A `http://` or `https://` address with a host. Required. |
| `--transparent` | Transparent page background. Qt only. WebView2 logs a warning and ignores it. |

Exit codes:

| Code | Meaning |
| :--- | :--- |
| `0` | The window closed normally. |
| `1` | The shell could not start. On Windows this includes a missing WebView2 runtime. |
| `2` | Bad arguments. |

## Native functions in the page

The shell adds `window.llShell` to every page it loads.

| Function | Returns |
| :--- | :--- |
| `llShell.pickFolder()` | A promise with the folder path, or `null` if the user cancels. |
| `llShell.pickFile()` | A promise with the file path, or `null` if the user cancels. |

## Window behavior

| Item | Qt (Linux) | WebView2 (Windows) |
| :--- | :--- | :--- |
| Title | LightLauncher | LightLauncher |
| Minimum size | 800 x 600 | 800 x 600 |
| Default size | 1100 x 720 | 1100 x 720 |
| Remembers position and size | Yes. Saved in the user config folder. | No. |
| Links to other sites | Open in the default browser. | Open in the default browser. |
| Wayland or X11 | Qt picks it automatically. | Not applicable. |

## Limits

| Item | Detail |
| :--- | :--- |
| Qt bridge | Pages talk to the shell through console messages. Only the launcher's own address can use it. |
| Qt page origin | The launcher origin is the scheme and host of the address you pass in. Other origins get no answers. |
| WebView2 | Not run on a real Windows machine yet. See the build page. |

## More

| Topic | Page |
| :--- | :--- |
| Build both shells | [Building](../docs/building.md) |
| Windows behavior | [Windows](../docs/windows.md) |
| Linux setup | [Linux](../docs/linux.md) |
