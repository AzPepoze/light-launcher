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
	/** All detected GPUs (primary first). `gpu` holds gpus[0] for backward compat. */
	gpus?: string[];
	ram: string;
	driver: string;
}

export interface SystemUsage {
	cpu: string;
	ram: string;
	gpu: string;
	gpus?: string[];
}

export interface UtilsStatus {
	isLsfgInstalled: boolean;
	lsfgVersion: string;
}
