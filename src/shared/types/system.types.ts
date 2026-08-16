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
	ram: string;
	driver: string;
}

export interface SystemUsage {
	cpu: string;
	ram: string;
	gpu: string;
}

export interface UtilsStatus {
	isLsfgInstalled: boolean;
	lsfgVersion: string;
}
