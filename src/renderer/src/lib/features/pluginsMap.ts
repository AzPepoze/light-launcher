import type { PluginInfo, PluginLogLine, PluginPreview, PluginSettingsResponse } from "./types";
import type {
	WirePermissions,
	WirePluginInfo,
	WirePluginLogLine,
	WirePluginPreview,
	WireRemovedPlugin,
	WireSettingsView
} from "./wire";

const LOG_LEVELS: readonly PluginLogLine["level"][] = ["debug", "info", "warn", "error"];

// Consent entries from the server, as in core/internal/plugins/consent.go.
const NETWORK_PREFIX = "network:";
const FS_PREFIX = "fs:";
const FLAG_LABELS = new Map([
	["secrets", "Store sign-in details"],
	["browser", "Open a browser window"],
	["exec", "Run programs on this computer"]
]);

/** The plain line shown for one consent entry, such as "network:example.com". */
export function consentLabel(entry: string): string {
	if (entry.startsWith(NETWORK_PREFIX)) return `Connect to ${entry.slice(NETWORK_PREFIX.length)}`;
	if (entry.startsWith(FS_PREFIX)) return `Access files: ${entry.slice(FS_PREFIX.length)}`;
	return FLAG_LABELS.get(entry) ?? entry;
}

/** The sorted consent entries for a manifest's permissions, same as the server's PermissionSummary. */
export function permissionSummary(permissions: WirePermissions): string[] {
	const entries = [
		...(permissions.network ?? []).map((host) => `${NETWORK_PREFIX}${host}`),
		...(permissions.fs ?? []).map((folder) => `${FS_PREFIX}${folder}`),
		...(permissions.secrets ? ["secrets"] : []),
		...(permissions.browser ? ["browser"] : []),
		...(permissions.exec ? ["exec"] : [])
	];
	return entries.sort();
}

/** Labels for the entries in next that the installed consent does not have. */
export function newPermissionLines(installed: string[], next: string[]): string[] {
	return next.filter((entry) => !installed.includes(entry)).map(consentLabel);
}

export function toPluginInfo(wire: WirePluginInfo): PluginInfo {
	return {
		id: wire.id,
		name: wire.name,
		version: wire.version,
		enabled: wire.enabled,
		signed: !wire.unsigned,
		status: wire.status === "broken" ? "error" : "ok",
		error: wire.error || null,
		disabledReason: wire.disabledReason || null,
		author: wire.author,
		description: wire.description,
		homepage: wire.homepage,
		hasSettings: wire.hasSettings,
		sources: (wire.sources ?? []).map((source) => source.sourceId),
		source: wire.source || null,
		consent: permissionSummary(wire.permissions)
	};
}

export function toPluginPreview(source: string, wire: WirePluginPreview): PluginPreview {
	const { manifest } = wire;
	const consent = wire.consent ?? [];
	return {
		source,
		id: manifest.id,
		name: manifest.name,
		version: manifest.version,
		signed: !wire.unsigned,
		permissions: consent.map(consentLabel),
		consent
	};
}

/** Unknown or missing levels count as info. */
export function toLogLine(wire: WirePluginLogLine): PluginLogLine {
	const level = LOG_LEVELS.find((known) => known === wire.level) ?? "info";
	return { time: wire.time, level, message: wire.text };
}

export function toPluginSettings(wire: WireSettingsView): PluginSettingsResponse {
	return { schema: wire.schema ?? null, values: wire.values ?? {} };
}

export function isRemovedPlugin(
	payload: WirePluginInfo | WireRemovedPlugin
): payload is WireRemovedPlugin {
	return "removed" in payload;
}
