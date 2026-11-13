[Back to README](../README.md)

# Known gaps and to-do list

This page lists what is missing, unproven or planned. Tick a box when it is done.

> [!NOTE]
> Status: `[ ]` open, `[x]` done.

## Windows

| Status | Item |
|---|---|
| [ ] | Run the WebView2 window on a real Windows machine. It has only been built, not run. |
| [ ] | Run the WebView2 `--assist` mode on real Windows or Wine. It is cross-compiled and its pure helpers are tested, but the COM capture path is unproven. |
| [ ] | Run the NSIS installer on Windows (`makensis` is not installed on the build machine). |
| [ ] | Code signing for the Windows binaries and installer. |
| [x] | Windows core and unit tests pass under Wine. |
| [x] | Download folders, install folders and file names follow Windows rules, so data made on Linux opens on Windows. |
| [ ] | Check old Windows 10 without WebView2 (the installer needs the bootstrapper). |

## Release and repos

| Status | Item |
|---|---|
| [ ] | Publish the plugin SDK. The plugins use a local `file:` link, so a fresh clone cannot install. |
| [ ] | Run `actionlint` and `shellcheck` on the workflows and scripts (not installed yet). |
| [x] | Make the plugin SDK and plugin repos public and push them. |
| [ ] | Plugin registry index (`plugins.json`) for community installs. |
| [ ] | Sign plugin packages and publish the public key. |

## Home page

| Status | Item |
|---|---|
| [ ] | Check the Home status line (library load and folder scan) and the "No games yet" empty state by eye in the real window. Only typecheck, tests and build so far. |
| [x] | Skeleton tiles on the first library load. Built and tested, not seen in a window. |
| [x] | The scan status shows the folder being scanned ("Scanning Library... 3 of 5"). Built and tested, not seen in a window. |
| [ ] | Game cards without an icon show the rocket fallback until their icon loads. |

## Plugin system

| Status | Item |
|---|---|
| [x] | Consent screen on install and update. |
| [x] | Drag and drop install (zip or folder). |
| [ ] | Install several plugins at once from one drop or picker. Each gets its own consent. Only typecheck, tests and build so far: check the "Plugin 2 of 3" dialog and the grouped site list by eye in the real window. |
| [ ] | Plugin screens: the consent ledger, queue bar, list rows and motion. Only typecheck, tests and build so far: check them by eye in the real window. |
| [ ] | A folder that holds several plugin folders is skipped with a hint, not expanded. Drop each plugin folder on its own. |
| [x] | Plugin updates from GitHub Releases: `repository` field, manual "Check for updates", update screen with consent, release `SHA256SUMS` check, and rollback when a new version does not start. |
| [ ] | Daily plugin release check (the launcher checks once soon after start and on "Check now"; the scheduler does not run it yet). |
| [x] | Release workflow in each plugin repo (not yet run on GitHub). |
| [x] | The scheduler rebuilds its jobs when a plugin is installed, updated, enabled or its settings change; no restart. |
| [ ] | Updating a dropped-in plugin without a `repository` field (drop the new version again). A dropped-in plugin that names a repository does update from GitHub. |
| [x] | Settings form built from each plugin's `settings.schema.json`. |
| [x] | Nested objects in a settings schema are not drawn (only top-level fields). |
| [x] | SDK docs for the `settings.changed` notification and the `x-ll-*` schema extensions. |
| [ ] | Check the settings form by eye in the real window (only typecheck, tests and build so far). |
| [x] | Host options for a download folder, a subfolder per game, install after download and delete archive, including the optional `source.prepareDownload` result hints. |
| [ ] | The game file download button is wired to `source.prepareDownload` but only tested with fakes and the fixture plugin. Click it in the real window with a plugin that opens the browser step. |
| [ ] | A file saved before this version has no plugin file id, so its `source.prepareDownload` call gets an empty `id`. Adding the game again from its link saves the ids. There is no button that only refreshes files. |
| [ ] | Closing the browser window during a file download is covered by a unit test only. Check the error text in the real window. |
| [ ] | The Download button's busy text ("Finish the step in the browser window") is checked by typecheck and build only. Look at it in the real window. |
| [x] | The launcher never calls `source.sync` and never schedules update checks. Plugin sync and interval settings have no effect until a scheduler exists. |
| [x] | A running plugin gets a `settings.changed` notice when settings are saved (no restart; plugins must read settings per call). |
| [x] | Embedded browser assist on Linux (GTK): ready, download (filename, size, Cookie, Referer, User-Agent), closed and persistent profile cookies verified against a local test page in a real window. |
| [ ] | Check the GTK shell's external-link opening and transparent window by eye. The link rules and flags are unit-tested only. |
| [ ] | F95 now uses the host `browser.assist` service. A real F95 download through the embedded window is untested. |
| [ ] | Embedded browser assist on Windows (WebView2) is untested on Windows. |
| [x] | If the request that started a browser step ends first, the host cancels `browser.assist` and closes the window. |
| [ ] | Plugin secrets saved before the cleanup change are not removed when the plugin is removed. |
| [x] | Stop plugins that stay idle for 5 minutes, with `pluginIdleTimeoutMinutes` to change or disable it. |
| [ ] | Check the "Sleeping" plugin badge by eye in the real window (only typecheck, tests and build so far). |
| [ ] | OS-level sandbox for plugins (permissions only cover host services today). |
| [ ] | Resource limits for plugins (CPU, memory). |
| [x] | The idle-stop tests timed out under load: the watcher armed its timer for "wait N seconds" a moment after it computed N, so a clock step in between moved the deadline. It now arms the timer for an absolute deadline (240 stress runs, no failures). |

