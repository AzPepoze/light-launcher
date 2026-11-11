[Back to README](../README.md)

# Releasing

This page explains how a new version of LightLauncher is made and published.

## How a release works

Release versions are picked by hand. The project uses [Conventional Commits](https://www.conventionalcommits.org/) to record changes.

| Step | Who              | What happens                                                                                    |
| :--- | :--------------- | :---------------------------------------------------------------------------------------------- |
| 1    | You              | Merge commits to `main` using Conventional Commit messages.                                     |
| 2    | You              | Pick the new version and push a tag, for example `git tag v1.1.0` then `git push origin v1.1.0`. |
| 3    | Release workflow | Builds the Linux and Windows files, makes `SHA256SUMS`, and publishes them in a GitHub Release.  |

GitHub generates the release notes from the commits and pull requests in the release.

## Version bumps

Pick the version bump from the commits since the last tag.

| Commit type                             | Version change                        |
| :-------------------------------------- | :------------------------------------ |
| `fix:`                                  | Patch, for example `1.0.0` to `1.0.1` |
| `feat:`                                 | Minor, for example `1.0.0` to `1.1.0` |
| `feat!:` or a `BREAKING CHANGE:` footer | Major, for example `1.0.0` to `2.0.0` |

Other types, such as `docs:` and `chore:`, do not need a release on their own.

## Workflows

| File                            | Runs on                   | Purpose                                                                                                |
| :------------------------------ | :------------------------ | :----------------------------------------------------------------------------------------------------- |
| `.github/workflows/ci.yml`      | Push and pull request     | Go vet and tests on Linux and Windows, UI checks, GTK shell build, WebView2 cross build, workflow lint. |
| `.github/workflows/release.yml` | Tag `v*`, or a manual run | Runs `scripts/build-linux.sh` and `scripts/build-windows.sh`, then publishes the release files.        |

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

## Notes

- The Windows files are not signed yet. Windows shows a warning when you first run them.
- The Release workflow asks GitHub for generated release notes. Check the release body after the first run.
- The Release workflow runs on Ubuntu. It cross-builds the Windows files there.
