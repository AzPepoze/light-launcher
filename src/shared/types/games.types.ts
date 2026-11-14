import type {
	WireGameActivity,
	WireLauncher,
	WireLauncherKind,
	WireRunningSession
} from "../contract/wire.generated";
import type { LaunchOptions } from "./config.types";

export type LauncherKind = WireLauncherKind;

/** One executable that starts a game folder. */
export type Launcher = WireLauncher;

// WireGameInfo allows a null launchers list; the renderer always keeps a list.
export interface GameInfo {
	name: string;
	path: string;
	icon: string;
	config: LaunchOptions;
	isRecent: boolean;
	isAutoScanned: boolean;
	folder: string;
	launchers: Launcher[];
	launcherCount: number;
}

// WireScannedFolderGroup allows a null games list; the renderer always keeps a list.
export interface ScannedFolderGroup {
	folderPath: string;
	folderName: string;
	games: GameInfo[];
}

export type RunningSession = WireRunningSession;

export type GameActivity = WireGameActivity;
