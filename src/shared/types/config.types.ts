export interface LsfgConfig {
	Enabled: boolean;
	Multiplier: string;
	PerfMode: boolean;
	DllPath: string;
	Gpu: string;
	FlowScale: string;
	Pacing: string;
	AllowFp16: boolean;
}

export interface GamescopeConfig {
	Enabled: boolean;
	Width: string;
	Height: string;
	OutputWidth: string;
	OutputHeight: string;
	RefreshRate: string;
	FramerateLimit: string;
	WindowMode: string;
	Scaler: string;
	Filter: string;
	Sharpness: string;
	HDR: boolean;
	AdaptiveSync: boolean;
	Mangoapp: boolean;
	CustomArgs: string;
}

export interface MemoryConfig {
	Enabled: boolean;
	Value: string;
}

export interface ExtrasConfig {
	EnableMangoHud: boolean;
	EnableGamemode: boolean;
	Lsfg: LsfgConfig;
	Gamescope: GamescopeConfig;
	Memory: MemoryConfig;
}

export interface LaunchOptions {
	ID: string;
	Name: string;
	LauncherPath: string;
	GamePath: string;
	UseGamePath: boolean;
	PrefixPath: string;
	ProtonPath: string;
	UseCustomProton: boolean;
	CustomArgs: string;
	Extras: ExtrasConfig;
}

export interface ScanFolderConfig {
	Path: string;
	Depth: number;
	ExcludeNames: string[];
}

export interface AppSettings {
	TransparentMode: boolean;
	NativeWayland?: boolean;
	ScanFolders: string[];
	ScanFolderConfigs: ScanFolderConfig[];
	Blacklist: string[];
	CustomPrefixDir?: string;
}
