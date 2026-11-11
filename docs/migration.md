[Back to README](../README.md)

# Migration

This page explains how to move your data from an older LightLauncher, or from interesting-launcher.

> [!IMPORTANT]
> This version does not move old data on its own. Follow the steps below. Back up your data first.

## From older builds

Older builds, from before the current folder layout, kept their data in `~/LightLauncher`. This version keeps it in the base folder. See [File locations](usage.md#file-locations).

### What goes where

| Old path (inside `~/LightLauncher`) | New path (inside the base folder) |
| :---------------------------------- | :-------------------------------- |
| `settings.json`                     | `settings.json`                   |
| `config/executables/`               | `config/executables/`             |
| `games/`                            | `config/executables/`             |
| `prefixes/`                         | `prefixes/`                       |
| `logs/`                             | `logs/`                           |

On Linux, the base folder is `~/.config/light-launcher`.

### Steps

1. Close LightLauncher and every running game.
2. Back up the old folder.

   ```sh
   cp -a ~/LightLauncher ~/LightLauncher-backup
   ```

3. Create the new folders, then copy. `cp -n` never overwrites a file that already exists.

   ```sh
   BASE=~/.config/light-launcher
   mkdir -p "$BASE/config/executables" "$BASE/prefixes" "$BASE/logs"
   cp -n ~/LightLauncher/settings.json "$BASE/"
   cp -rn ~/LightLauncher/config/executables/. "$BASE/config/executables/"
   cp -rn ~/LightLauncher/games/. "$BASE/config/executables/"
   cp -rn ~/LightLauncher/prefixes/. "$BASE/prefixes/"
   cp -rn ~/LightLauncher/logs/. "$BASE/logs/"
   ```

4. Start LightLauncher and check each game.

> [!WARNING]
> Old game settings can still point at `~/LightLauncher/prefixes`. The launcher does not rewrite these paths. Check the prefix path of each game. Prefix symlinks are fixed when a game starts.

### Old Electron flags

These no longer do anything:

- `--wayland`, `--x11`, `--ozone-platform`. They are accepted and ignored.
- `~/.config/light-launcher-flags.conf`. The launcher does not read it.

## From interesting-launcher

interesting-launcher keeps its library in a database file named `library.db`. It is in the `data/` folder next to the program. See its install guide.

### What is imported

| Imported                                                | Not imported                                                   |
| :------------------------------------------------------ | :------------------------------------------------------------- |
| Games: title, source, version, tags, notes, description | Cover image files. The web address of each cover is kept.      |
| Install folder, executable and arguments                | Saved secrets and sign-in data. Sign in again with the plugin. |
| Playtime and favorites                                  | `config.json` and profiles                                     |
| Download choices, only for games that have none yet     | Downloads and logs                                             |

### How games are matched

Games are matched by source and source id.

| Case                | What happens                                                                                                                    |
| :------------------ | :------------------------------------------------------------------------------------------------------------------------------ |
| Game is new         | It is added.                                                                                                                    |
| Game already exists | Its own data is kept. Install data and notes are filled only where they are empty. Playtime and favorite take the higher value. |

The old database is only read. It is never changed.

### Steps

1. Close interesting-launcher, so `library.db` is not being written.
2. Back up `library.db`.
3. Run the import with the path to a copy of `library.db`. The import reports how many games were added, how many were updated, and how many download choices were copied.

> [!NOTE]
> The import is the API method `LibraryImportLegacy`. No button in the window calls it yet. You need a developer tool that can call the local API.

## More

| Topic          | Page                             |
| :------------- | :------------------------------- |
| File locations | [Usage](usage.md#file-locations) |
| Known gaps     | [Usage](usage.md#known-gaps)     |
| Install        | [Install](install.md)            |