## F95 plugin

| Status | Item |
|---|---|
| [ ] | MEGA downloads need the AES file key, but `source.prepareDownload` has no field for it, so MEGA links fail with a clear error. Needs an SDK change. |
| [ ] | Ask F95 Zone staff for written permission. Until then the README says it is not approved. |
| [x] | Contact line in the README (an issue on the now public repo). |
| [ ] | Cannot set: patch-over-full, skip extras, include pre-releases (the parser does not keep that information). |
| [x] | Delete the F95 plugin's own CDP code once the host browser assist service is used. |

## itch.io plugin

| Status | Item |
|---|---|
| [ ] | Favourites-only import and bundle-or-key filter (the itch API does not expose them). |
| [ ] | Upload type preference (only `default` is documented). |
| [x] | Default request delay is 1000 ms (it was 0). |

## Library and UI

| Status | Item |
|---|---|
| [x] | Covers are cached and shown through the cover route. |
| [x] | Library page groups local games and plugin games. |
| [x] | Cached screenshots are served through the media route and shown in the detail panel. |
| [ ] | Covers behind a login (F95) may fail because media fetches send no cookies. |
| [x] | Remove the old `/api/cover/{id}` route (the media route replaced it). |
| [x] | `bun run build:launcher` embeds the UI, so the launcher starts without `--ui`. |
| [ ] | The NativeWayland setting is saved but has no effect; GTK picks the display system. |
| [x] | The GTK shell remembers the window position and size between runs. |
| [ ] | The GTK shell can only restore the window position on X11; Wayland does not let a client place a toplevel. |
| [x] | Cover and screenshot images are evicted from memory least-recently-used first (200 entries). |
| [x] | Source filter shows labels from the library instead of raw ids. |
| [x] | The game detail can link an existing install folder. |
| [x] | The Updates page shows the update-check history with paging. |
| [ ] | Check the new Library page and drop zone by eye in the real window. |
| [ ] | Middle-click autoscroll is checked by typecheck, tests and build only; try it in the real window. |
| [x] | Home card per game folder: the backend is done, but the "Choose launcher" menu on the card is not in the UI yet. |
| [x] | Long game names scroll on hover on Home cards and Library tiles. The marquee code is copied in two places (see the shared marquee row below). |
| [ ] | Helper programs are found by name only. A helper with another name still shows as a launcher. |
| [ ] | A forced scan of 2,000 executables takes about 80 ms. Most of it is reading the config files, not grouping. |
| [ ] | No manual setting for how deep the game folders are. The automatic rule can misjudge a game folder that has three or more sub-apps and no exe of its own. |
| [x] | The Library UI reads sources, tags and stats from `LibraryOverview` and pages `LibraryList`. |
| [ ] | `LibrarySyncSource` rejects sources that only have the sync capability (no `listGames`). |
| [x] | The Library sidebar for plugin sources: the overview carries each source's `sidebar`, and the page browses it through `LibraryBrowse` and `LibraryFilters`. Fixed groups, dynamic choices and a sort with direction are drawn. |
| [ ] | Check the plugin Library sidebar by eye in the real window with the fixture plugin (only typecheck, tests and build so far). |
| [x] | The plugin SDK docs describe `sidebar`, the `browse:filters` capability, `source.browse` and `source.listFilters`. |
| [ ] | The itch plugin does not implement `source.browse` or `source.listFilters` yet. F95 does. |

