[Back to README](../README.md)

# Security

This page explains what LightLauncher protects, and what it does not.

## Summary

| Area      | What is done                                                            | Limit                                                        |
| :-------- | :---------------------------------------------------------------------- | :----------------------------------------------------------- |
| Local API | Listens on `127.0.0.1` only, on a new port each run. Needs a token.     | Any program that runs as your user can read the lock file.   |
| Plugins   | The launcher checks each plugin's permissions before it runs a service. | Not an OS sandbox. A plugin runs as you.                     |
| Signing   | Packages can be signed with ed25519 keys you trust.                     | No keys are trusted by default. Most packages are unsigned.  |
| Secrets   | Sign-in values go in a secure store.                                    | The fallback is an encrypted file next to its key.           |
| Downloads | Plugin downloads use `https://` only, with a 256 MB limit.              | Checksums only show the files match the list in the package. |

## Local API

The window talks to the launcher over a local web address. The launcher checks each request.

```text
  Window (UI)  ---- X-LL-Token header ---->  127.0.0.1:<random port>
                                             |
                                             +-- rejects other addresses
                                             +-- rejects other host names
                                             +-- rejects other origins
                                             +-- rejects a wrong or missing token
```

| Check     | Rule                                                                                        |
| :-------- | :------------------------------------------------------------------------------------------ |
| Address   | Must be a loopback address, such as `127.0.0.1`.                                            |
| Host name | Must match the launcher's own address.                                                      |
| Origin    | If sent, must match the launcher's own address.                                             |
| Token     | Every `/api/` request needs the `X-LL-Token` header. The token is 64 random hex characters. |

The token is new each time the launcher starts. It is sent in the URL fragment (`#t=...`). A fragment is not sent to the server in the request line.

The launcher saves the port and token in `launcher.lock` in the base folder. On Linux, the file is readable by your user only. A second start uses it to hand over a game. See [File locations](usage.md#file-locations).

> [!WARNING]
> With `--no-shell`, the launcher logs an address that includes the token. Do not paste that log anywhere public.

## Plugin permissions

A plugin declares its permissions in `plugin.json`. The launcher checks them before it runs each host service.

| Permission | Checked by the launcher                           | What happens without it                                            |
| :--------- | :------------------------------------------------ | :----------------------------------------------------------------- |
| `network`  | Web requests (`http.request`) to the listed hosts | Request is refused with a permission error.                        |
| `fs`       | File extraction, by folder                        | Request is refused. A path outside the granted folders is refused. |
| `secrets`  | Secret read, write and delete                     | Request is refused.                                                |
| `browser`  | Start the embedded browser window                 | Request is refused.                                                |
| `exec`     | Not checked by any service yet                    | Nothing is stopped. See below.                                     |

## Browser assist

A plugin with the `browser` permission can ask the launcher to open the embedded browser window. This is a manual step, not automation. The user signs in, waits, solves any challenge and presses the download button by hand. The window shows a normal web page. The host does not click, type, fill forms, solve CAPTCHAs, spoof fingerprints or hide that the browser is embedded, and it passes no stealth flags.

| Area         | What the host does                                                                                        |
| :----------- | :-------------------------------------------------------------------------------------------------------- |
| Page address | Starts `light-launcher-shell --assist <url> --profile <dir>`. Only `http` and `https` addresses with a host are accepted. |
| Window       | Reads newline-delimited JSON from the shell's stdout. Logs go to stderr and are not parsed.               |
| Download     | Accepts one `download` event: address, file name, size and the headers the shell knows (`Cookie`, `Referer`, `User-Agent` when present). The shell cancels that download, so the browser writes nothing. |
| Limits       | One window at a time. Each output line is capped, and so is the whole output. Unknown events and malformed lines are ignored. |
| Stop         | Kills the shell and its helpers on a timeout, a cancel, a closed window, a capture, or when the launcher exits. |

Cookies and logins stay in `<base folder>/browser-profiles/<plugin id>`. Anything that runs as your user can read that folder. Sign out of the site or delete the folder to remove the saved login. Set `browserAssistEnabled` to `false` in `settings.json` to turn the whole feature off.

## Trust and signing

A package has a `SHA256SUMS` file with a checksum for every file. The launcher refuses a package when a file does not match, or when a file is not listed.

A package can also have a `SIGNATURE` file. It is an ed25519 signature of the `SHA256SUMS` file. The launcher accepts it only when a key in `plugins/trusted_keys` verifies it.

| Case                               | Result                                           |
| :--------------------------------- | :----------------------------------------------- |
| No `SIGNATURE` file                | Unsigned. Installs only with `--allow-unsigned`. |
| `SIGNATURE` matches a trusted key  | Signed. Installs.                                |
| `SIGNATURE` matches no trusted key | Refused, even with `--allow-unsigned`.           |

> [!IMPORTANT]
> A checksum shows that the files match the list. It does not show who made the list. Only the signature does that. See [Plugins](plugins.md#trust-levels).

## Secrets

Sign-in values, such as cookies and API keys, are kept in a secure store. Removing a plugin deletes them.

| System                               | Where the secret goes                                                            |
| :----------------------------------- | :------------------------------------------------------------------------------- |
| Linux with `secret-tool` (libsecret) | The desktop keyring.                                                             |
| Linux without `secret-tool`          | An encrypted file in the plugins folder. The key is in a second file next to it. |
| Windows                              | Windows Credential Manager.                                                      |
| Other systems                        | The encrypted file, as on Linux without `secret-tool`.                           |

The encrypted file uses AES-256-GCM. Its key file is only readable by your user. Anything that runs as your user can still read both files.

## Honest limits

- **Not an OS sandbox.** Plugin programs run as you. They can read and write files you can read and write.
- **Plugins can open their own connections.** The launcher blocks only the services it provides. The SDK docs say the same.
- **`exec` is not enforced.** The permission is listed, but no service checks it. A plugin program can start other programs by itself.
- **Trust is only as good as your keys.** No keys are trusted by default. Unsigned packages can contain any code.
- **Saved pages can hold personal data.** The F95 Zone plugin can save failed pages when you set `dumpDir`. Those pages can contain your account name.

## What you can do

| Action                                                    | Why                                                         |
| :-------------------------------------------------------- | :---------------------------------------------------------- |
| Install plugins only from sources you trust.              | A plugin program runs as you.                               |
| Add a key to `trusted_keys` only if you know who owns it. | Trusted keys let packages install without a warning.        |
| Sign out of a site before you remove its plugin.          | Removing a plugin does not delete its saved sign-in values. |
| Keep `launcher.lock` private.                             | It holds the API token.                                     |
| Do not share session cookies or API keys.                 | They give full access to your accounts.                     |

## More

| Topic                                       | Page                                  |
| :------------------------------------------ | :------------------------------------ |
| Plugin install and trust                    | [Plugins](plugins.md)                 |
| Plugin protocol and permissions for authors | [Writing plugins](writing-plugins.md) |
| File locations                              | [Usage](usage.md#file-locations)      |
