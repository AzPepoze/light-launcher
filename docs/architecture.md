[Back to README](../README.md)

# Architecture

LightLauncher has a core program written in Go and a small native window, called the shell.

## Parts

| Part                                 | Written in           | What it does                                                                                                                                 |
| :----------------------------------- | :------------------- | :------------------------------------------------------------------------------------------------------------------------------------------- |
| Launcher (`light-launcher`)          | Go                   | Serves the UI on `127.0.0.1`. Keeps the game library and settings. Starts the shell.                                                         |
| Instance (`light-launcher-instance`) | Go                   | Runs one game session and watches it until it exits.                                                                                         |
| UI                                   | Svelte               | The screens you see. Runs inside the shell's web page.                                                                                       |
| Shell (`light-launcher-shell`)       | Go with a web engine | Opens the window and shows the UI. Also opens file dialogs.                                                                                  |
| Plugin                               | Any language         | A separate program that adds a source. The launcher starts it and talks to it over text messages. See [Writing plugins](writing-plugins.md). |

## How the parts connect

```text
 light-launcher (Go core)
   - serves the UI on 127.0.0.1
   - game library, settings, instances
          |
          | starts the shell with the page address
          v
 light-launcher-shell (native window)
   - Qt6 WebEngine on Linux
   - WebView2 on Windows
   - file dialogs for window.llShell
          |
          | links to other sites
          v
 default browser
```

The shell does not hold any game logic. It only shows the page and runs the native dialogs.

## Why it works this way

| Choice              | Reason                                                                      |
| :------------------ | :-------------------------------------------------------------------------- |
| UI in a web page    | One UI for Linux and Windows.                                               |
| Go core             | One language for the library, settings and game sessions.                   |
| Small shell         | Only the window differs per system.                                         |
| Qt on Linux         | Gives a Chromium web engine without the size of Electron.                   |
| WebView2 on Windows | Windows 11 includes the runtime. On Windows 10 it may need to be installed. |

## Native functions

The page calls `window.llShell` for things a web page cannot do alone.

| Function               | Result                                            |
| :--------------------- | :------------------------------------------------ |
| `llShell.pickFolder()` | The chosen folder path, or `null` when cancelled. |
| `llShell.pickFile()`   | The chosen file path, or `null` when cancelled.   |

For command line flags and window behavior, see the [shell page](../shell/README.md).

## More

| Topic                    | Page                      |
| :----------------------- | :------------------------ |
| Install                  | [Install](install.md)     |
| Flags and file locations | [Usage](usage.md)         |
| Plugins                  | [Plugins](plugins.md)     |
| Security                 | [Security](security.md)   |
| Build the files          | [Building](building.md)   |
| Windows                  | [Windows](windows.md)     |
| Linux                    | [Linux](linux.md)         |
| Releases                 | [Releasing](releasing.md) |
