[Back to README](../README.md)

# Windows

This page explains what LightLauncher does on Windows and how to install it.

> [!WARNING]
> The Windows window (WebView2 shell) has not been run on a Windows PC yet. Report any problem you find.

## What works on Windows

| Feature | Windows | Linux only |
| :--- | :---: | :---: |
| Add a game and start its `.exe` | Yes | |
| Browse the game library and settings | Yes | |
| Start a game with Proton and umu-run | | Yes |
| Prefixes | | Yes |
| Lossless Scaling (LSFG) | | Yes |
| Gamescope, MangoHud, GameMode | | Yes |
| Steam Proton versions | | Yes |

Features marked Linux only are hidden in the Windows app.

## Install

Pick one of the two files from the [release page](https://github.com/AzPepoze/light-launcher/releases).

| File | Use it when |
| :--- | :--- |
| `light-launcher-<version>-windows-x64-setup.exe` | You want a normal install with Start Menu and Desktop shortcuts. |
| `light-launcher-<version>-windows-x64.zip` | You want to run it from any folder. Nothing is installed. |

### Installer steps

1. Run the `-setup.exe` file.
2. If the installer asks about WebView2, choose **Yes**. It needs an internet connection.
3. Choose the folder. The default is `Program Files`.
4. Finish. Start LightLauncher from the Start Menu or the Desktop.

### Portable steps

1. Unzip the file to any folder.
2. Run `light-launcher.exe`.

> [!NOTE]
> The launcher, the instance and the shell must stay in the same folder.

## WebView2 runtime

LightLauncher shows its window with the Microsoft Edge WebView2 Runtime.

| Situation | What you see |
| :--- | :--- |
| The runtime is installed | The window opens. |
| The runtime is missing | A message with the download page. Choose **Yes** to open it in your browser. |

Download page: `https://developer.microsoft.com/microsoft-edge/webview2/`

## Warnings from Windows

The files are not signed yet. Windows may show a warning like **Windows protected your PC**.

1. Click **More info**.
2. Click **Run anyway**.

Check the file first. See the checksum steps in [Releasing](releasing.md#check-a-download).

## Uninstall

1. Open **Settings > Apps > Installed apps**.
2. Find **LightLauncher** and choose **Uninstall**.
3. The uninstaller asks if you also want to delete your data.

| Data | Where it is | Deleted when you answer Yes |
| :--- | :--- | :--- |
| Game settings, prefixes and logs | `%APPDATA%\light-launcher` | Yes |

> [!IMPORTANT]
> The data folder belongs to the Windows account that runs the uninstaller. Run it with the same account that you used for LightLauncher.

## More

| Topic | Page |
| :--- | :--- |
| Build the Windows files | [Building](building.md) |
| Linux | [Linux](linux.md) |
| Shell options | [Shells](../shell/README.md) |
