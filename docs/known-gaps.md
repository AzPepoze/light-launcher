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
| [x] | The SDK package is publishable: public metadata, `prepublishOnly`, and a tag-driven npm workflow with provenance. |
| [ ] | Publish the SDK to npm (needs the `@light-launcher` scope/token), then switch the plugin repos from the local `file:` dependency to the published version. |
| [x] | `actionlint` and `shellcheck -x` run clean on the workflows and `scripts/*.sh`; the few shellcheck findings were fixed. |
| [x] | Make the plugin SDK and plugin repos public and push them. |
| [x] | The launcher reads a versioned HTTPS `plugins.json` index and shows installable entries in the Community dialog; the URL is a setting. |
| [ ] | Host and publish a real `plugins.json` and its package URLs/digests; until then the Community dialog says the list is not configured. |
| [x] | Plugin packages can be signed (`llp keygen` / `llp sign`); the app lists trusted key fingerprints and can add/remove keys. |
| [ ] | Publish the project's public signing key and a community registry host; no release URL or key is published yet. |

## Home page

| Status | Item |
|---|---|
| [ ] | Check the Home status line (library load and folder scan) and the "No games yet" empty state by eye in the real window. Only typecheck, tests and build so far. |
| [x] | Skeleton tiles on the first library load. Built and tested, not seen in a window. |
| [x] | The scan status shows the folder being scanned ("Scanning Library... 3 of 5"). Built and tested, not seen in a window. |
| [x] | Game cards show a neutral skeleton while an icon is pending, share in-flight loads, and use the rocket only when the icon is truly missing or failed. Grid and list views match. |

## Plugin system

| Status | Item |
|---|---|
| [x] | Consent screen on install and update. |
| [x] | Drag and drop install (zip or folder). |
| [ ] | Install several plugins at once from one drop or picker. Each gets its own consent. Only typecheck, tests and build so far: check the "Plugin 2 of 3" dialog and the grouped site list by eye in the real window. |
| [ ] | Plugin screens: the consent ledger, queue bar, list rows and motion. Only typecheck, tests and build so far: check them by eye in the real window. |
| [ ] | A folder that holds several plugin folders is skipped with a hint, not expanded. Drop each plugin folder on its own. |
| [x] | Plugin updates from GitHub Releases: `repository` field, manual "Check for updates", update screen with consent, release `SHA256SUMS` check, and rollback when a new version does not start. |
| [x] | The scheduler runs a daily `plugins.update-check` job; the startup check and "Check now" remain. |
| [x] | Release workflow in each plugin repo (not yet run on GitHub). |
| [x] | The scheduler rebuilds its jobs when a plugin is installed, updated, enabled or its settings change; no restart. |
| [x] | Dropping a newer local package updates the installed plugin in place through `PluginInstallOrUpdate`; data/secrets survive, failed starts roll back, same version gives a clear error. GitHub updates stay strict. |
| [ ] | The install dialog still says "Install plugin" when the same package will update an installed plugin. Use the preview's manifest id and the installed list to switch the title/confirm copy. |
| [x] | Settings form built from each plugin's `settings.schema.json`. |
| [x] | Nested objects in a settings schema are not drawn (only top-level fields). |
| [x] | SDK docs for the `settings.changed` notification and the `x-ll-*` schema extensions. |
| [ ] | Check the settings form by eye in the real window (only typecheck, tests and build so far). |
| [x] | Host options for a download folder, a subfolder per game, install after download and delete archive, including the optional `source.prepareDownload` result hints. |
| [ ] | The game file download button is wired to `source.prepareDownload` but only tested with fakes and the fixture plugin. Click it in the real window with a plugin that opens the browser step. |
| [x] | A plugin game's detail has a "Refresh files" action that calls `LibraryRefreshFiles`, replaces old file rows/plugin ids, and updates the displayed detail. RPC/mapping tested; button visibility is typechecked only. |
| [ ] | Closing the browser window during a file download is covered by a unit test only. Check the error text in the real window. |
| [ ] | The Download button's busy text ("Finish the step in the browser window") is checked by typecheck and build only. Look at it in the real window. |
| [x] | The launcher never calls `source.sync` and never schedules update checks. Plugin sync and interval settings have no effect until a scheduler exists. |
| [x] | A running plugin gets a `settings.changed` notice when settings are saved (no restart; plugins must read settings per call). |
| [x] | Embedded browser assist on Linux (GTK): ready, download (filename, size, Cookie, Referer, User-Agent), closed and persistent profile cookies verified against a local test page in a real window. |
| [ ] | Check the GTK shell's external-link opening and transparent window by eye. The link rules and flags are unit-tested only. |
| [ ] | F95 now uses the host `browser.assist` service. A real F95 download through the embedded window is untested. |
| [ ] | Embedded browser assist on Windows (WebView2) is untested on Windows. |
| [x] | If the request that started a browser step ends first, the host cancels `browser.assist` and closes the window. |
| [x] | Removing a plugin deletes every secret for it, including keys saved before the per-key tracking existed; keyring stores fall back to recorded keys when they cannot enumerate. |
| [x] | Stop plugins that stay idle for 5 minutes, with `pluginIdleTimeoutMinutes` to change or disable it. |
| [ ] | Check the "Sleeping" plugin badge by eye in the real window (only typecheck, tests and build so far). |
| [x] | The idle-stop tests timed out under load: the watcher armed its timer for "wait N seconds" a moment after it computed N, so a clock step in between moved the deadline. It now arms the timer for an absolute deadline (240 stress runs, no failures). |

