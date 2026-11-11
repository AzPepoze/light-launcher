[Back to README](../README.md)

# Game folders

LightLauncher looks in the folders you watch and finds games there. Each game folder becomes one card on Home.

## One card per game folder

| Step   | What happens                                                                                    |
| :----- | :---------------------------------------------------------------------------------------------- |
| Find   | Every `.exe` file under a watched folder, down to the depth you set.                            |
| Group  | Files in the same game folder make one card. See [Categories](#categories).                     |
| Drop   | Helper programs are removed. See [Helper programs](#helper-programs).                           |
| Pick   | One file is the main launcher. The card starts that file. See [Main launcher](#main-launcher). |

```text
Games/                       <- watched folder
├── MyGame/                  <- one card: MyGame
│   ├── MyGame.exe
│   └── Binaries/Win64/MyGame-Win64-Shipping.exe
└── Solo.exe                 <- one card: Solo
```

> [!NOTE]
> An `.exe` file directly inside a watched folder is its own game. It makes one card.

## Categories

A folder that holds several game folders is a category, not a game. The games inside it get their own cards.

| Folder                                          | Rule                                                  | Card          |
| :---------------------------------------------- | :---------------------------------------------------- | :------------ |
| No `.exe` of its own, 3 or more game folders    | Category                                              | None          |
| Has an `.exe` of its own                        | Game                                                  | One card      |
| Has fewer than 3 sub-folders with `.exe` files  | Game, even if it has several `.exe` files inside      | One card      |

```text
Games/                       <- watched folder
├── Action/                  <- category: no exe, 3 game folders
│   ├── Alpha/               <- one card: Alpha
│   │   └── Alpha.exe
│   ├── Beta/                <- one card: Beta
│   │   └── Beta.exe
│   └── Gamma/               <- one card: Gamma
│       └── Gamma.exe
└── Puzzle/                  <- category
    ├── Delta/               <- one card: Delta
    │   └── Delta.exe
    └── Epsilon/             <- one card: Epsilon
        └── Epsilon.exe
```

> [!WARNING]
> A game folder with no `.exe` of its own and 3 or more sub-folders with `.exe` files looks like a category. Put a launcher `.exe` directly in that folder to keep it as one game.

## Main launcher

The launcher rules run in order. The first rule that decides wins.

| Order | Rule                                                                                  | Example                            |
| :---- | :------------------------------------------------------------------------------------ | :--------------------------------- |
| 1     | The launcher you chose before, if the file still exists.                              | You chose `Other.exe`.             |
| 2     | Unreal Engine shipping build: `<Game>-Win64-Shipping.exe` inside `Binaries/Win64`.   | `MyGame-Win64-Shipping.exe`      |
| 3     | A 32-bit name loses. A name with `32`, `x86` or `Win32` counts as 32-bit.            | `Game.exe` beats `Game32.exe`      |
| 4     | The name is closest to the folder name.                                               | `Game.exe` in the `Game` folder    |
| 5     | The bigger file wins.                                                                 |                                    |
| 6     | The shorter path wins, then the alphabetical order.                                   |                                    |

## Card names

A card takes the first name it finds. The order is:

| Order | Name                                                                                  | Example                              |
| :---- | :------------------------------------------------------------------------------------ | :----------------------------------- |
| 1     | A name saved in the game's config (the `Name` field).                                 | `My Game`                            |
| 2     | The product name inside the `.exe`. The file description is the fallback.            | `Hollow Knight`                      |
| 3     | The folder name, cleaned. Used when the launcher name is generic or the folder has several launchers. | `Mad_Island_0.34` becomes `Mad Island` |
| 4     | The launcher file name without `.exe`.                                                | `Hollow.exe` becomes `Hollow`        |

Generic names say nothing about the game. They are skipped in steps 2 and 3:

- Engine and tool names, such as `Unity`, `Unreal Engine`, `Ren'Py`, `Godot Engine`, `NW.js` and `Electron`.
- Windows names, such as `Microsoft(R) Windows(R) Operating System`.
- Launcher names, such as `game`, `launcher`, `start`, `play`, `run`, `main`, `app`, `client`, `win`, `pc`, `nw` and `UnityPlayer`.
- Unreal and 32 or 64-bit builds, such as `<folder>-32`, `<folder>-64` and `*-Win64-Shipping`.

Folder names are cleaned like this:

| Rule                                                                  | Example                                  |
| :-------------------------------------------------------------------- | :--------------------------------------- |
| `_` and `.` become spaces, when the name has no spaces yet.          | `Hollow_Knight` becomes `Hollow Knight`  |
| Version endings are removed.                                          | `Game-v1.2.3` and `Game [v0.5]` become `Game` |
| Platform endings are removed.                                         | `Game_Win`, `Game-pc` and `Game-win` become `Game` |

> [!NOTE]
> If cleaning would leave nothing, the folder name is used as it is.

## Other launchers

The other `.exe` files of a game folder stay on the card as alternatives. Each one has a kind:

| Kind            | Meaning                                                              |
| :-------------- | :------------------------------------------------------------------- |
| Main game       | The launcher the card starts.                                        |
| Unreal launcher | The starter file at the top of an Unreal game folder.                |
| 32-bit          | A 32-bit file that lost the rules above.                             |
| Other           | Any other `.exe` file in the folder.                                 |

> [!IMPORTANT]
> Each launcher keeps its own settings. Starting another launcher uses that launcher's settings. The settings are never merged.

## Helper programs

Helper programs are never shown as launchers. The names below are removed from each game folder. Names match in any letter case.

| Kind                  | Examples                                                |
| :-------------------- | :------------------------------------------------------ |
| Crash handlers        | `UnityCrashHandler64.exe`, `CrashReportClient.exe`      |
| Setup and installers  | `setup.exe`, `DXSETUP.exe`, `installer*.exe`, `unins*.exe`, `*uninstall*.exe` |
| Redistributables      | `vc_redist*.exe`, `dotNetFx*.exe`                       |
| Notification helpers  | `*notification*.exe`                                    |
| Browser sub-processes | `*subprocess*.exe`                                      |

> [!NOTE]
> This list is built in. The excluded names in the scan folder settings are a different list and are not changed.

## Saved choices

When you choose a launcher, the choice is saved in `launcher-preferences.json` in the LightLauncher folder. See [File locations](usage.md#file-locations).

> [!WARNING]
> If that file cannot be read, it is copied to `launcher-preferences.json.bak`. The next save starts a new file, so the choices start empty. Keep the `.bak` file if you want the old choices back.

## Games added by hand

A game you add by hand has one launcher. Its card is not grouped with other files.

## More

| Topic | Page                  |
| :---- | :-------------------- |
| Usage | [Usage](usage.md)     |
| Install | [Install](install.md) |
