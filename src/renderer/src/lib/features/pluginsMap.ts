import { siteLines } from "./consentSites";
import { FILES_PREFIX, NETWORK_PREFIX, permissionTag } from "./permissionTags";
import type {
	PermissionLines,
	PermissionRow,
	PluginInfo,
	PluginLogLine,
	PluginPreview,
	PluginSettingsResponse,
	PluginUpdateStatus,
	PluginUpdatesStatus
} from "./types";
import type {
	WirePermissions,
	WirePluginInfo,
	WirePluginLogLine,
	WirePluginPreview,
	WirePluginUpdateStatus,
	WirePluginUpdatesStatus,
	WireRemovedPlugin,
	WireSettingsView
} from "./wire";

const LOG_LEVELS: readonly PluginLogLine["level"][] = ["debug", "info", "warn", "error"];

// Consent entries from the server, as in core/internal/plugins/consent.go.
const FLAG_LABELS = new Map([
	["secrets", "Store sign-in details"],
	["browser", "Open a browser window"],
	["exec", "Run programs on this computer"]
]);

// Folder grants the server knows, as in manifest_validate.go.
const FOLDER_LABELS = new Map([
	["$GAMES", "Access your games folder"],
	["$DOWNLOADS", "Access your downloads folder"],
	["$DATA", "Access the launcher data folder"]
]);
const APP_DATA_PREFIX = "$APPDATA_";

function isNetworkEntry(entry: string): boolean {
	return entry.startsWith(NETWORK_PREFIX);
}

function folderLabel(folder: string): string {
	if (folder.startsWith(APP_DATA_PREFIX)) {
		return `Access app data folder (${folder.slice(APP_DATA_PREFIX.length)})`;
	}
	return FOLDER_LABELS.get(folder) ?? `Access files: ${folder}`;
}

/** The plain line shown for one non-network consent entry, such as "fs:$GAMES" or "exec". */
export function consentLabel(entry: string): string {
	if (entry.startsWith(FILES_PREFIX)) return folderLabel(entry.slice(FILES_PREFIX.length));
	return FLAG_LABELS.get(entry) ?? entry;
}

function consentRow(entry: string): PermissionRow {
	return { text: consentLabel(entry), ...permissionTag(entry) };
}

/** Splits consent entries into grouped site lines and one row for each other entry. */
export function consentLines(entries: string[]): PermissionLines {
	const hosts = entries
		.filter(isNetworkEntry)
		.map((entry) => entry.slice(NETWORK_PREFIX.length));
	const other = entries.filter((entry) => !isNetworkEntry(entry)).map(consentRow);
	return { sites: siteLines(hosts), other };
}

/** The sorted consent entries for a manifest's permissions, same as the server's PermissionSummary. */
export function permissionSummary(permissions: WirePermissions): string[] {
	const entries = [
		...(permissions.network ?? []).map((host) => `${NETWORK_PREFIX}${host}`),
		...(permissions.fs ?? []).map((folder) => `${FILES_PREFIX}${folder}`),
		...(permissions.secrets ? ["secrets"] : []),
		...(permissions.browser ? ["browser"] : []),
		...(permissions.exec ? ["exec"] : [])
	];
	return entries.sort();
}

/** Consent lines for the entries in next that the installed consent does not have. */
export function newPermissionLines(installed: string[], next: string[]): PermissionLines {
	return consentLines(next.filter((entry) => !installed.includes(entry)));
}

export function toPluginInfo(wire: WirePluginInfo): PluginInfo {
	return {
		id: wire.id,
		name: wire.name,
		version: wire.version,
		enabled: wire.enabled,
		signed: !wire.unsigned,
		status: wire.status === "broken" ? "error" : "ok",
		idle: wire.idle ?? false,
		lastActiveAt: wire.lastActiveAt ?? null,
		error: wire.error || null,
		disabledReason: wire.disabledReason || null,
		author: wire.author,
		description: wire.description,
		homepage: wire.homepage,
		repository: wire.repository || null,
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
		permissions: consentLines(consent),
		consent,
		repositoryChanged: wire.repositoryChanged ?? false,
		installedRepository: wire.installedRepository || null
	};
}

function optionalText(value: string | null | undefined): string | null {
	return value || null;
}

export function toPluginUpdateStatus(wire: WirePluginUpdateStatus): PluginUpdateStatus {
	return {
		id: wire.id,
		installedVersion: wire.installedVersion,
		available: wire.available,
		version: optionalText(wire.version),
		assetUrl: optionalText(wire.assetUrl),
		sha256Url: optionalText(wire.sha256Url),
		sha256: optionalText(wire.sha256),
		notes: optionalText(wire.notes),
		publishedAt: optionalText(wire.publishedAt),
		error: optionalText(wire.error)
	};
}

export function toPluginUpdatesStatus(wire: WirePluginUpdatesStatus | null): PluginUpdatesStatus {
	return {
		checkedAt: optionalText(wire?.checkedAt),
		plugins: (wire?.plugins ?? []).map(toPluginUpdateStatus)
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
