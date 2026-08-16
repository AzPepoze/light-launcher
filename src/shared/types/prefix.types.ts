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

export interface ProtonVariant {
	ID: string;
	Name: string;
	Description: string;
	RepoOwner: string;
	RepoName: string;
}

export interface GitHubAsset {
	name: string;
	browser_download_url: string;
	size: number;
	content_type: string;
}

export interface GitHubRelease {
	tag_name: string;
	name: string;
	published_at: string;
	html_url: string;
	assets: GitHubAsset[];
}
