[Back to README](../README.md)

# Community plugin index

A community plugin index is a static HTTPS file, usually named `plugins.json`, that lists installable plugin packages. The launcher can show this list in the **Community** dialog on the Plugins page.

The index is only a catalog. It does not install anything, and it does not make a package trusted. Every install still opens the permission preview and consent screen, checks the package checksum, and checks the signature against your trusted keys.

## Set the index URL

Open **Settings → Community plugins**, enter the URL in **Community plugin index**, and press **Save**. The field also saves when it loses focus. An empty field disables the community list.

The same value can be set in `settings.json` in the base folder:

```json
{ "pluginRegistryURL": "https://plugins.example/plugins.json" }
```

| Rule        | What it means                                                              |
| :---------- | :------------------------------------------------------------------------- |
| Scheme      | `https` only. An `http` URL, or an `https` redirect to `http`, is refused.  |
| Empty value | Disables the community list. The launcher makes no request.                |
| Spaces      | Trimmed when the value is saved.                                           |
| Fetching    | The index is read when you open the Community list, not when Settings opens or on a timer. |

A fetch failure appears in the Community dialog with a **Retry** button. It does not disable the Plugins page, and it does not change the saved URL.

## File shape

The file is a JSON object with a version and a list of entries:

```json
{
  "version": 1,
  "plugins": [
    {
      "id": "demo",
      "name": "Demo plugin",
      "description": "Finds demo games.",
      "author": "Ada",
      "repository": "github:ada/demo",
      "version": "1.2.3",
      "downloadUrl": "https://plugins.example/demo-1.2.3.llplugin",
      "sha256": "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
      "homepage": "https://example.test/demo"
    }
  ]
}
```

Top-level fields:

| Field     | Required | Rule                                                       |
| :-------- | :------- | :--------------------------------------------------------- |
| `version` | Yes      | The integer `1`. Other values are refused by this launcher. |
| `plugins` | Yes      | An array of entries. An empty array is valid.              |

Entry fields:

| Field         | Required | Rule                                                                                                  |
| :------------ | :------- | :---------------------------------------------------------------------------------------------------- |
| `id`          | Yes      | 2 to 40 characters: lowercase letters, digits and dashes, starting with a letter. Must be unique.     |
| `name`        | Yes      | 1 to 80 characters, not only spaces.                                                                  |
| `version`     | Yes      | A semantic version such as `1.2.3`; a prerelease or build suffix is allowed.                          |
| `downloadUrl` | Yes      | An absolute `https` URL with a host, pointing at the `.llplugin` package.                              |
| `sha256`      | Yes      | 64 hexadecimal digits for the package at `downloadUrl`. Upper or lower case; the launcher normalizes it. |
| `description` | No       | At most 1000 characters.                                                                              |
| `author`      | No       | At most 80 characters.                                                                                |
| `repository`  | No       | `github:owner/repo` or an `https` URL. Metadata only; `plugin.json` in the package still controls updates. |
| `homepage`    | No       | An `https` URL shown as the entry's homepage.                                                         |

Unknown fields are ignored, so an index can carry extra metadata such as a generation date. Invalid JSON, a trailing extra value, a duplicate `id`, or a failed entry check refuses the whole index.

### Limits

| Limit            | Value                                                   |
| :--------------- | :------------------------------------------------------ |
| Entries          | At most 500.                                            |
| Response size    | At most 512 KiB.                                        |
| Fetch timeout    | 10 seconds.                                             |
| Redirects        | At most 5, and every hop must stay on `https`.          |
| HTTP status      | `200` only.                                             |

## Host the index

Any static HTTPS host works, such as GitHub Pages, Cloudflare Pages, an S3 bucket behind CloudFront, or an nginx server:

- Serve the file with `Content-Type: application/json; charset=utf-8`.
- No authentication is needed, and no CORS header is needed: the launcher process fetches the file, not browser page code.
- Host every `downloadUrl` package over HTTPS too.
- Keep the file below 512 KiB and let a request finish within 10 seconds.
- Replace the file as one atomic write, so a reader never sees half an index.

## Packages are still checked

The index is not signed. It only tells the launcher where a package is and which digest it should have.

| Check                 | What happens                                                                                           |
| :-------------------- | :----------------------------------------------------------------------------------------------------- |
| Digest                | The package must match the entry's `sha256` before it is unpacked.                                     |
| Checksums             | Every file must match the package's `SHA256SUMS`.                                                      |
| Signature             | A `SIGNATURE` must match a key in [Trusted plugin keys](plugins.md#trusted-keys).                      |
| Unsigned package      | The consent screen shows the unsigned warning, and the install goes ahead only after you accept it.    |
| Permissions and preview | The usual preview lists what the plugin may use, and nothing is installed until you tick the consent. |

The launcher never installs from the index by itself. It never updates a plugin from the index by itself either. A community entry starts the same manual install flow as a package you picked from disk.
