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

Map the plugin's settings to these fields: a `downloadFolder` setting to `subfolder`, `installAfterDownload` to `install`, `installFolder` to `installSubfolder`, and `deleteArchive` to `deleteArchive`. Use `root: "games"` when the user picked a folder inside the games folder. Read the settings on each call, as described above.

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
| `type`                                | Every field     | `string`, `number`, `integer`, `boolean` or `array`. Others are not drawn. |
| `title`                               | Every field     | The label. Defaults to the property name.                                  |
| `description`                         | Every field     | A hint under the input.                                                    |
| `examples`                            | Every field     | Shown under the input as "Examples: ...".                                  |
| `default`                             | Every field     | Used when no value is saved. The form offers "Reset to default".           |
| `enum`                                | String, number  | A select. For a list, the items are picked from it.                        |
| `minimum`, `maximum`                  | Number, integer | The allowed range.                                                         |
| `minLength`, `maxLength`              | String          | The length in characters.                                                  |
| `items`                               | Array           | The item type: `string`, or an `enum` of strings or numbers.               |
| `minItems`, `maxItems`, `uniqueItems` | Array           | The list size, and whether items may repeat.                               |
| `required`                            | Top level       | Names the fields that must have a value.                                   |
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
