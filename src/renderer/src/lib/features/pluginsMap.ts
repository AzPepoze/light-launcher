import type { PluginInfo, PluginLogLine, PluginPreview, PluginSettingsResponse } from "./types";
import type { WirePermissions, WirePluginInfo, WirePluginPreview, WireRemovedPlugin } from "./wire";

/** Log lines look like "[warn] message". The Go side sends no times. */
const LOG_LINE = /^\[(debug|info|warn|error)\] ?(.*)$/s;

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
		permissions: permissionLabels(manifest.permissions)
	};
}

export function parseLogLine(line: string): PluginLogLine {
	const match = LOG_LINE.exec(line);
	if (!match) return { time: "", level: "info", message: line };
	return { time: "", level: match[1] as PluginLogLine["level"], message: match[2] };
}

/** The API has no settings schema yet, so the form stays empty. */
export function toPluginSettings(values: Record<string, unknown> | null): PluginSettingsResponse {
	return { schema: null, values: values ?? {} };
}

export function isRemovedPlugin(
	payload: WirePluginInfo | WireRemovedPlugin
): payload is WireRemovedPlugin {
	return "removed" in payload;
}
