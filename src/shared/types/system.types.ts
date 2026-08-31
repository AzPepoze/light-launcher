export interface SystemToolsStatus {
	hasGamescope: boolean;
	hasMangoHud: boolean;
	hasGameMode: boolean;
	hasVulkanInfo: boolean;
}

export interface SystemInfo {
	os: string;
	kernel: string;
	cpu: string;
	gpu: string;
	gpus?: string[];
	ram: string;
	driver: string;
}

export interface SystemUsage {
	cpu: number;
	ram: number;
	ramUsedGb: number;
	ramTotalGb: number;
	gpus: number[];
}

export interface UtilsStatus {
	isLsfgInstalled: boolean;
	lsfgVersion: string;
}
