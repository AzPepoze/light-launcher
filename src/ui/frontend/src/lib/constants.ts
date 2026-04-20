import * as core from "@bindings/light-launcher/pkg/core/models";

export const DEFAULT_LAUNCH_OPTIONS: core.LaunchOptions = {
	...({
		MainExecutablePath: "",
		LauncherPath: "",
		HaveGameExe: false,
		PrefixPath: "",
		ProtonPattern: "",
		ProtonPath: "",
		CustomArgs: "",
		EnableGamescope: false,
		GamescopeW: "1920",
		GamescopeH: "1080",
		GamescopeR: "60",
		EnableMangoHud: false,
		EnableGamemode: false,
		EnableLsfgVk: false,
		LsfgMultiplier: "2",
		LsfgPerfMode: false,
		LsfgDllPath: "",
		LsfgGpu: "",
		LsfgFlowScale: "0.8",
		LsfgPacing: "none",
		LsfgAllowFp16: false,
		EnableMemoryMin: false,
		MemoryMinValue: "4G",
	} as core.LaunchOptions),
};

export const LSFG_DEFAULT_OPTIONS: Partial<core.LaunchOptions> = {
	LsfgMultiplier: "2",
	LsfgPerfMode: false,
	LsfgDllPath: "",
	LsfgGpu: "",
	LsfgFlowScale: "0.8",
	LsfgPacing: "none",
	LsfgAllowFp16: false,
};

export const GAMESCOPE_DEFAULTS = {
	width: "1920",
	height: "1080",
	refreshRate: "60",
};

export const MEMORY_DEFAULTS = {
	value: "4G",
	min: 512,
	max: null as number | null, // Set at runtime based on system RAM
};
