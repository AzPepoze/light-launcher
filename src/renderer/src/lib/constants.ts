import * as core from "@shared";

export const DEFAULT_LAUNCH_OPTIONS: core.LaunchOptions = {
	...({
		Name: "Launcher",
		LauncherPath: "",
		GamePath: "",
		UseGamePath: false,
		PrefixPath: "",
		ProtonPath: "",
		CustomArgs: "",
		Extras: {
			EnableMangoHud: false,
			EnableGamemode: false,
			Lsfg: {
				Enabled: false,
				Multiplier: core.DefaultMultiplier,
				PerfMode: false,
				DllPath: "",
				Gpu: "",
				FlowScale: "0.8",
				Pacing: "none",
				AllowFp16: false
			},
			Gamescope: {
				Enabled: false,
				Width: "1920",
				Height: "1080",
				RefreshRate: "60"
			},
			Memory: {
				Enabled: false,
				Value: "4G"
			}
		}
	} as core.LaunchOptions)
};
