export interface LsfgProfileData {
	name: string;
	multiplier: number;
	performanceMode: boolean;
	gpu: string;
	flowScale: number;
	pacing: string;
	dllPath: string;
	allowFp16: boolean;
}

export interface LsfgInternalProfile {
	Name: string;
	Multiplier: number;
	PerformanceMode: boolean;
	GPU: string;
	FlowScale: number;
	Pacing: string;
	AllowFP16: boolean;
	Disable?: boolean;
}
