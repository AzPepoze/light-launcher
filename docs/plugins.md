[Back to README](../README.md)

# Plugins

A plugin adds a source to LightLauncher. A source is a website or service that lists games, such as F95 Zone or itch.io.

## What a plugin is

A plugin is a separate program with a small `plugin.json` file. LightLauncher starts the program and talks to it. If the program crashes, LightLauncher keeps running.

```text
  LightLauncher  <---- messages ---->  Plugin program
  (asks for games,                     (reads the website,
   downloads, sign-in)                  returns games)
```

Plugins are packed in a `.llplugin` file. That file is a zip archive with the program, the manifest and checksums.

## Install a plugin

| Source              | Command                                                             |
| :------------------ | :------------------------------------------------------------------ |
| A file on your disk | `light-launcher plugin install /path/to/plugin.llplugin`            |
| An `https://` link  | `light-launcher plugin install https://example.com/plugin.llplugin` |

The install dialog shows the plugin's permissions. The command line prints them after the install.

> [!NOTE]
> The launcher has no built-in plugin catalog. Install plugins from sources you trust.

## Update, turn off, or remove

| Task                         | Command                                           |
| :--------------------------- | :------------------------------------------------ |
| See installed plugins        | `light-launcher plugin list`                      |
| Update                       | `light-launcher plugin update <id>`               |
| Turn off                     | `light-launcher plugin disable <id>`              |
| Turn on                      | `light-launcher plugin enable <id>`               |
| Go back one version          | `light-launcher plugin rollback <id>`             |
| Remove                       | `light-launcher plugin remove <id>`               |
| Remove and delete saved data | `light-launcher plugin remove <id> --delete-data` |

`update` checks the place the plugin came from. A plugin installed from a file updates from that file path.

> [!WARNING]
> `remove --delete-data` deletes the plugin's data folder, with its settings and stored data. This cannot be undone.
>
> Removing a plugin also deletes its saved secrets, such as sign-in cookies and API keys.

## Updates

A plugin can name its update source in `plugin.json`:

```json
{ "repository": "github:owner/repo" }
```

| Where it came from | How it updates                                                                   |
| :----------------- | :------------------------------------------------------------------------------- |
| A GitHub release   | The Plugins page shows "Update available (x.y.z)". Press the button and confirm. |
| A file or folder   | No update badge. Install the new package again.                                  |

Updates are always manual. The launcher checks for releases once shortly after start, and when you press "Check for updates". It never installs an update by itself, and there is no auto-update setting.

The update opens the same consent screen as an install. New permissions are listed first. A package that names a different repository than the installed version shows a warning, so an update cannot quietly move a plugin to another repository.

The check and the download only use GitHub hosts. The launcher ignores releases without a `.llplugin` asset and verifies the release `SHA256SUMS` file when one is attached. A version that does not start after the update is rolled back to the previous one.

## Trust levels

The launcher checks each package before it installs it.

| Result                  | What it means                                                  | Installs?                    |
| :---------------------- | :------------------------------------------------------------- | :--------------------------- |
| Signed by a trusted key | The signature matches a key you trust.                         | Yes                          |
| Unsigned                | The package has no signature file.                             | Only with `--allow-unsigned` |
| Bad signature           | A signature file exists, but no trusted key matches it.        | No                           |
| Bad checksum            | A file does not match its checksum.                            | No                           |
| Not compatible          | The plugin needs a newer launcher or another protocol version. | No                           |

### Trusted keys

Trusted keys are public keys. You add one by putting it in this file, one key per line:

```text
<base folder>/plugins/trusted_keys
```

Lines that start with `#` are comments. There is no command for this yet.

> [!IMPORTANT]
> A package with a signature file is refused when no matching key is in `trusted_keys`. Add the key first.

> [!WARNING]
> `--allow-unsigned` installs code that nobody has checked. Use it only for code you trust.

## Permissions

Each plugin lists what it may use in its `plugin.json`. The launcher shows this list before install.

| Permission | What it lets the plugin do through the launcher                                                         |
| :--------- | :------------------------------------------------------------------------------------------------------ |
| `network`  | Make web requests to the listed sites only.                                                             |
| `fs`       | Read and write files in named folders, such as `$GAMES` or `$DOWNLOADS`.                                |
| `secrets`  | Save and read secrets, such as sign-in cookies, in a secure store. See [Security](security.md#secrets). |
| `browser`  | Open your browser and catch a download link.                                                            |
| `exec`     | Start other programs. See the note below.                                                               |

> [!WARNING]
> `exec` is listed in the permission list, but no launcher service checks it yet. A plugin program can still run other programs by itself. See [Security](security.md#honest-limits).

## Accounts and settings

Some sources need you to sign in.

| Source          | What you give                                                    | Where it is kept                                       |
| :-------------- | :--------------------------------------------------------------- | :----------------------------------------------------- |
| F95 Zone plugin | Your F95 Zone session cookie.                                    | The secret store. See [Security](security.md#secrets). |
| itch.io plugin  | Your itch.io API key. You can copy it from the itch desktop app. | The secret store.                                      |

To sign in, open the plugin's accounts section, then paste the value. The plugin checks it and saves it.

> [!WARNING]
> A session cookie gives full access to your account. Do not share it. Sign out of the site to end the session.

Plugin settings are a form the plugin defines. Your values are saved in `plugin-data/<id>/settings.json` in the base folder. See [File locations](usage.md#file-locations).

## Sleeping plugins

A plugin that nobody has called for 5 minutes is stopped to free its memory. The plugin shows "Sleeping" in the plugin list. The next request starts it again, and the result is the same as if it had stayed open.

Change the wait in `settings.json` in the base folder. Add `pluginIdleTimeoutMinutes` to the settings object:

```json
{ "pluginIdleTimeoutMinutes": 15 }
```

| Value   | Effect                                           |
| :------ | :----------------------------------------------- |
| `0`     | Never stop a plugin for being idle.              |
| `1`-`1440` | Minutes without a call before the plugin sleeps. |
| Missing | Use the default of 5 minutes.                    |

A plugin with a call in progress, an open browser step, or a running progress job keeps running.

## Known plugins

These two plugins are built as separate projects. They are not in the launcher.

| Plugin   | What it does                                                                              |
| :------- | :---------------------------------------------------------------------------------------- |
| F95 Zone | Reads forum threads with your account, finds download links, and checks for new versions. |
| itch.io  | Searches itch.io, reads the games you own, finds downloads, and checks for new uploads.   |

## Legal notice

> [!IMPORTANT]
>
> - Use plugins only with accounts you own.
> - Download only content you have the right to get.
> - Plugins do not bypass paywalls or access control.
> - Plugins from other authors are not part of the LightLauncher project.

## More

| Topic                    | Page                                  |
| :----------------------- | :------------------------------------ |
| Command list             | [Usage](usage.md#plugin-commands)     |
| Trust, checks and limits | [Security](security.md)               |
| Write your own plugin    | [Writing plugins](writing-plugins.md) |