## Library, Updates and Downloads UI

| Status | Item |
|---|---|
| [x] | Shared UI kit: Icon, Button, IconButton, Notice, EmptyState, LoadingState, Badge, Section, SettingRow, ProgressBar, TransferStages, TransferMeta, Sparkline and ChoiceRow. |
| [x] | Library with a source rail, a filter sidebar and tag search. |
| [x] | Plugin-provided card cover ratio, card actions and sidebar. |
| [x] | Card size and blur settings. |
| [x] | Downloads and Updates redesign with progress, speed and time left. |
| [x] | Update-check settings, global and per source, with a due time per game and a 15 minute scheduler. |
| [x] | Notification center with history. |
| [x] | Plugin errors shown as notifications. |
| [x] | `AuthRequired` error code: a missing API key gives one notice instead of hundreds of errors. |
| [x] | The API token is kept in `sessionStorage`, so it survives a reload. |
| [ ] | Nothing has been seen in a running window for the latest UI work: Library tabs, right side game panel, grouped Files, choice rows, notification center and Updates settings. |
| [ ] | Updates and Downloads moved into Library as top tabs, and the game detail as a right sidebar (in progress). |
| [x] | Home loading and scan status: status line, skeleton tiles, folder name in the scan line, empty state instead of endless rocket tiles. Built and tested, not seen in a window. |
| [x] | Game detail is a right sidebar panel (about 420 px, overlay with backdrop on narrow windows). Built and tested, not seen in a window. |
| [x] | Game files are grouped by platform with icons. Built and tested, not seen in a window. |
| [x] | Updates (with one sub-tab per source) and Downloads are Library tabs and gone from the nav. Built and tested, not seen in a window. |
| [x] | Update history rows carry `source`; the source tab filters and pages on the server, so its pages stay full. |
| [x] | Downloads sampling is owned by the long-lived library state; it continues while the Downloads tab is closed and stops when no transfer is active. |
| [ ] | Confirm the real plugin `kind` values for files; the Full and Patch badges only handle "full" and "patch". |
| [x] | The Library overview carries `kind`, `capabilities`, `sync` and `updateCheck`; the Updates page reads them from the overview instead of `PluginsList`. |
| [ ] | The Updates and Downloads tabs unmount when you switch tabs; a running check continues and the page restores progress from saved status. |
| [x] | Library sidebar and rail have no background or divider; choices are rows with a tinted selected state, no checkbox. |
| [x] | Library cards: no Installed badge, name and version in a hover overlay on the cover, denser 12 px grid. |
| [x] | Updates page shows "Last checked X ago · Next check in Y" from `nextRunAt`. |
| [x] | Notification center as a full page, with its bell at the bottom of the left rail. Built and tested, not seen in a window. |
| [x] | Rebuilt `bin/light-launcher` and `bin/light-launcher-instance` with the embedded UI. |
| [x] | Folder-scanned and hand-added games are imported into the library SQLite database as the local source; F95 and itch rows live in the same list and the overview lists all sources. Home and Library read the same rows. Migrated a copy of the real 258-local/106-F95/39-itch database. Not seen in a window. |
| [x] | F95 1.0.1 and itch 0.1.0 rebuilt against the current SDK, packed, installed, and `source.listFilters` answered through the host with real engine options. |
| [x] | Windows plugin binaries rebuilt and both `.llplugin` packages repacked; each Windows `.exe` now contains `source.listFilters`. The installed Linux copies are unchanged; a Windows install uses the new packages. |
| [ ] | Local games are now library rows with launch settings, but launching one from the Library and its prefix/Proton/playtime behavior has only been unit-tested. Check in the real window. |
| [ ] | Plugin commands now appear in the command palette; no test covers the palette entry and no window check has been done. |
| [ ] | Stale itch "API key is not configured" rows in the update check history need a one-time cleanup. |
| [x] | Per-source "Check every" now sets a minimum gap before that source's games are due again. |
| [x] | The overview sends the `sync` flag; `canSyncGames` reads it and local games never show Sync now. |
| [x] | The first enabled plugin that claims a source id owns it; later plugins with a colliding id are marked broken with a clear error. |
| [x] | `pages`, `hooks`, `launchAdapters`, `importers` and `themes` are rejected in apiVersion 1 with a clear error; the SDK schema and docs no longer advertise them. |
| [x] | F95's initialize-time `settings.get` no longer times out: the supervisor held its state mutex across the handshake, so the host service blocked until the plugin gave up. Starts now use a separate start lock, and a regression test covers host calls during initialize. |
| [x] | The shared Dropdown has an optional search box for lists over 8 choices; select-type Library sidebar groups pass the plugin's `searchable` flag. |
| [ ] | The plugin games folder path is not exposed to the frontend, so the scan-folder warning ("a plugin already scans this folder") is not wired. |
| [x] | Plugin update checks carry `not_found`, `rate_limited` or `untrusted_host`; the UI no longer matches error text. |
| [x] | IPC errors carry a stable `code`; `method_not_found`, `auth_required` and the plugin RPC codes reach `InvokeError`. |
| [x] | No `.btn` buttons remain in the plugin screens. PluginRow's full-row button and the sites disclosure stay raw on purpose: Button cannot express their layout/aria structure without neutralizing the kit. |
| [x] | The text marquee is one shared component used by GameTile and GridCard. |
| [ ] | Add a lint test that stops new `class="btn` and `material-icons` spans in features. |
| [ ] | Migrate the remaining `material-icons` spans and `btn` classes. |
| [ ] | The Updates settings source list calls PluginsList on open and shows no source icons. |
| [ ] | Checkbox is kept for PluginConsent only. |

## Migration

| Status | Item |
|---|---|
| Won't do | Moving old Electron-era data on first start. The only user chose a clean start, so v2 may break old data. |
| Won't do | Reading the old Electron flags file and rewriting old config paths. |
| Won't do | Import of old interesting-launcher data. v2 starts clean. |
| [ ] | Keep schemas of new data versioned from now on, so later releases can migrate. |

## Code quality

| Status | Item |
|---|---|
| [x] | Duplicate config and activity code in `appsvc` and `runtimesvc`. |
| [x] | `llp validate` checks that the settings file exists. |
| [x] | The itch `parseRef` is as strict as the F95 parser. |
| [x] | Update flow downloads a package twice (once to preview, once to install). |
| [x] | Failed plugin installs leave their upload folder until the next start. |

## Product choices still open

| Status | Item |
|---|---|
| [ ] | Update the launcher itself from its GitHub Releases (separate install step on each OS). Not now. |
| [ ] | Optional auto-update for plugins when permissions do not change. Not wanted for now: updates are always manual. |
| [ ] | Deep links (`lightlauncher://`). |
| [ ] | Thai and English text, and accessibility checks. |
| [ ] | macOS support (out of scope for now). |
