[Back to README](../README.md)

# Releasing

This page explains how a new version of LightLauncher is made and published.

## How a release works

Commit messages decide the next version number. The project uses [Conventional Commits](https://www.conventionalcommits.org/).

| Step | Who              | What happens                                                                              |
| :--- | :--------------- | :---------------------------------------------------------------------------------------- |
| 1    | You              | Merge commits to `main` using Conventional Commit messages.                               |
| 2    | release-please   | Opens or updates a release pull request with the new version and `CHANGELOG.md`.          |
| 3    | You              | Merge the release pull request.                                                           |
| 4    | release-please   | Creates the version tag and a GitHub Release.                                             |
| 5    | Release workflow | Builds the Linux and Windows files, makes `SHA256SUMS`, and attaches them to the release. |

> [!NOTE]
> Step 5 starts only when the tag is pushed by a token that can trigger workflows. See [Secrets](#secrets).

## Version bumps

| Commit type                             | Version change                        |
| :-------------------------------------- | :------------------------------------ |
| `fix:`                                  | Patch, for example `1.0.0` to `1.0.1` |
| `feat:`                                 | Minor, for example `1.0.0` to `1.1.0` |
| `feat!:` or a `BREAKING CHANGE:` footer | Major, for example `1.0.0` to `2.0.0` |

Other types, such as `docs:` and `chore:`, do not bump the version on their own.

## Workflows

| File                                   | Runs on                   | Purpose                                                                                                |
| :------------------------------------- | :------------------------ | :----------------------------------------------------------------------------------------------------- |
| `.github/workflows/ci.yml`             | Push and pull request     | Go vet and tests on Linux and Windows, UI checks, GTK shell build, WebView2 cross build, workflow lint. |
| `.github/workflows/release-please.yml` | Push to `main`            | Opens the release pull request and creates the tag.                                                    |
| `.github/workflows/release.yml`        | Tag `v*`, or a manual run | Runs `scripts/build-linux.sh` and `scripts/build-windows.sh`, then publishes the release files.        |

## Release files

| File                                             | What it is                          |
| :----------------------------------------------- | :---------------------------------- |
| `light-launcher-<version>-linux-x64.tar.gz`      | Linux files in one archive          |
| `light-launcher-<version>-windows-x64.zip`       | Windows files, no install needed    |
| `light-launcher-<version>-windows-x64-setup.exe` | Windows installer (NSIS)            |
| `SHA256SUMS`                                     | Checksums for the three files above |

The build steps are described in [Building](building.md).

## Check a download

Check the file against `SHA256SUMS` before you run it.

Linux:

```sh
sha256sum -c SHA256SUMS
```

Windows (PowerShell or Command Prompt):

```text
certutil -hashfile light-launcher-<version>-windows-x64-setup.exe SHA256
```

Compare the result with the line for that file in `SHA256SUMS`.

## Manual run

Use this to test the build without a release.

1. Open **Actions** on GitHub.
2. Pick **Release** and click **Run workflow**.
3. Keep **dry-run** checked. The build runs, but nothing is published.

To publish by hand, uncheck **dry-run** and pick a tag in the branch list.

## Secrets

Add secrets in **Settings > Secrets and variables > Actions**.

| Secret                 | Required     | Used by              | What it is                                                                                                                           |
| :--------------------- | :----------- | :------------------- | :----------------------------------------------------------------------------------------------------------------------------------- |
| `RELEASE_PLEASE_TOKEN` | Optional     | `release-please.yml` | A personal access token or GitHub App token. Lets a release tag start `release.yml`. Without it, start the Release workflow by hand. |

`GITHUB_TOKEN` is created by GitHub for each run. You do not add it.

> [!WARNING]
> Never print a secret in a workflow log or commit it to the repository.

## Notes

- The Windows files are not signed yet. Windows shows a warning when you first run them.
- Release notes come from release-please. The Release workflow also asks for generated notes. Check the release body after the first run.
- The Release workflow runs on Ubuntu. It cross-builds the Windows files there.
