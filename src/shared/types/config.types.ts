// These types mirror core/internal/types. The field declarations live in the
// generated wire file so a Go change cannot drift from the renderer by hand.
// Regenerate with: go run ./core/cmd/gentypes
import type {
	WireAppSettings,
	WireLaunchOptions,
	WireScanFolderConfig
} from "../contract/wire.generated";

export type LaunchOptions = WireLaunchOptions;

export type ScanFolderConfig = WireScanFolderConfig;

export type AppSettings = WireAppSettings;
