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

| File                                   | Runs on                   | Purpose                                                                               |
| :------------------------------------- | :------------------------ | :------------------------------------------------------------------------------------ |
| `.github/workflows/ci.yml`             | Push and pull request     | Lint, format check, typecheck, tests, `go vet`, `go test`. Runs on Linux and Windows. |
| `.github/workflows/release-please.yml` | Push to `main`            | Opens the release pull request and creates the tag.                                   |
| `.github/workflows/release.yml`        | Tag `v*`, or a manual run | Builds and publishes the release files.                                               |
| `.github/workflows/aur.yml`            | Manual run only           | Builds an updated `PKGBUILD` as an artifact. Does not push.                           |

## Release files

| File type          | Built on | What it is                     |
| :----------------- | :------- | :----------------------------- |
| `.AppImage`        | Linux    | Single file app                |
| `.tar.gz`          | Linux    | Archive                        |
| `.exe` (installer) | Windows  | Installer (NSIS)               |
| `.exe` (portable)  | Windows  | Portable app, no install       |
| `SHA256SUMS`       | Linux    | Checksums for every file above |

> [!IMPORTANT]
> Check a download with `sha256sum -c SHA256SUMS` before you run it.

## Manual run

Use this to test the build without a release.

1. Open **Actions** on GitHub.
2. Pick **Release** and click **Run workflow**.
3. Keep **dry-run** checked. The build runs, but nothing is published.

To publish by hand, uncheck **dry-run** and pick a tag in the branch list.

## Secrets

Add secrets in **Settings > Secrets and variables > Actions**.

| Secret                  | Required     | Used by              | What it is                                                                                                                           |
| :---------------------- | :----------- | :------------------- | :----------------------------------------------------------------------------------------------------------------------------------- |
| `RELEASE_PLEASE_TOKEN`  | Optional     | `release-please.yml` | A personal access token or GitHub App token. Lets a release tag start `release.yml`. Without it, start the Release workflow by hand. |
| `WINDOWS_CERT_BASE64`   | Optional     | `release.yml`        | The Windows code signing certificate (`.pfx` file) as base64 text. Without it, the Windows files are not signed.                     |
| `WINDOWS_CERT_PASSWORD` | Optional     | `release.yml`        | The password for the `.pfx` file. Needed when `WINDOWS_CERT_BASE64` is set.                                                          |
| `AUR_SSH_KEY`           | Not used yet | Future AUR push      | An SSH private key registered on the AUR. Add it when the AUR push step is written.                                                  |

`GITHUB_TOKEN` is created by GitHub for each run. You do not add it.

> [!WARNING]
> Never print a secret in a workflow log or commit it to the repository.

## Notes

- The Windows packaging step is separate from the rest of the build. The Electron shell will be replaced later, so that step may change or move.
- Release notes come from release-please. The Release workflow also asks for generated notes. Check the release body after the first run.
