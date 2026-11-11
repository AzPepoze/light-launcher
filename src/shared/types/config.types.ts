// These types mirror core/internal/types. The field declarations live in the
// generated wire file so a Go change cannot drift from the renderer by hand.
// Regenerate with: go run ./core/cmd/gentypes
import type {
	WireAppSettings,
	WireExtrasConfig,
	WireGamescopeConfig,
	WireLaunchOptions,
	WireLsfgConfig,
	WireMemoryConfig,
	WireScanFolderConfig
} from "../contract/wire.generated";

export type LsfgConfig = WireLsfgConfig;

export type GamescopeConfig = WireGamescopeConfig;

export type MemoryConfig = WireMemoryConfig;

export type ExtrasConfig = WireExtrasConfig;

export type LaunchOptions = WireLaunchOptions;

export type ScanFolderConfig = WireScanFolderConfig;

export type AppSettings = WireAppSettings;
