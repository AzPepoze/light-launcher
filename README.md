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

## CONTENTS

- [CONTENTS](#contents)
- [SCREENSHOTS](#screenshots)
- [FEATURES](#features)
- [ARCHITECTURE \& EFFICIENCY](#architecture--efficiency)
- [PREREQUISITES](#prerequisites)
- [QUICK START](#quick-start)
- [USAGE](#usage)
- [CLI REFERENCE](#cli-reference)
  - [Main Launcher CLI](#1-main-launcher-cli-light-launcher)
  - [Instance Manager CLI](#2-instance-manager-cli-light-launcher-instance)
- [WAYLAND \& DISPLAY PLATFORM (OZONE)](#wayland--display-platform-ozone)
- [CONTRIBUTING](#contributing)
- [STONKS!](#stonks)

## SCREENSHOTS

|             HOME             |
| :--------------------------: |
|    ![home](docs/home.png)    |
|          ADD GAMES           |
| ![home](docs/add_games.png)  |
|             RUN              |
|     ![run](docs/run.png)     |
|            PREFIX            |
|  ![prefix](docs/prefix.png)  |
|           VERSION            |
| ![version](docs/version.png) |
|            UTILS             |
|   ![utils](docs/utils.png)   |

## FEATURES

| Feature                | Description                                                                      |
| :--------------------- | :------------------------------------------------------------------------------- |
| **Detached Instances** | Games run independently. Closing launcher does not close games.                  |
| **Multi-Game Support** | Run multiple games simultaneously with custom Proton versions and prefixes.      |
| **Tray Management**    | Dedicated tray icon per game for status monitoring and graceful exit.            |
| **Live Terminal Logs** | Pipes real-time logs to terminal (Kitty, Alacritty, etc.) for instant debugging. |
| **Log Rotation**       | Auto-rotates logs in `~/.config/light-launcher/logs` (keeps last 10 runs).       |
| **umu-run Core**       | Uses Unified Linux Runtime (`umu`) for Proton compatibility.                     |
| **Flexible Prefixes**  | Supports isolated or shared WINE prefixes across Proton builds.                  |
| **Ozone Support**      | Runs natively on Wayland or X11/Xwayland via Ozone.                              |

## ARCHITECTURE & EFFICIENCY

1. **UI (Electron / Svelte 5):** Interface for configuration, scanning, and monitoring.
2. **Instance Manager (`light-launcher-instance`):** Standalone Go binary supervising game lifecycle and tray.

## PREREQUISITES

- [**umu-launcher**](https://github.com/Open-Wine-Components/umu-launcher) (Required for execution)
- [**Steam**](https://store.steampowered.com/about/) (Installed and configured)
- [**ProtonPlus**](https://github.com/Vysp3r/ProtonPlus) (Recommended for managing Proton versions)

> [!TIP]
>
> Use **ProtonPlus** or **Steam** to install Proton versions. LightLauncher detects them automatically.

## QUICK START

Get up and running in 3 steps:

### 1. Install `umu-launcher`

```bash
# Arch Linux
sudo pacman -S umu-launcher # or yay -S umu-launcher
```

### 2. Install or Build LightLauncher

**Option A: Arch Linux (PKGBUILD)**

```bash
mkdir -p /tmp/light-launcher && cd /tmp/light-launcher
curl -L -O https://raw.githubusercontent.com/AzPepoze/light-launcher/main/install/arch/PKGBUILD
makepkg -si
cd .. && rm -rf light-launcher
```

**Option B: Build from source (All Distributions)**

```bash
git clone https://github.com/AzPepoze/light-launcher.git
cd light-launcher
bun install && bun run build
```

### 3. Launch & Play

```bash
# Start launcher
light-launcher        # System install
bun run dev           # Local development

# Or launch a game directly
light-launcher /path/to/game.exe                # System install
light-launcher-instance -game /path/to/game.exe # Direct instance runner
```

## USAGE

```bash
# Launch Graphical Launcher
light-launcher [OPTIONS] [GAME_PATH]

# Launch Instance Daemon Directly
light-launcher-instance [OPTIONS] -game <PATH>
```

---

## CLI REFERENCE

### 1. Main Launcher CLI (`light-launcher`)

```bash
light-launcher [OPTIONS] [GAME_PATH]
```

#### [OPTIONS]

| Option / Flag                 | Type     | Default | Description                                   |
| :---------------------------- | :------- | :------ | :-------------------------------------------- |
| `[game_path]`                 | `string` | —       | Path to game `.exe` to launch or configure    |
| `--wayland`                   | `flag`   | —       | Force Native Wayland display via Ozone        |
| `--x11`                       | `flag`   | —       | Force X11 / Xwayland display via Ozone        |
| `--ozone-platform=<platform>` | `string` | `x11`   | Set Ozone platform (`wayland`, `x11`, `auto`) |
| `--edit-lsfg`                 | `flag`   | —       | Open directly into LSFG configuration view    |

#### Examples

```bash
# Launch default UI
light-launcher

# Launch in Native Wayland mode
light-launcher --wayland

# Launch or configure a game
light-launcher ~/Games/EldenRing/eldenring.exe

# Open directly into LSFG configuration
light-launcher --edit-lsfg
```

---

### 2. Instance Manager CLI (`light-launcher-instance`)

```bash
light-launcher-instance [OPTIONS] -game <GAME_PATH>
```

#### [OPTIONS] — Core Launch Settings

| Flag                     | Type     | Default   | Description                         |
| :----------------------- | :------- | :-------- | :---------------------------------- |
| `-game <path>`           | `string` | —         | **(Required)** Game executable path |
| `-launcher <path>`       | `string` | —         | Wrapper launcher executable         |
| `-prefix <path>`         | `string` | `Default` | `WINEPREFIX` directory              |
| `-proton-path <path>`    | `string` | —         | Proton tool directory               |
| `-proton-pattern <name>` | `string` | —         | Proton pattern for UMU lookup       |
| `-logs`                  | `bool`   | `true`    | Open live terminal log window       |

#### [OPTIONS] — MangoHud & GameMode

| Flag                  | Description             |
| :-------------------- | :---------------------- |
| `-mangohud`, `-mango` | Enable MangoHud overlay |
| `-gamemode`           | Enable Feral GameMode   |

#### [OPTIONS] — Memory Protection

| Flag                      | Type     | Default | Description                            |
| :------------------------ | :------- | :------ | :------------------------------------- |
| `-memory-min`             | `flag`   | —       | Enable RAM trimming & protection       |
| `-memory-min-value <val>` | `string` | `4G`    | Protection threshold (e.g. `4G`, `8G`) |

#### [OPTIONS] — Gamescope

| Flag                     | Type     | Default      | Description                                 |
| :----------------------- | :------- | :----------- | :------------------------------------------ |
| `-gamescope`             | `flag`   | —            | Enable Gamescope compositor                 |
| `-gs-w <width>`          | `string` | `1280`       | Internal render width                       |
| `-gs-h <height>`         | `string` | `720`        | Internal render height                      |
| `-gs-out-w <width>`      | `string` | —            | Output display width                        |
| `-gs-out-h <height>`     | `string` | —            | Output display height                       |
| `-gs-r <rate>`           | `string` | `60`         | Refresh rate limit (`60`, `144`, etc.)      |
| `-gs-fr-limit <fps>`     | `string` | —            | Frame rate cap                              |
| `-gs-window-mode <mode>` | `string` | `borderless` | `fullscreen`, `borderless`, `windowed`      |
| `-gs-scaler <scaler>`    | `string` | `auto`       | `auto`, `integer`, `fit`, `fill`, `stretch` |
| `-gs-filter <filter>`    | `string` | `linear`     | `linear`, `nearest`, `fsr`, `nis`, `pixel`  |
| `-gs-sharpness <level>`  | `string` | `0`          | Upscaling sharpness (`0` - `20`)            |
| `-gs-hdr`                | `flag`   | —            | Enable HDR output                           |
| `-gs-adaptive-sync`      | `flag`   | —            | Enable VRR / Adaptive Sync                  |
| `-gs-mangoapp`           | `flag`   | —            | Enable Mangoapp overlay in Gamescope        |
| `-gs-custom-args <args>` | `string` | —            | Additional Gamescope flags                  |

#### [OPTIONS] — Lossless Scaling (LSFG-VK)

| Flag                      | Type     | Default  | Description                        |
| :------------------------ | :------- | :------- | :--------------------------------- |
| `-lsfg`                   | `flag`   | —        | Enable LSFG frame generation layer |
| `-lsfg-multiplier <mult>` | `string` | `2`      | Frame multiplier (`2`, `3`, `4`)   |
| `-lsfg-perf`              | `flag`   | —        | Enable Performance Mode            |
| `-lsfg-dll <path>`        | `string` | —        | Path to `Lossless.dll`             |
| `-lsfg-gpu <index>`       | `string` | —        | Target GPU index                   |
| `-lsfg-flow <scale>`      | `string` | `1.0`    | Flow scale factor                  |
| `-lsfg-pacing <mode>`     | `string` | `smooth` | Pacing mode (`smooth`, etc.)       |
| `-lsfg-fp16`              | `flag`   | —        | Enable FP16 half-precision         |

#### Examples

```bash
# 1. Basic Launch with custom prefix and Proton version
light-launcher-instance \
  -game ~/Games/Cyberpunk2077/bin/x64/Cyberpunk2077.exe \
  -prefix ~/.config/light-launcher/prefixes/Cyberpunk \
  -proton-path ~/.local/share/Steam/compatibilitytools.d/GE-Proton9-25

# 2. Launch with MangoHud & GameMode
light-launcher-instance \
  -game ~/Games/EldenRing/eldenring.exe \
  -mangohud \
  -gamemode

# 3. Launch with Gamescope (720p internal -> 1080p output with FSR)
light-launcher-instance \
  -game ~/Games/Witcher3/bin/x64/witcher3.exe \
  -gamescope \
  -gs-w 1280 -gs-h 720 \
  -gs-out-w 1920 -gs-out-h 1080 \
  -gs-filter fsr -gs-sharpness 5 \
  -gs-window-mode borderless

# 4. Launch with Lossless Scaling (2x multiplier)
light-launcher-instance \
  -game ~/Games/Cyberpunk2077/bin/x64/Cyberpunk2077.exe \
  -lsfg \
  -lsfg-multiplier 2 \
  -lsfg-flow 1.0 \
  -lsfg-pacing smooth

# 5. Combined Advanced Launch (Gamescope + LSFG + MangoHud + GameMode)
light-launcher-instance \
  -game ~/Games/Game.exe \
  -prefix ~/.config/light-launcher/prefixes/MyGame \
  -proton-path ~/.local/share/Steam/compatibilitytools.d/GE-Proton9-25 \
  -gamemode \
  -mangohud \
  -gamescope -gs-w 1920 -gs-h 1080 -gs-r 144 \
  -lsfg -lsfg-multiplier 2
```

---

## WAYLAND & DISPLAY PLATFORM (OZONE)

LightLauncher supports running natively under Wayland or X11/Xwayland via Chromium's Ozone abstraction layer:

1. **In-App Settings**: Navigate to **Appearance & Settings -> Display & Window Platform** to toggle between Native Wayland and X11 mode (automatically restarts the app to apply).
2. **CLI Flags**: Pass `--wayland` or `--x11` when starting the launcher to override the configured platform for that session.
3. **Flags Configuration File**: Create `~/.config/light-launcher-flags.conf` to pass custom Electron/Chromium flags (matching the Arch / VS Code convention):
   ```text
   # ~/.config/light-launcher-flags.conf
   --ozone-platform=wayland
   --enable-features=UseOzonePlatform,WaylandWindowDecorations
   ```

## CONTRIBUTING

Contributions, bug reports, and suggestions are welcome! Please check out [CONTRIBUTING.md](CONTRIBUTING.md) for local development setup, architecture details, available scripts, and contribution guidelines.

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
