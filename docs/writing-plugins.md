[Back to README](../README.md)

# Writing plugins

This page gives an overview of how a plugin works. The full guide, with code, is in the plugin SDK repository (`light-launcher-plugin-sdk`), in its `docs/` folder:

| SDK page                   | Read it to learn                                                    |
| :------------------------- | :------------------------------------------------------------------ |
| `docs/writing-a-plugin.md` | How to build a first source plugin step by step.                    |
| `docs/manifest.md`         | Every field in `plugin.json`, and the folder names for permissions. |
| `docs/protocol.md`         | Every message the launcher and a plugin send.                       |

## Parts of a plugin

| Part            | What it is                                                                        |
| :-------------- | :-------------------------------------------------------------------------------- |
| `plugin.json`   | The manifest. It names the plugin, its version, its program, and what it may use. |
| Program         | One program per platform, such as `bin/linux-x64/my-plugin`.                      |
| Settings schema | Optional. A JSON Schema file. The launcher builds a settings form from it.        |
| Checksums       | A `SHA256SUMS` file with a checksum for each file in the package.                 |
| Signature       | Optional. A `SIGNATURE` file. See [Plugins](plugins.md#trust-levels).             |
| Repository      | Optional. `github:owner/repo` enables update checks on the Plugins page.          |

The package is a zip file with the `.llplugin` extension.

## Publish updates on GitHub Releases

Add the repository to `plugin.json`:

```json
{ "repository": "github:owner/repo" }
```

Publish each version as a GitHub release:

| Part      | Rule                                                                              |
| :-------- | :-------------------------------------------------------------------------------- |
| Tag       | The plugin version, such as `v2.1.0`.                                             |
| Package   | Name the asset `<id>-<version>.llplugin`, for example `my-plugin-2.1.0.llplugin`. |
| Checksums | Optional. Attach `SHA256SUMS` with `<sha256>  <id>-<version>.llplugin`.           |
| Notes     | The release body is shown as plain text on the plugin page.                       |

The launcher reads the latest release. A release with no `.llplugin` asset does not offer an update. Pre-releases are skipped unless the launcher enables them, and updates are always manual: the user presses Update and confirms the permissions.

The installed manifest must keep naming the same repository. An update that names another repository still installs, but the consent screen shows a warning.

## How the messages flow

Plugins use JSON-RPC 2.0. Each message is one line of JSON. The launcher writes to the plugin's standard input. The plugin writes to its standard output.

```text
  LightLauncher                              Plugin program
  -------------                              --------------
  initialize      ------------------------>
                  <------------------------  result: version, capabilities
  source.search   ------------------------>
                  <------------------------  http.request   (plugin asks the host)
  result: page    ------------------------>
                  <------------------------  result: games
  shutdown        ------------------------>
                  <------------------------  result: done, then the plugin exits
```

> [!IMPORTANT]
> Standard output carries messages only. Write logs through the host's log call. A stray line on standard output breaks the connection.

## How the launcher runs a plugin

| Step         | What happens                                                                                           |
| :----------- | :----------------------------------------------------------------------------------------------------- |
| Install      | The launcher unpacks the package, checks checksums and the signature, and checks the protocol version. |
| Start        | It starts the program for your platform and sends `initialize`. The plugin has 10 seconds to answer.   |
| Calls        | Each call from the user goes to the plugin. A call has 60 seconds to finish by default.                |
| Health check | Every 15 seconds the launcher pings the plugin. A plugin that does not answer is stopped.              |
| Crash        | The launcher restarts the plugin after a short wait. The wait doubles each time, up to 30 seconds.     |
| Crash limit  | Three crashes within 5 minutes turn the plugin off.                                                    |
| Sleep        | After 5 minutes without a call, the launcher stops the plugin. The next call starts it again.           |
| Stop         | The launcher sends `shutdown`. The plugin has a short time to finish, then it is stopped.              |

> [!IMPORTANT]
> A plugin may be stopped when idle, so do not keep important state only in memory. Save it through `storage.set` or your settings, and read it again on start or per call. Do not assume the same process stays alive between calls.

## What a plugin can ask the host for

For web requests, secrets, files and downloads, a plugin asks the launcher. The launcher checks the manifest before it does the work.

| Host service                                                                        | Needs in `plugin.json`   |
| :---------------------------------------------------------------------------------- | :----------------------- |
| `http.request`                                                                      | `network` lists the host |
| `secrets.get`, `secrets.set`, `secrets.delete`                                      | `secrets: true`          |
| `extract`                                                                           | `fs` lists a folder      |
| `downloads.enqueue` or `source.prepareDownload` with `root: "games"` or `install: true` | `fs` lists `$GAMES`      |
| `browser.assist`                                                                    | `browser: true`          |
| `storage.get`, `storage.set`, `downloads.enqueue`, `library.upsert`, `settings.get` | None                     |

> [!WARNING]
> A plugin that opens its own network connections is not stopped by the launcher. See [Security](security.md#honest-limits).

## Download hints

`source.prepareDownload` may return the same optional folder and install fields as `downloads.enqueue`. The launcher validates them with the same rules and queues the download with them.

| Result field       | Type    | Use it for                                                                                     |
| :----------------- | :------ | :--------------------------------------------------------------------------------------------- |
| `root`             | string  | `downloads` (default) or `games`.                                                              |
| `subfolder`        | string  | A relative folder under `root`, such as `Team/Demo`.                                           |
| `install`          | boolean | Install the archive after the download.                                                        |
| `installSubfolder` | string  | A relative folder under the games folder for the installed game. Needs `install`.              |
| `deleteArchive`    | boolean | Delete the archive after a good install. Omit it to use the launcher setting. Needs `install`. |

Omit every field to keep the default: save into the downloads folder and do not install. `root: "games"` and `install: true` need `$GAMES` in the plugin's `fs` permission; the launcher refuses them otherwise.

### Folder name rules

Every name in `subfolder` and `installSubfolder`, and every download `filename`, must be a plain name that works on Linux and Windows.

| Rule | Details |
| :--- | :--- |
| One segment | No `/` or `\`, and no drive letters. Join folders with `/` in `subfolder`. |
| Allowed characters | No `<`, `>`, `:`, `"`, `\|`, `?` or `*`. Spaces, dots inside the name, unicode and emoji are fine. |
| End of name | Must not end with a dot or a space. |
| Reserved names | Not a Windows device name, with or without an extension: `CON`, `PRN`, `AUX`, `NUL`, `COM1` to `COM9`, `LPT1` to `LPT9`. |
| Length | A name is at most 255 bytes and a whole `subfolder` path at most 240 bytes. |

Map the plugin's settings to these fields: a `downloadFolder` setting to `subfolder`, `installAfterDownload` to `install`, `installFolder` to `installSubfolder`, and `deleteArchive` to `deleteArchive`. Use `root: "games"` when the user picked a folder inside the games folder. Read the settings on each call, as described above.

## List the user's games

A source can list every game the user owns, follows or bookmarks there. The launcher reads the list page by page and saves the games in the library. Use this for large lists. Use `source.search` to browse a whole site.

To offer it, add the `listGames` capability to the source in `plugin.json`:

```json
{
  "contributes": {
    "sources": [
      { "id": "shop", "label": "Shop", "capabilities": ["browse", "listGames"], "tile": "poster" }
    ]
  }
}
```

### Request

The launcher sends `source.listGames` with these params.

| Param    | Type    | Rule                                                                 |
| :------- | :------ | :------------------------------------------------------------------- |
| `source` | string  | The source id, such as `shop`.                                       |
| `scope`  | string  | Always `mine` for now: the games the user owns, follows or bookmarked. |
| `cursor` | string  | Missing on the first page. Otherwise the `nextCursor` of the last page. |
| `limit`  | integer | From 1 to 200. The launcher asks for 100.                            |

### Result

| Field        | Type    | Rule                                                                 |
| :----------- | :------ | :------------------------------------------------------------------- |
| `games`      | array   | Games in the same shape as `source.search` results (`GameDTO`). At most `limit` games. |
| `nextCursor` | string  | Leave it out on the last page.                                       |
| `total`      | integer | Optional. The number of games, when the plugin knows it.             |

```json
{
  "games": [
    { "source": "shop", "sourceId": "1042", "url": "https://shop.example/g/1042", "title": "Example Game" }
  ],
  "nextCursor": "eyJvZmZzZXQiOjEwMH0",
  "total": 250
}
```

### Cursor and page rules

| Rule                                                              | Why                                                  |
| :---------------------------------------------------------------- | :--------------------------------------------------- |
| Make each cursor an opaque string that works on its own.          | The launcher does not read it.                       |
| Keep a cursor valid while the whole import runs, which can take minutes. | The launcher asks for the next page only after the last one. |
| Answer a bad cursor with an invalid params error (code `-32602`). | The import stops and reports the error.              |
| Do not return the same `sourceId` twice in one import.            | Repeated games are counted more than once.           |

> [!IMPORTANT]
> A bad `limit` or `scope` must return an invalid params error (code `-32602`). Do not clamp it silently.

### Speed and caching

| Advice                                                                 | Details                                                                          |
| :--------------------------------------------------------------------- | :------------------------------------------------------------------------------- |
| Answer each page quickly.                                              | The launcher waits for each page before it asks for the next one.                |
| Cache the upstream pages yourself.                                     | Keep them in memory or in `storage.set` with a time limit, such as 10 minutes.  |
| Do not keep the whole list in memory for a huge account.               | Fetch and answer one page at a time.                                             |

### Card shape

The `tile` field of the source sets the shape of its cards in the Library:

| `tile`   | Shape         | Use it for                  |
| :------- | :------------ | :-------------------------- |
| `poster` | 3:4 (default) | Tall cover art.             |
| `wide`   | 16:9          | Banner-style cover art.     |
| `square` | 1:1           | Icons and square covers.    |

A value other than these three fails the manifest check.

## Browse with a sidebar

A source can offer a filter sidebar and a browse page. The Library asks the plugin for one page of games at a time. Browse results are for display only: the launcher does not save them.

To offer it, add the `browse:filters` capability and a `sidebar` to the source:

```json
{
  "contributes": {
    "sources": [
      {
        "id": "shop",
        "label": "Shop",
        "capabilities": ["browse", "browse:filters"],
        "tile": "wide",
        "sidebar": [
          { "type": "search", "id": "text", "placeholder": "Search titles" },
          { "type": "multi", "id": "tags", "label": "Tags", "dynamic": true, "searchable": true },
          {
            "type": "sort",
            "id": "sort",
            "orderable": true,
            "options": [{ "value": "title", "label": "Title" }],
            "default": { "value": "title", "order": "asc" }
          }
        ]
      }
    ]
  }
}
```

A source with a `sidebar` must have the `browse:filters` capability. Without it the manifest fails the check.

### Sidebar groups

| `type`    | What the user sees            | Fields                                                  |
| :-------- | :---------------------------- | :------------------------------------------------------ |
| `search`  | A text box.                   | `id`, `placeholder` (optional).                         |
| `toggles` | On and off rows.              | `id`, `options`.                                        |
| `select`  | One choice from a list.       | `id`, `label`, and `options` or `dynamic`.              |
| `multi`   | Several choices from a list.  | `id`, `label`, and `options` or `dynamic`, `searchable`. |
| `sort`    | A sort choice and direction.  | `id`, `options`, `orderable`, `default` (optional).     |

| Rule                                                             | Why                                              |
| :--------------------------------------------------------------- | :----------------------------------------------- |
| Group ids are lowercase letters, digits or dashes, unique per source. | The launcher sends the values under these ids.   |
| Options have a `value` and a `label`. Values are unique in a group. | Each choice has one value to send back.          |
| `dynamic` takes the choices from `source.listFilters`. Do not also list `options`. | The list can be large, such as all tags.         |
| A `default` is only for a `sort` group, and its `value` must be one of its options. | Sets the sort the Library starts with.           |
| `order` is `asc` or `desc`.                                      | The direction of a sort.                         |

A bad group, option or default fails the manifest check with the field name.

### source.browse

The launcher sends `source.browse` for each page.

| Param     | Type    | Rule                                                                              |
| :-------- | :------ | :-------------------------------------------------------------------------------- |
| `source`  | string  | The source id, such as `shop`.                                                    |
| `query`   | string  | Optional. The search text.                                                        |
| `filters` | object  | Optional. Keys are group ids. Values are a string, a list of strings or a boolean. |
| `sort`    | string  | Optional. A sort value from the `sort` group.                                     |
| `order`   | string  | Optional. `asc` or `desc`.                                                        |
| `cursor`  | string  | Missing on the first page. Otherwise the `nextCursor` of the last page.          |
| `limit`   | integer | From 1 to 100.                                                                    |

The result has `games` (in the same shape as `source.search` results), `nextCursor` (left out on the last page) and `total` (optional, the number of games that match).

| Rule                                                                 | Why                                                    |
| :------------------------------------------------------------------- | :----------------------------------------------------- |
| Ignore filter keys that are not known groups.                        | The launcher may send a group this plugin does not use. |
| A cursor works only with the same `query`, `filters`, `sort` and `order` it came from. | Paging with changed filters would skip or repeat games. |
| Answer a cursor you do not know with an invalid params error (code `-32602`) and `"data": { "code": "invalid_cursor" }`. | The launcher then shows "start again from the first page". |
| Answer a bad `limit`, `order` or `sort` with an invalid params error (code `-32602`). | Do not clamp it silently.                              |

### source.listFilters

The launcher sends `source.listFilters` to fill a `dynamic` group, such as the tag list. `query` is the text typed in the group's search box.

| Param    | Type    | Rule                                                 |
| :------- | :------ | :--------------------------------------------------- |
| `source` | string  | The source id.                                       |
| `group`  | string  | The id of a `dynamic` group.                         |
| `query`  | string  | Optional. The text the person typed.                 |
| `limit`  | integer | From 1 to 200.                                       |

The result is `options`: a list of `{ "value", "label", "count" }`. `count` is optional and counts the whole source, not the current filters.

### Time limits

| Method               | The launcher waits for | After that                              |
| :------------------- | :--------------------- | :-------------------------------------- |
| `source.browse`      | 20 seconds.            | It stops waiting and returns an error.  |
| `source.listFilters` | 10 seconds.            | It stops waiting and returns an error.  |

The launcher does not retry a call by itself.

## Settings form

The launcher draws a settings form from a JSON Schema file. Point to it in `plugin.json`:

```json
{ "contributes": { "settings": { "$ref": "settings.schema.json" } } }
```

Here is a small schema:

```json
{
	"type": "object",
	"properties": {
		"folder": { "type": "string", "title": "Download folder", "x-ll-widget": "folder" },
		"hosts": {
			"type": "array",
			"title": "Host order",
			"items": { "type": "string", "enum": ["a", "b"] },
			"uniqueItems": true,
			"default": ["a", "b"]
		},
		"timeout": {
			"type": "integer",
			"minimum": 1,
			"maximum": 60,
			"default": 10,
			"x-ll-advanced": true
		}
	}
}
```

### Supported keywords

| Keyword                               | Used on         | Effect                                                                     |
| :------------------------------------ | :-------------- | :------------------------------------------------------------------------- |
| `type`                                | Every field     | `string`, `number`, `integer`, `boolean`, `array` or `object`. Others are not drawn. |
| `properties`                          | Object          | The fields of a nested object. The form draws them as a group under the object title. |
| `required`                            | Object          | Names the fields that must have a value.                                   |
| `title`                               | Every field     | The label. Defaults to the property name.                                  |
| `description`                         | Every field     | A hint under the input.                                                    |
| `examples`                            | Every field     | Shown under the input as "Examples: ...".                                  |
| `default`                             | Every field     | Used when no value is saved. The form offers "Reset to default".           |
| `enum`                                | String, number  | A select. For a list, the items are picked from it.                        |
| `minimum`, `maximum`                  | Number, integer | The allowed range.                                                         |
| `minLength`, `maxLength`              | String          | The length in characters.                                                  |
| `items`                               | Array           | The item type: `string`, or an `enum` of strings or numbers.               |
| `minItems`, `maxItems`, `uniqueItems` | Array           | The list size, and whether items may repeat.                               |
| `additionalProperties`                | Top level       | `false` rejects keys that are not in the schema.                           |

### Launcher extensions

These keywords start with `x-ll-`. Other tools ignore them.

| Extension       | Values           | Effect                                                                                                                         |
| :-------------- | :--------------- | :----------------------------------------------------------------------------------------------------------------------------- |
| `x-ll-group`    | Any text         | Fields with the same group sit under one heading. Groups appear in the order they first show up.                               |
| `x-ll-widget`   | `folder`, `file` | A path box for a string. A Browse button opens the system picker when the desktop app provides one. The text box always works. |
| `x-ll-advanced` | `true`           | The field moves into a collapsed "Advanced" section at the end of the form.                                                    |

### Saving and reading

| Event                       | What happens                                                                                                 |
| :-------------------------- | :----------------------------------------------------------------------------------------------------------- |
| User clicks Save            | The launcher checks every value. Each bad field gets its own message. If any field is bad, nothing is saved. |
| Save succeeds               | The settings file is replaced in one step. A crash during the save keeps the old file.                       |
| Plugin calls `settings.get` | It gets the saved values, with the schema `default` for each key that is not saved.                          |
| Save succeeds, plugin runs  | The plugin gets a `settings.changed` notification with `{"settings": {...}}`. The plugin is not restarted.   |

> [!TIP]
> Call `settings.get` each time your plugin needs a value, or reload when `settings.changed` arrives. A plugin that keeps the values from its start keeps the old ones until it restarts.

## Next steps

- Read the SDK guide, starting with `docs/writing-a-plugin.md`.
- Check your plugin with the SDK's `llp validate`, `llp test` and `llp pack` commands.
- Install the package with `light-launcher plugin install`. See [Usage](usage.md#plugin-commands).

## More

| Topic                     | Page                    |
| :------------------------ | :---------------------- |
| Install, update and trust | [Plugins](plugins.md)   |
| Permissions and limits    | [Security](security.md) |
| Commands                  | [Usage](usage.md)       |
