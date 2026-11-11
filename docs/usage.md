[Back to README](../README.md)

# Usage

This page lists the commands, their flags, and where LightLauncher keeps its files.

## Commands at a glance

| Command                   | What it is for                                                              |
| :------------------------ | :-------------------------------------------------------------------------- |
| `light-launcher`          | Opens the launcher window.                                                  |
| `light-launcher-instance` | Runs one game. The launcher uses it for you.                                |
| `light-launcher plugin`   | Installs and manages plugins. See [Plugin commands](#plugin-commands).     |
| `light-launcher-shell`    | The window program. The launcher starts it. You do not need to run it.      |

## light-launcher

```text
light-launcher [flags] [GAME_PATH]
```

| Flag                                             | Default     | What it does                                                                 |
| :----------------------------------------------- | :---------- | :--------------------------------------------------------------------------- |
| `GAME_PATH`                                      | none        | Opens the launcher on this game.                                             |
| `--shell <name>`                                 | `auto`      | The window type. See [Shell values](#shell-values).                          |
| `--no-shell`                                     | off         | Runs the UI with no window. The launcher logs the address to open.           |
| `--ui <dir>`                                     | built-in UI | Serves the UI from this folder. The folder must contain `index.html`.        |
| `--edit-lsfg`                                    | off         | Opens the Lossless Scaling (LSFG) editor on start. Linux only.               |

### Shell values

| Value      | What it does on Linux                          | What it does on Windows                       |
| :--------- | :--------------------------------------------- | :-------------------------------------------- |
| `auto`     | Starts `light-launcher-shell` (Qt).            | Starts `light-launcher-shell.exe` (WebView2). |
| `qt`       | Starts the shell program next to the launcher. | Same as `auto`.                               |
| `webview2` | Starts the shell program next to the launcher. | Same as `auto`.                               |
| `browser`  | Opens the default browser.                     | Opens the default browser.                    |

> [!NOTE]
> `qt` and `webview2` start the same kind of program: the shell next to the launcher. The shell file decides the window engine.

If the shell program is missing, the launcher opens the default browser instead.

### Starting twice

If a launcher is already running, a second start does not open a new window.

| You run                            | Result                                                                  |
| :--------------------------------- | :---------------------------------------------------------------------- |
| `light-launcher`                   | The second start checks the running launcher and exits.                 |
| `light-launcher /path/to/game.exe` | The game path goes to the running launcher, and the second start exits. |

### Examples

```sh
# Open the launcher
light-launcher

# Open the launcher on one game
light-launcher ~/Games/Game/game.exe

# Use the default browser instead of a window
light-launcher --shell=browser

# Run the UI with no window
light-launcher --no-shell
```

### Exit codes

| Code | Meaning                                     |
| :--- | :------------------------------------------ |
| `0`  | Done, or handed over to a running launcher. |
| `1`  | The launcher could not start.               |
| `2`  | Wrong flags or an unknown shell name.       |

## light-launcher-instance

The launcher starts this program for each game. It keeps the game running, adds a tray icon, and can open a log window.

You can also start a game with it directly:

```sh
light-launcher-instance -game ~/Games/Game/game.exe
```

`-game` is required. Without it, the program stops with an error.

### Core flags

| Flag                      | Default                                 | What it does                                                                |
| :------------------------ | :-------------------------------------- | :-------------------------------------------------------------------------- |
| `-game <path>`            | none                                    | The game `.exe`. Required.                                                  |
| `-launcher <path>`        | the game path                           | A program to run instead of the game file, such as the game's own launcher. |
| `-prefix <path>`          | `prefixes/Default` in the config folder | The Wine prefix folder. See [File locations](#file-locations).              |
| `-proton-path <path>`     | none                                    | The Proton folder.                                                          |
| `-proton-pattern <name>`  | none                                    | A Proton name pattern for umu-run.                                          |
| `-game-name <name>`       | the file name                           | The name shown on the tray icon.                                            |
| `-discord-client-id <id>` | none                                    | Discord app ID for Rich Presence.                                           |
| `-track-playtime`         | `true`                                  | Saves recent activity and playtime.                                         |
| `-logs`                   | `true`                                  | Opens a terminal window with live logs.                                     |

### Tool flags

| Flag                       | Default | What it does                                                  |
| :------------------------- | :------ | :------------------------------------------------------------ |
| `-mangohud`, `-mango`      | off     | Turns on the MangoHud overlay.                                |
| `-gamemode`                | off     | Turns on Feral GameMode.                                      |
| `-memory-min`              | off     | Turns on memory protection. Needs `-memory-min-value`.        |
| `-memory-min-value <size>` | none    | Memory to protect, such as `4G`. Without it, no limit is set. |

### Gamescope flags

Gamescope is a compositor that scales the game. These flags pass its settings. When a flag is not set, Gamescope uses its own default.

| Flag                                      | What it sets                                   |
| :---------------------------------------- | :--------------------------------------------- |
| `-gamescope`                              | Turns Gamescope on.                            |
| `-gs-w <width>`, `-gs-h <height>`         | Internal render size.                          |
| `-gs-out-w <width>`, `-gs-out-h <height>` | Output size.                                   |
| `-gs-r <rate>`                            | Refresh rate.                                  |
| `-gs-fr-limit <fps>`                      | Frame rate cap.                                |
| `-gs-window-mode <mode>`                  | `fullscreen`, `borderless` or `windowed`.      |
| `-gs-scaler <name>`                       | `auto`, `integer`, `fit`, `fill` or `stretch`. |
| `-gs-filter <name>`                       | `linear`, `nearest`, `fsr`, `nis` or `pixel`.  |
| `-gs-sharpness <0-20>`                    | Sharpness for upscaling filters.               |
| `-gs-hdr`                                 | Turns on HDR output.                           |
| `-gs-adaptive-sync`                       | Turns on VRR (adaptive sync).                  |
| `-gs-mangoapp`                            | Shows the MangoHud overlay inside Gamescope.   |
| `-gs-custom-args <args>`                  | Extra Gamescope arguments.                     |

### Lossless Scaling flags

Lossless Scaling (LSFG) adds frame generation through lsfg-vk.

| Flag                                  | Default  | What it sets            |
| :------------------------------------ | :------- | :---------------------- |
| `-lsfg`                               | off      | Turns LSFG on.          |
| `-lsfg-multiplier <n>` (`-lsfg-mult`) | `2`      | Frame multiplier.       |
| `-lsfg-perf`                          | off      | Performance mode.       |
| `-lsfg-dll <path>` (`-lsfg-dll-path`) | none     | Path to `Lossless.dll`. |
| `-lsfg-gpu <index>`                   | none     | GPU index.              |
| `-lsfg-flow <scale>`                  | `1.0`    | Flow scale.             |
| `-lsfg-pacing <mode>`                 | `smooth` | Pacing mode.            |
| `-lsfg-fp16`                          | off      | Allows FP16 math.       |

### Examples

```sh
# Basic start with a Proton version
light-launcher-instance \
  -game ~/Games/Cyberpunk2077/bin/x64/Cyberpunk2077.exe \
  -prefix ~/.config/light-launcher/prefixes/Cyberpunk \
  -proton-path ~/.local/share/Steam/compatibilitytools.d/GE-Proton9-25

# MangoHud and GameMode
light-launcher-instance -game ~/Games/EldenRing/eldenring.exe -mangohud -gamemode

# Gamescope: 720p inside, 1080p output, FSR filter
light-launcher-instance \
  -game ~/Games/Witcher3/bin/x64/witcher3.exe \
  -gamescope -gs-w 1280 -gs-h 720 -gs-out-w 1920 -gs-out-h 1080 \
  -gs-filter fsr -gs-sharpness 5
```

## Background jobs

While the launcher is open it keeps two kinds of work running in the background.

| Job | When it runs | What it does |
| :--- | :--- | :--- |
| Plugin sync | Every 24 hours, or the plugin's `syncIntervalHours` setting | Calls `source.sync` for plugins that list `sync:games` or `sync:index`. |
| Update check | Every 30 minutes | Checks games that were not checked in the last 6 hours. |

The first run starts 30 to 120 seconds after the launcher opens, and a restart keeps the saved run times. A job that fails waits longer each time, up to 6 hours. Disabled plugins are skipped.

## Download a game file

A game from a plugin lists its files on the game page. Click **Download** to queue one.

| What you see | What it means |
| :--- | :--- |
| The file appears in Downloads | The file is downloading. |
| **Finish the step in the browser window** | The plugin needs a step that only you can do. A browser window opens. Finish the step there, such as pressing the site's download button. |
| An error message under the file list | The window was closed, the step ran out of time (the plugin sets the limit, up to 10 minutes), or another browser window is already open. Click **Download** to try again. |

> [!IMPORTANT]
> Nothing is automated. The launcher never presses buttons or fills in forms for you. Some downloads open a browser window, and you finish the step in it.

> [!NOTE]
> Only one browser window runs at a time. A second download that needs a window shows an error until the first window is closed.

A file of a game without a plugin (for example a game you added by hand) downloads from the link saved with it.

## Plugin commands

```text
light-launcher plugin list
light-launcher plugin install <path-or-url> [--allow-unsigned]
light-launcher plugin update <id> [--allow-unsigned]
light-launcher plugin remove <id> [--delete-data]
light-launcher plugin enable <id>
light-launcher plugin disable <id>
light-launcher plugin rollback <id>
```

| Command              | What it does                                                      |
| :------------------- | :---------------------------------------------------------------- |
| `list`               | Shows each plugin, its version and its status.                    |
| `install`            | Installs a plugin from a file or an `https://` link.              |
| `update`             | Installs the newest version from the place it was installed from. |
| `remove`             | Removes the plugin. `--delete-data` also deletes its saved data.  |
| `enable` / `disable` | Turns a plugin on or off.                                         |
| `rollback`           | Goes back to the previous version.                                |

`--allow-unsigned` installs a plugin that has no signature. See [Plugins](plugins.md#trust-levels).

## Shell program

```text
light-launcher-shell [--transparent] <url>
```

The shell opens the given `http://` or `https://` address. The launcher runs it for you. See [Shells](../shell/README.md) for its exit codes and window behavior.

## File locations

LightLauncher uses one folder for its files. The folder name is `light-launcher`.

| Platform | Base folder                                                                                        |
| :------- | :------------------------------------------------------------------------------------------------- |
| Linux    | `$XDG_CONFIG_HOME/light-launcher`, or `~/.config/light-launcher` when `XDG_CONFIG_HOME` is not set |
| Windows  | `%APPDATA%\light-launcher`                                                                         |

Inside the base folder:

| Path                                  | What is in it                                                           |
| :------------------------------------ | :---------------------------------------------------------------------- |
| `settings.json`                       | The launcher settings.                                                  |
| `config/executables/<id>/config.json` | The settings of one game.                                               |
| `prefixes/`                           | Wine prefixes. The default prefix is `prefixes/Default`.                |
| `prefixes/<name>/light-launcher.json` | The settings of one prefix.                                             |
| `logs/`                               | Game logs. The last 10 are kept.                                        |
| `plugins/`                            | Installed plugins, trusted keys and state.                              |
| `plugin-data/<id>/`                   | Data each plugin saves, including its settings.                         |
| `browser-profiles/<id>/`              | Cookies and logins from the browser window of one plugin.               |
| `launcher.lock`                       | The port and token of the running launcher. Only your user can read it. |

On Linux, the window size and position are saved in `~/.config/light-launcher/window-geometry`.

> [!IMPORTANT]
> `launcher.lock` holds the API token. Do not share it. See [Security](security.md).

## Known gaps

These parts are not finished in this version.

| Part               | State                                                                                             |
| :----------------- | :------------------------------------------------------------------------------------------------ |
| Old data           | Not moved on start. See [Migration](migration.md).                                                |
| Old game paths     | Game settings that point at `~/LightLauncher/prefixes` are not rewritten. Check each prefix path. |
| Old display flags  | `--wayland`, `--x11`, `--ozone-platform` and `light-launcher-flags.conf` are not supported.       |

## More

| Topic                      | Page                      |
| :------------------------- | :------------------------ |
| Install                    | [Install](install.md)     |
| Plugins                    | [Plugins](plugins.md)     |
| Security                   | [Security](security.md)   |
| Moving from older versions | [Migration](migration.md) |
| Linux                      | [Linux](linux.md)         |
| Windows                    | [Windows](windows.md)     |
