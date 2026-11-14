import type {
	WireSystemInfo,
	WireSystemToolsStatus,
	WireUtilsStatus
} from "../contract/wire.generated";

export type SystemToolsStatus = WireSystemToolsStatus;

export type SystemInfo = WireSystemInfo;

// WireSystemUsage allows a null gpus list; the renderer always keeps a list.
export interface SystemUsage {
	cpu: number;
	ram: number;
	ramUsedGb: number;
	ramTotalGb: number;
	gpus: number[];
}

export type UtilsStatus = WireUtilsStatus;

/** Launch tools and UI parts the current OS supports. Mirrors the Go platform.Features. */
export interface PlatformFeatures {
	proton: boolean;
	prefix: boolean;
	lsfg: boolean;
	gamescope: boolean;
	mangoHud: boolean;
	gameMode: boolean;
	liveTerminalLogs: boolean;
}

export interface PlatformInfo {
	os: string;
	features: PlatformFeatures;
}
