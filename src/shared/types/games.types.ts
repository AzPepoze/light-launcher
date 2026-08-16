import type { LaunchOptions } from "./config.types";

export interface GameInfo {
	name: string;
	path: string;
	icon: string;
	config: LaunchOptions;
	isRecent: boolean;
	isAutoScanned: boolean;
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
}
