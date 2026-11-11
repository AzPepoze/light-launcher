[Back to README](../README.md)

# Migration

This page explains how to move your data from an older LightLauncher.

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

These are not supported and are no longer accepted:

- `--wayland`, `--x11`, `--ozone-platform`. Use `QT_QPA_PLATFORM=xcb light-launcher` to force X11.
- `~/.config/light-launcher-flags.conf`. The launcher does not read it.

## More

| Topic          | Page                             |
| :------------- | :------------------------------- |
| File locations | [Usage](usage.md#file-locations) |
| Known gaps     | [Usage](usage.md#known-gaps)     |
| Install        | [Install](install.md)            |
