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

The package is a zip file with the `.llplugin` extension.

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
| Stop         | The launcher sends `shutdown`. The plugin has a short time to finish, then it is stopped.              |

## What a plugin can ask the host for

For web requests, secrets, files and downloads, a plugin asks the launcher. The launcher checks the manifest before it does the work.

| Host service                                                                        | Needs in `plugin.json`   |
| :---------------------------------------------------------------------------------- | :----------------------- |
| `http.request`                                                                      | `network` lists the host |
| `secrets.get`, `secrets.set`, `secrets.delete`                                      | `secrets: true`          |
| `extract`                                                                           | `fs` lists a folder      |
| `browser.assist`                                                                    | `browser: true`          |
| `storage.get`, `storage.set`, `downloads.enqueue`, `library.upsert`, `settings.get` | None                     |

> [!WARNING]
> A plugin that opens its own network connections is not stopped by the launcher. See [Security](security.md#honest-limits).

## Next steps

- Read the SDK guide, starting with `docs/writing-a-plugin.md`.
- Check your plugin with the SDK's `llp validate`, `llp test` and `llp pack` commands.
- Install the package with `light-launcher plugin install`. See [Usage](usage.md#plugin-commands).

> [!NOTE]
> The plugin command in the launcher is not wired in this version. See [Known gaps](usage.md#known-gaps).

## More

| Topic                     | Page                    |
| :------------------------ | :---------------------- |
| Install, update and trust | [Plugins](plugins.md)   |
| Permissions and limits    | [Security](security.md) |
| Commands                  | [Usage](usage.md)       |
