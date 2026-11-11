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
| [ ] | First release needs `Release-As: 2.0.0` (the manifest still says 1.0.0). |
| [ ] | Publish the plugin SDK. The plugins use a local `file:` link, so a fresh clone cannot install. |
| [ ] | Run `actionlint` and `shellcheck` on the workflows and scripts (not installed yet). |
| [x] | Make the plugin SDK and plugin repos public and push them. |
| [ ] | Plugin registry index (`plugins.json`) for community installs. |
| [ ] | Sign plugin packages and publish the public key. |

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
| [ ] | The scheduler reads plugin sync intervals once at start, so a plugin installed later or a changed interval needs a restart. |
| [ ] | Updating a dropped-in plugin without a `repository` field (drop the new version again). A dropped-in plugin that names a repository does update from GitHub. |
| [x] | Settings form built from each plugin's `settings.schema.json`. |
| [ ] | Nested objects in a settings schema are not drawn (only top-level fields). |
| [ ] | SDK docs for the `settings.changed` notification and the `x-ll-*` schema extensions. |
| [ ] | Check the settings form by eye in the real window (only typecheck, tests and build so far). |
| [x] | Host options for a download folder, a subfolder per game, install after download and delete archive, including the optional `source.prepareDownload` result hints. |
| [ ] | The game file download button is wired to `source.prepareDownload` but only tested with fakes and the fixture plugin. Click it in the real window with a plugin that opens the browser step. |
| [ ] | A file saved before this version has no plugin file id, so its `source.prepareDownload` call gets an empty `id`. Adding the game again from its link saves the ids. There is no button that only refreshes files. |
| [ ] | Closing the browser window during a file download is covered by a unit test only. Check the error text in the real window. |
| [ ] | The Download button's busy text ("Finish the step in the browser window") is checked by typecheck and build only. Look at it in the real window. |
| [x] | The launcher never calls `source.sync` and never schedules update checks. Plugin sync and interval settings have no effect until a scheduler exists. |
| [x] | A running plugin gets a `settings.changed` notice when settings are saved (no restart; plugins must read settings per call). |
| [x] | Embedded browser assist on Linux (Qt): the host starts the shell in assist mode and reads its download event. Run in a real window against a local test page. |
| [ ] | No plugin calls the host `browser.assist` service yet (F95 still uses its own Chrome code), so a real plugin download through the embedded window is untested. |
| [ ] | Embedded browser assist on Windows (WebView2) is untested on Windows. |
| [ ] | If the request that started a browser step ends first, the browser window stays open until its own time limit. The host does not pass that cancel on to `browser.assist` yet. |
| [x] | Qt assist mode reads cookies restored from disk on Qt 6.11 from the profile cookie database, where QtWebEngine stores the values in plain text. Encrypted values would be skipped. |
| [ ] | Plugin secrets saved before the cleanup change are not removed when the plugin is removed. |
| [x] | Stop plugins that stay idle for 5 minutes, with `pluginIdleTimeoutMinutes` to change or disable it. |
| [ ] | Check the "Sleeping" plugin badge by eye in the real window (only typecheck, tests and build so far). |
| [ ] | OS-level sandbox for plugins (permissions only cover host services today). |
| [ ] | Resource limits for plugins (CPU, memory). |
| [x] | The idle-stop tests timed out under load: the watcher armed its timer for "wait N seconds" a moment after it computed N, so a clock step in between moved the deadline. It now arms the timer for an absolute deadline (240 stress runs, no failures). |

## F95 plugin

| Status | Item |
|---|---|
| [ ] | The file-host resolvers (Pixeldrain, Mega, MediaFire and others) are not called by `prepareDownload`, so only direct links work. Decide whether to connect them. |
| [ ] | Ask F95 Zone staff for written permission. Until then the README says it is not approved. |
| [x] | Contact line in the README (an issue on the now public repo). |
| [ ] | Cannot set: patch-over-full, skip extras, include pre-releases (the parser does not keep that information). |
| [ ] | Delete the F95 plugin's own CDP code once the host browser assist service is used. |

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
| [x] | Cached screenshots are served through the media route and shown in the detail panel. |
| [ ] | Covers behind a login (F95) may fail because media fetches send no cookies. |
| [x] | Remove the old `/api/cover/{id}` route (the media route replaced it). |
| [x] | `make build` embeds the UI, so the launcher starts without `--ui`. |
| [ ] | The NativeWayland setting is saved but has no effect; Qt picks the display system. |
| [x] | Cover and screenshot images are evicted from memory least-recently-used first (200 entries). |
| [x] | Source filter shows labels from the library instead of raw ids. |
| [x] | The game detail can link an existing install folder. |
| [x] | The Updates page shows the update-check history with paging. |
| [ ] | Check the new Library page and drop zone by eye in the real window. |

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
| [ ] | `llp validate` does not check that the settings file exists (pack fails instead). |
| [ ] | The `parseRef` parameter check differs between itch (no validation) and f95 (strict). |
| [ ] | Update flow downloads a package twice (once to preview, once to install). |
| [ ] | Failed plugin installs leave their upload folder until the next start. |

## Product choices still open

| Status | Item |
|---|---|
| [ ] | Update the launcher itself from its GitHub Releases (separate install step on each OS). Not now. |
| [ ] | Optional auto-update for plugins when permissions do not change. Not wanted for now: updates are always manual. |
| [ ] | Deep links (`lightlauncher://`). |
| [ ] | Thai and English text, and accessibility checks. |
| [ ] | macOS support (out of scope for now). |
