<h1 align="center">
  <!-- <img src="docs" alt="Logo" width="128" height="128" style="filter: brightness(0) invert(1);"/><br> -->
  ✦ LightLauncher ✦
</h1>

<p align="center">
  <strong>◈ Proton Instance Manager for Linux ◈</strong>
  <br>
  <strong>◈ Powered by Go & umu-run ◈</strong>
</p>

<p align="center">
  <!-- <a href="https://github.com/AzPepoze/light-launcher/releases/latest">
    <img src="https://img.shields.io/github/v/release/AzPepoze/light-launcher?style=for-the-badge&label=%E2%97%88%20RELEASE%20%E2%97%88&labelColor=%23181818&color=%23ffffff" alt="Latest Release">
  </a> -->
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/AzPepoze/light-launcher?style=for-the-badge&label=%E2%97%88%20LICENSE%20%E2%97%88&labelColor=%23181818&color=%23ffffff" alt="License">
  </a>
  <a href="https://github.com/AzPepoze/light-launcher/stargazers">
    <img src="https://img.shields.io/github/stars/AzPepoze/light-launcher?style=for-the-badge&label=%E2%97%88%20STARS%20%E2%97%88&labelColor=%23181818&color=%23ffffff" alt="Stars">
  </a>
</p>

LightLauncher is a games launcher designed to run non-Steam games on Linux using **umu** (Unified Linux Runtime). It provides a seamless way to manage and execute Windows applications through Proton without needing the Steam client to be active for each session.

> [!WARNING]
>
> This project is still in **early development**. You may encounter bugs or breaking changes. Feel free to report issues or contribute!

> [!NOTE]
>
> The name **LightLauncher** comes from the word **"Go"** (as in "Go for it!"), representing the ability to launch Proton immediately. It's also a happy coincidence that the project is built using the **Go (Golang)** programming language! XD

## CONTENTS

- [CONTENTS](#contents)
- [SCREENSHOTS](#screenshots)
- [FEATURES](#features)
- [ARCHITECTURE \& EFFICIENCY](#architecture--efficiency)
- [PREREQUISITES](#prerequisites)
- [INSTALLATION](#installation)
  - [Arch Linux](#arch-linux)
  - [Other Distributions](#other-distributions)
- [BUILD](#build)
- [USAGE](#usage)
- [STONKS!](#stonks)

## SCREENSHOTS

|            HOME            |
| :------------------------: |
|    ![home](docs/home.png)    |
|         ADD GAMES         |
| ![home](docs/add_games.png) |
|            RUN            |
|     ![run](docs/run.png)     |
|           PREFIX           |
|  ![prefix](docs/prefix.png)  |
|          VERSION          |
| ![version](docs/version.png) |
|           UTILS           |
|   ![utils](docs/utils.png)   |

## FEATURES

- **Independent Instance Manager** – Each game runs as a standalone detached process. Closing the main launcher does not affect running games.
- **Multi-Game Support** – Run multiple Windows applications simultaneously with unique Proton configurations and prefixes.
- **Process Isolation** – Every game gets its own System Tray icon for individual management (Graceful Stop/Status).
- **Native Terminal Integration** – Real-time logs are piped to your preferred terminal (Kitty, Alacritty, etc.) for live debugging.
- **Automatic Log Management** – Persistent logging to `~/LightLauncher/logs` with automatic rotation (keeps last 10 sessions)
- **umu-run Core** – Utilizes the Unified Linux Runtime (umu) to provide superior execution for non-Steam games.

## ARCHITECTURE & EFFICIENCY

1. **UI (Wails/Go):** Used only for configuration. It **closes completely** after launching a game, freeing all WebKit memory.
2. **Instance Manager (`light-launcher-instance`):** A tiny Go binary that manages the process life-cycle and tray.

## PREREQUISITES

- **umu-launcher** (Required for execution)
- **Steam** (Installed and configured)
- **ProtonPlus** (Recommended for managing and adding Proton versions)

> [!TIP]
>
> Use **ProtonPlus** or **Steam** to download and install different Proton versions. LightLauncher will automatically detect them in your Steam compatibility tools directory.

## INSTALLATION

### Arch Linux

```bash
# Download and install using makepkg
mkdir -p /tmp/light-launcher && cd /tmp/light-launcher
curl -L -O https://raw.githubusercontent.com/AzPepoze/light-launcher/main/install/arch/PKGBUILD
curl -L -O https://raw.githubusercontent.com/AzPepoze/light-launcher/main/install/arch/light-launcher.install
makepkg -si
cd .. && sudo rm -rf light-launcher
```

### Other Distributions

Build from source:

```bash
git clone https://github.com/AzPepoze/light-launcher.git
cd light-launcher
make build
# Binaries are in ./bin/
```

## BUILD

Build from source locally:

```bash
# Install dependencies: go, nodejs, wails
make build

# Binaries will be in ./bin/
```

## USAGE

**Arch Linux** (after `makepkg -si`):

```bash
light-launcher                    # Launch UI
light-launcher path/to/game.exe   # Direct launch
```

**Other Distributions** (from ./bin/):

```bash
./bin/light-launcher                    # Launch UI
./bin/light-launcher path/to/game.exe   # Direct launch
```

## STONKS!

<div align="center">
  <a href="https://www.star-history.com/#AzPepoze/light-launcher&type=date&legend=top-left">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=AzPepoze/light-launcher&type=date&theme=dark&legend=top-left" />
      <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=AzPepoze/light-launcher&type=date&legend=top-left" />
      <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=AzPepoze/light-launcher&type=date&legend=top-left" width="600" />
    </picture>
  </a>
  <br>
  <br>
  <strong>✦ Made with ♥︎ by AzPepoze ✦</strong>
</div>
