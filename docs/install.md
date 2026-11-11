[Back to README](../README.md)

# Install

This page shows how to install LightLauncher on Linux and Windows.

> [!NOTE]
> On Arch, build from the PKGBUILD files in `install/arch`.

## Pick a file

| System          | Option                                           | Section                                   |
| :-------------- | :----------------------------------------------- | :---------------------------------------- |
| Arch Linux      | Build from source with `PKGBUILD`                | [Arch from source](#arch-from-source)     |
| Arch Linux      | Use the release files with `PKGBUILD-bin`        | [Arch release files](#arch-release-files) |
| Any other Linux | `light-launcher-<version>-linux-x64.tar.gz`      | [Tar archive](#tar-archive)               |
| Windows         | `light-launcher-<version>-windows-x64-setup.exe` | [Windows installer](#windows-installer)   |
| Windows         | `light-launcher-<version>-windows-x64.zip`       | [Windows portable](#windows-portable)     |

Get the files from the [release page](https://github.com/AzPepoze/light-launcher/releases).

## What you need

Some words used below:

- **Proton**: a compatibility layer from Valve. It runs Windows games on Linux.
- **umu-launcher** (`umu-run`): the tool that starts a Windows game with Proton on Linux.
- **WebKitGTK 6.0**: the web engine that draws the window on Linux.
- **WebView2**: Microsoft's web engine that draws the window on Windows.

| Item                     | Linux      | Windows    | Why                                                |
| :----------------------- | :--------- | :--------- | :------------------------------------------------- |
| umu-launcher (`umu-run`) | Required   | Not needed | Starts Windows games with Proton.                  |
| WebKitGTK 6.0 (GTK4)     | Required   | Not needed | Draws the launcher window.                         |
| WebView2 Runtime         | Not needed | Required   | Draws the launcher window. Windows 11 includes it. |
| 7-Zip                    | Optional   | Optional   | Unpacks `.7z` files. Must be on PATH.              |
| Steam                    | Optional   | Not used   | Lets you use Proton versions from Steam.           |
| ProtonPlus               | Optional   | Not used   | Helps you manage Proton versions.                  |

> [!TIP]
> Without 7-Zip, RAR files use a slower built-in reader. `.7z` files cannot be unpacked.

## Arch from source

The `PKGBUILD` file downloads the code from GitHub, builds it and installs it.

1. Install the tools it needs.

   ```sh
   sudo pacman -S --needed umu-launcher webkitgtk-6.0 git go bun gtk4
   ```

2. Build and install the package. `-s` installs missing build tools. `-i` installs the result.

   ```sh
   cd install/arch
   makepkg -si
   ```

## Arch release files

The `PKGBUILD-bin` file installs the release archive instead of building from source.

1. Open `install/arch/PKGBUILD-bin`.
2. Set `pkgver` to the release version.
3. Replace `SKIP` with the `linux-x64` checksum from the release's `SHA256SUMS` file. See [Check a download](releasing.md#check-a-download).
4. Build and install it.

   ```sh
   cd install/arch
   makepkg -si -p PKGBUILD-bin
   ```

## Tar archive

Use this on any Linux distribution.

1. Make a folder and extract the archive into it.

   ```sh
   mkdir -p ~/apps/light-launcher
   tar -xzf light-launcher-<version>-linux-x64.tar.gz -C ~/apps/light-launcher
   ```

2. Start the launcher.

   ```sh
   ~/apps/light-launcher/light-launcher
   ```

> [!NOTE]
> Keep `light-launcher`, `light-launcher-instance` and `light-launcher-shell` in the same folder. The launcher starts the other two from there.

The archive does not add a menu entry. See [Linux](linux.md) for window options and optional tools.

## Windows installer

1. Run `light-launcher-<version>-windows-x64-setup.exe`.
2. If the installer asks to install WebView2, choose **Yes**. It needs an internet connection.
3. Choose the folder. The default is `Program Files`.
4. Finish. Start LightLauncher from the Start Menu or the Desktop.

## Windows portable

1. Unzip `light-launcher-<version>-windows-x64.zip` into any folder.
2. Run `light-launcher.exe`.

Nothing is installed with this option. You still need WebView2.

> [!WARNING]
> The Windows window has not been run on a real Windows PC yet. Report any problem you find.

## Next steps

| Topic                                   | Page                      |
| :-------------------------------------- | :------------------------ |
| Flags and file locations                | [Usage](usage.md)         |
| Add sources such as F95 Zone or itch.io | [Plugins](plugins.md)     |
| Move games from older versions          | [Migration](migration.md) |
| Windows details                         | [Windows](windows.md)     |
| Linux details                           | [Linux](linux.md)         |
