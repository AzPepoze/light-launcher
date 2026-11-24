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

| Need                     | Why                               | Package (Arch)   |
| :----------------------- | :-------------------------------- | :--------------- |
| umu-launcher (`umu-run`) | Starts Windows games with Proton. | `umu-launcher`   |
| WebKitGTK 6.0 (GTK4)     | Shows the window.                 | `webkitgtk-6.0`  |

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

## Game drives

A game on a separate disk, such as an NTFS partition, needs a drive letter inside the Wine prefix.

| Step  | What happens                                                                                  |
| :---- | :-------------------------------------------------------------------------------------------- |
| Start | Before the game starts, LightLauncher finds the disk that holds the game.                     |
| Map   | If the disk has no letter yet, it links the disk to the first free letter from D: to Y:.      |
| Keep  | An existing letter is never changed, so a letter you set yourself still wins.                 |

> [!NOTE]
> Disks on the root filesystem are skipped. They are already visible as Z:.

### Drive setup

Set up drives once, and they apply to every prefix.

1. Open **Drive setup**.
2. Under **Detected disks**, click **Add as drive** on a disk.
3. Or, under **Add a folder**, pick a folder and a letter.
4. Click **Save**.

Each prefix gets the saved drives when a game starts.

| Letter | Use                                                                                    |
| :----- | :------------------------------------------------------------------------------------- |
| C:     | Replaces the prefix's own C: drive. Remove it to go back to Wine's default.            |
| D: to Y: | Any folder or disk.                                                                   |
| Z:     | Not allowed. It is the Linux root.                                                     |

> [!WARNING]
> C: replaces the prefix's own C: drive. Games then see the chosen disk as C:.

The **Drives** tab on a prefix page shows which drives are active in that prefix. It is read-only. Change drives in **Drive setup**.

## Window options

| Setting                  | How                                                    |
| :----------------------- | :----------------------------------------------------- |
| Show on Wayland          | GTK picks Wayland or X11 by itself.                    |
| Force X11                | `GDK_BACKEND=x11 light-launcher`                       |
| Window position and size | Not saved. The GTK window opens at its default size.   |

## Why there is no AppImage

An AppImage would need to contain WebKitGTK and GTK4. That adds a large set of libraries, so the file would be much bigger than the other Linux files.

## More

| Topic                 | Page                         |
| :-------------------- | :--------------------------- |
| Build the Linux files | [Building](building.md)      |
| Windows               | [Windows](windows.md)        |
| Shell options         | [Shells](../shell/README.md) |
| Publishing            | [Releasing](releasing.md)    |
