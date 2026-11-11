import type { LaunchOptions } from "./config.types";

export type LauncherKind = "main" | "unreal-stub" | "x86" | "other";

/** One executable that starts a game folder. */
export interface Launcher {
	name: string;
	path: string;
	kind: LauncherKind;
	primary: boolean;
}

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

export interface ScannedFolderGroup {
	folderPath: string;
	folderName: string;
	games: GameInfo[];
}

export interface RunningSession {
	pid: number;
	gamePath: string;
	gameName: string;
	startedAt?: number;
}

export interface GameActivity {
	gamePath: string;
	gameName: string;
	profileId?: string;
	customIconPath?: string;
	lastPlayedAt: number;
	totalPlaytimeSeconds: number;
	activeSince?: number;
}
