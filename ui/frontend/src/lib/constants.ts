import * as core from "@bindings/light-launcher/internal/types/models";

export const DEFAULT_LAUNCH_OPTIONS: core.LaunchOptions = {
	...({
		Name: "Launcher",
		RunnerPath: "",
		GamePath: "",
		UseGamePath: false,
		PrefixPath: "",
		ProtonPattern: "",
		ProtonPath: "",
		CustomArgs: "",
		Extras: {
			EnableMangoHud: false,
			EnableGamemode: false,
			Lsfg: {
				Enabled: false,
				Multiplier: "2",
				PerfMode: false,
				DllPath: "",
				Gpu: "",
				FlowScale: "0.8",
				Pacing: "none",
				AllowFp16: false,
			},
			Gamescope: {
				Enabled: false,
				Width: "1920",
				Height: "1080",
				RefreshRate: "60",
			},
			Memory: {
				Enabled: false,
				Value: "4G",
			},
		},
	} as core.LaunchOptions),
};

export const LSFG_DEFAULT_OPTIONS = {
	Enabled: false,
	Multiplier: "2",
	PerfMode: false,
	DllPath: "",
	Gpu: "",
	FlowScale: "0.8",
	Pacing: "none",
	AllowFp16: false,
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