## F95 plugin

| Status | Item |
|---|---|
| [x] | MEGA downloads work: F95 1.0.2 returns an `aes-128-ctr` key/nonce and the host decrypts the finished file before install. Limitations: AES-CTR has no MAC (a wrong key on a non-install download is undetected), and the key is stored in plaintext in the downloads row. |
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
| [x] | The NativeWayland setting now sets `GDK_BACKEND=wayland` (or `x11`) for the GTK shell, so the toggle takes effect after restart. |
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
| [x] | A forced scan of 2,000 executables takes about 80 ms, which is acceptable. |
| [x] | Each scan folder has an editable depth (`-1` for no limit) and ignore patterns; the automatic grouping heuristic still applies inside those bounds. |
| [x] | The Library UI reads sources, tags and stats from `LibraryOverview` and pages `LibraryList`. |
| [x] | `LibrarySyncSource` accepts sync-only sources (`sync:games` / `sync:index`) and saves games returned by `source.sync`. |
| [x] | Scheduled sync uses `Library.SyncSource`, so sync-only sources save their games on the timer just like a manual sync. |
| [x] | The Library sidebar for plugin sources: the overview carries each source's `sidebar`, and the page browses it through `LibraryBrowse` and `LibraryFilters`. Fixed groups, dynamic choices and a sort with direction are drawn. |
| [ ] | Check the plugin Library sidebar by eye in the real window with the fixture plugin (only typecheck, tests and build so far). |
| [ ] | Plugin browse grid in a source group: picked choices show the plugin's results. Built and unit-tested; not seen in a window. |
| [x] | Browse results open the add dialog with their link. Retry asks the failed source again. A plugin group stays visible while its browse is on. Built; not seen in a window. |
| [ ] | Check the browse grid against a real plugin that sends no total count (the grid grows one page at a time). |
| [ ] | Home: scanned games use the cover from the Go payload (`cover`, `coverLocal`) before the exe icon. Built and tested; not seen in a window. |
| [ ] | Home: a folder group shows its header and skeleton before the scan ends, then fills as each folder finishes. Not seen in a window. |
| [ ] | Home: the folder count shows "..." while the folder loads. Not seen in a window. |
| [ ] | Library: the local exe-icon fallback for coverless games needs a real window check. |
| [ ] | Library: local games ignore the sort direction. `toLibraryFilters` does not send `order`, so the list always uses the sort's own order. |
| [ ] | Library source chips: clicking a source scrolls to its group. The highlight does not follow scrolling yet (no scroll spy). |
| [ ] | Library redesign (toolbar, source chips, filter panel, tiles, game drawer): check by eye in the real window. Only typecheck and tests so far. |
| [x] | The plugin SDK docs describe `sidebar`, the `browse:filters` capability, `source.browse` and `source.listFilters`. |
| [x] | itch implements `source.browse`/`source.listFilters` with a search sidebar. It scrapes itch.io's public search HTML; there is no stable search API or pager, and sort is best-effort. |

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
| [x] | Updates and Downloads are top tabs in Library, and the game detail is a right sidebar. |
| [x] | Home loading and scan status: status line, skeleton tiles, folder name in the scan line, empty state instead of endless rocket tiles. Built and tested, not seen in a window. |
| [x] | Game detail is a right sidebar panel (about 420 px, overlay with backdrop on narrow windows). Built and tested, not seen in a window. |
| [x] | Game files are grouped by platform with icons. Built and tested, not seen in a window. |
| [x] | Updates (with one sub-tab per source) and Downloads are Library tabs and gone from the nav. Built and tested, not seen in a window. |
| [x] | Update history rows carry `source`; the source tab filters and pages on the server, so its pages stay full. |
| [x] | Downloads sampling is owned by the long-lived library state; it continues while the Downloads tab is closed and stops when no transfer is active. |
| [x] | F95 file kinds are `full`, `patch`, `mirror` and `attachment`; Full/Patch badges are intentional and other kinds stay unbadged. |
| [x] | The Library overview carries `kind`, `capabilities`, `sync` and `updateCheck`; the Updates page reads them from the overview instead of `PluginsList`. |
| [x] | Updates, Downloads and Incoming mount lazily on first visit and then stay mounted but hidden; in-flight and expanded state survives tab switches. Per-tab scroll still uses the shared page scroller. |
| [x] | Library sidebar and rail have no background or divider; choices are rows with a tinted selected state, no checkbox. |
| [x] | Library cards: no Installed badge, name and version in a hover overlay on the cover, denser 12 px grid. |
| [x] | Updates page shows "Last checked X ago · Next check in Y" from `nextRunAt`. |
| [x] | Notification center as a full page, with its bell at the bottom of the left rail. Built and tested, not seen in a window. |
| [x] | Rebuilt `bin/light-launcher` and `bin/light-launcher-instance` with the embedded UI. |
| [x] | Folder-scanned and hand-added games are imported into the library SQLite database as the local source; F95 and itch rows live in the same list and the overview lists all sources. Home and Library read the same rows. Migrated a copy of the real 258-local/106-F95/39-itch database. Not seen in a window. |
| [x] | F95 1.0.1 and itch 0.1.0 rebuilt against the current SDK, packed, installed, and `source.listFilters` answered through the host with real engine options. |
| [x] | Windows plugin binaries rebuilt and both `.llplugin` packages repacked; each Windows `.exe` now contains `source.listFilters`. The installed Linux copies are unchanged; a Windows install uses the new packages. |
| [ ] | Local games are now library rows with launch settings, but launching one from the Library and its prefix/Proton/playtime behavior has only been unit-tested. Check in the real window. |
| [x] | Plugin commands are covered in the command palette: entries carry the plugin/command ids, running one calls `PluginCall`, and a plugin without commands adds none. |
| [ ] | The plugin command palette entries are unit-tested only; check the badge, labels and result toast in a running window. |
| [x] | One-time cleanup removes the stale itch "API key is not configured" history rows (including the `(code -32603)` form); real failures stay. |
| [x] | Per-source "Check every" now sets a minimum gap before that source's games are due again. |
| [x] | The overview sends the `sync` flag; `canSyncGames` reads it and local games never show Sync now. |
| [x] | The first enabled plugin that claims a source id owns it; later plugins with a colliding id are marked broken with a clear error. |
| [x] | `pages`, `hooks`, `launchAdapters`, `importers` and `themes` are rejected in apiVersion 1 with a clear error; the SDK schema and docs no longer advertise them. |
| [x] | F95's initialize-time `settings.get` no longer times out: the supervisor held its state mutex across the handshake, so the host service blocked until the plugin gave up. Starts now use a separate start lock, and a regression test covers host calls during initialize. |
| [x] | The shared Dropdown has an optional search box for lists over 8 choices; select-type Library sidebar groups pass the plugin's `searchable` flag. |
| [x] | Sources expose the plugin games folder; scan-folder rows show a non-blocking warning when a folder overlaps it. Path matching is unit-tested. |
| [x] | Plugin update checks carry `not_found`, `rate_limited` or `untrusted_host`; the UI no longer matches error text. |
| [x] | IPC errors carry a stable `code`; `method_not_found`, `auth_required` and the plugin RPC codes reach `InvokeError`. |
| [x] | No `.btn` buttons remain in the plugin screens. PluginRow's full-row button and the sites disclosure stay raw on purpose: Button cannot express their layout/aria structure without neutralizing the kit. |
| [x] | The text marquee is one shared component used by GameTile and GridCard. |
| [x] | A test scans every renderer `.svelte` file and fails on new `class="btn` or `material-icons` spans; the one external-link anchor is explicitly allowed. |
| [x] | The remaining `material-icons` spans and `.btn` controls moved to `Icon`/`Button`/`IconButton`; only the GameDetail external-link anchor stays raw. |
| [x] | The Updates settings source list no longer calls `PluginsList`; it reads one local library overview and shows each source icon and label. |
| [x] | `SettingRow` has an optional icon slot; the Updates settings source rows use it instead of local row markup. |
| [x] | `Checkbox` is intentionally kept for `PluginConsent` only. |

## Browser mode

| Status | Item |
|---|---|
| [ ] | The `RestartApp` call emits a `restart-app` event that no code handles (`core/internal/services/appsvc/register.go` line ~85), so the restart after the transparency and display platform toggles does nothing. Fix or remove the event later. |
| [ ] | `SettingsState.setOpenInBrowser` has no test. `vitest.config.ts` has no Svelte plugin and no `@stores` alias, so the class cannot be imported in tests. Add both to test the setter. |
| [ ] | Check that `PickFile` and `PickFolder` work in a plain browser tab on Linux and on Windows. The renderer falls back to invoke when `window.llShell` is missing (`src/renderer/src/lib/transport.ts` around line 151). |
| [ ] | The `--web` flag opens the Windows default browser with `rundll32 url.dll,FileProtocolHandler <url>`. Confirm that the `#t=` token in the URL survives on Windows. Not tested from Linux. |

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
