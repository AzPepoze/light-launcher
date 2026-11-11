[Back to README](../README.md)

# Linux

This page explains how to install LightLauncher on Linux and what it needs to run.

## Install options

| Option            | Use it when                         | Command or file                             |
| :---------------- | :---------------------------------- | :------------------------------------------ |
| Arch, from source | You want to build it yourself.      | `install/arch/PKGBUILD`                     |
| Arch, prebuilt    | You want the release files on Arch. | `install/arch/PKGBUILD-bin`                 |
| Tar archive       | Any other distro.                   | `light-launcher-<version>-linux-x64.tar.gz` |

### Arch from source

```sh
cd install/arch
makepkg -si
```

### Arch prebuilt

Set `pkgver` and the checksum in `PKGBUILD-bin` to match the release. See [Releasing](releasing.md#check-a-download).

```sh
cd install/arch
makepkg -si -p PKGBUILD-bin
```

### Tar archive

1. Extract the archive to a folder.
2. Run the launcher from that folder.

```sh
mkdir -p ~/apps/light-launcher
tar -xzf light-launcher-<version>-linux-x64.tar.gz -C ~/apps/light-launcher
~/apps/light-launcher/light-launcher
```

> [!NOTE]
> Keep `light-launcher`, `light-launcher-instance` and `light-launcher-shell` in the same folder. The launcher starts the other two from there.

## What it needs

| Need                     | Why                               | Package (Arch)  |
| :----------------------- | :-------------------------------- | :-------------- |
| umu-launcher (`umu-run`) | Starts Windows games with Proton. | `umu-launcher`  |
| Qt6 WebEngine            | Shows the window.                 | `qt6-webengine` |

Optional:

| Tool                         | What it adds                                                            |
| :--------------------------- | :---------------------------------------------------------------------- |
| Steam                        | Steam Proton versions                                                   |
| ProtonPlus                   | Manage Proton versions                                                  |
| lsfg-vk (Lossless Scaling)   | Frame generation (LSFG)                                                 |
| MangoHud                     | Performance overlay                                                     |
| Gamescope                    | Micro compositor for scaling                                            |
| GameMode                     | Game performance tuning                                                 |
| 7-Zip (`7zz`, `7z` or `7za`) | Unpacks `.7z` files. RAR files use a slower built-in reader without it. |

## Window options

| Setting                  | How                                                                         |
| :----------------------- | :-------------------------------------------------------------------------- |
| Show on Wayland          | Qt picks Wayland or X11 by itself.                                          |
| Force X11                | `QT_QPA_PLATFORM=xcb light-launcher`                                        |
| Window position and size | Saved in `~/.config/light-launcher/window-geometry` when the window closes. |

## Why there is no AppImage

An AppImage would need to contain Qt WebEngine. That adds a large set of Chromium libraries, so the file would be much bigger than the other Linux files.

## More

| Topic                 | Page                         |
| :-------------------- | :--------------------------- |
| Build the Linux files | [Building](building.md)      |
| Windows               | [Windows](windows.md)        |
| Shell options         | [Shells](../shell/README.md) |
| Publishing            | [Releasing](releasing.md)    |
