[Back to README](../README.md)

# Known gaps and to-do list

This page lists what is missing, unproven or planned. Tick a box when it is done.

> [!NOTE]
> Status: `[ ]` open, `[x]` done.

## Windows

| Status | Item |
|---|---|
| [ ] | Run the WebView2 window on a real Windows machine. It has only been built, not run. |
| [ ] | Run the NSIS installer on Windows (`makensis` is not installed on the build machine). |
| [ ] | Code signing for the Windows binaries and installer. |
| [x] | Windows core and unit tests pass under Wine. |
| [ ] | Check old Windows 10 without WebView2 (the installer needs the bootstrapper). |

## Release and repos

| Status | Item |
|---|---|
| [ ] | First release needs `Release-As: 2.0.0` (the manifest still says 1.0.0). |
| [ ] | Publish the plugin SDK. The plugins use a local `file:` link, so a fresh clone cannot install. |
| [ ] | Run `actionlint` and `shellcheck` on the workflows and scripts (not installed yet). |
| [ ] | Make the plugin SDK and plugin repos public and push them. |
| [ ] | Plugin registry index (`plugins.json`) for community installs. |
| [ ] | Sign plugin packages and publish the public key. |

## Plugin system

| Status | Item |
|---|---|
| [x] | Consent screen on install and update. |
| [x] | Drag and drop install (zip or folder). |
| [ ] | Updating a plugin that was dropped in (it has no web source, so drop the new version again). |
| [ ] | Settings form built from each plugin's `settings.schema.json` (in progress). |
| [ ] | Host option for a download folder, a subfolder per game, install after download and delete archive (in progress). |
| [ ] | The launcher never calls `source.sync` and never schedules update checks. Plugin sync and interval settings have no effect until a scheduler exists. |
| [ ] | Restart a plugin when its settings are saved (today changes apply on the next plugin start). |
| [ ] | Browser assist (Chrome or Edge over CDP) is not implemented in the host. |
| [ ] | Plugin secrets saved before the cleanup change are not removed when the plugin is removed. |
| [ ] | OS-level sandbox for plugins (permissions only cover host services today). |
| [ ] | Resource limits for plugins (CPU, memory). |

## F95 plugin

| Status | Item |
|---|---|
| [ ] | The file-host resolvers (Pixeldrain, Mega, MediaFire and others) are not called by `prepareDownload`, so only direct links work. Decide whether to connect them. |
| [ ] | Ask F95 Zone staff for written permission. Until then the README says it is not approved. |
| [ ] | Add a contact line to the README that works while the repo is private (an issue only works on a public repo). |
| [ ] | Cannot set: patch-over-full, skip extras, include pre-releases (the parser does not keep that information). |
| [ ] | Real browser-assist run against a real Chrome (the debug-port change is untested). |

## itch.io plugin

| Status | Item |
|---|---|
| [ ] | Favourites-only import and bundle-or-key filter (the itch API does not expose them). |
| [ ] | Upload type preference (only `default` is documented). |
| [ ] | Decide a polite default request delay (it is 0 today). |

## Library and UI

| Status | Item |
|---|---|
| [x] | Covers are cached and shown through the cover route. |
| [x] | Library page groups local games and plugin games. |
| [ ] | Cached screenshots are downloaded but not served; the detail panel uses the remote link. |
| [ ] | Covers behind a login (F95) may fail because media fetches send no cookies. |
| [ ] | Cover images are kept in memory for the whole session (no eviction for very large libraries). |
| [ ] | Source filter shows raw ids (`f95`) instead of labels. |
| [ ] | Check which view "quick launch" means (grid is 1:1, the side view is not). |
| [ ] | No UI for "link existing folder" and the update-check history. |
| [ ] | Import of old interesting-launcher data has no button. |
| [ ] | Check the new Library page and drop zone by eye in the real window. |

## Migration

| Status | Item |
|---|---|
| [ ] | Old data is not moved on first start. |
| [ ] | The old Electron flags file is not read. |
| [ ] | Old config paths are not rewritten. |
| [ ] | Run the new app against real old data (not tried yet). |

## Code quality

| Status | Item |
|---|---|
| [ ] | Duplicate config and activity code in `appsvc` and `runtimesvc`. |
| [ ] | `llp validate` does not check that the settings file exists (pack fails instead). |
| [ ] | The `parseRef` parameter check differs between itch (no validation) and f95 (strict). |
| [ ] | Update flow downloads a package twice (once to preview, once to install). |
| [ ] | Failed plugin installs leave their upload folder until the next start. |

## Product choices still open

| Status | Item |
|---|---|
| [ ] | Auto-update for the launcher and for plugins, with rollback. |
| [ ] | Deep links (`lightlauncher://`). |
| [ ] | Thai and English text, and accessibility checks. |
| [ ] | macOS support (out of scope for now). |
