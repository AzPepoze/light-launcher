import type { LaunchOptions } from "./config.types";

export interface ProtonTool {
	Name: string;
	Path: string;
	IsSteam: boolean;
	DisplayName: string;
}

export interface PrefixConfigWithProton {
	config: LaunchOptions;
	protonDisplayName: string;
	protonName: string;
	protonPath: string;
	protonIsSteam: boolean;
}

export interface PrefixStats {
	name: string;
	createdAt: number | null;
	sizeBytes: number | null;
}
