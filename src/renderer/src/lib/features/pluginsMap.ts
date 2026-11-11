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

export function permissionLabels(permissions: WirePermissions): string[] {
	return [
		...(permissions.network ?? []).map((host) => `Connect to ${host}`),
		...(permissions.fs ?? []).map((path) => `Access files: ${path}`),
		...(permissions.secrets ? ["Store sign-in details"] : []),
		...(permissions.browser ? ["Open a browser window"] : []),
		...(permissions.exec ? ["Run programs on this computer"] : [])
	];
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
		sources: (wire.sources ?? []).map((source) => source.sourceId)
	};
}

export function toPluginPreview(source: string, wire: WirePluginPreview): PluginPreview {
	const { manifest } = wire;
	return {
		source,
		id: manifest.id,
		name: manifest.name,
		version: manifest.version,
		signed: !wire.unsigned,
		permissions: permissionLabels(manifest.permissions),
		consent: wire.consent ?? []
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
