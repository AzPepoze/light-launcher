import {
	newPermissionLines,
	toLogLine,
	toPluginInfo,
	toPluginPreview,
	toPluginSettings,
	toPluginUpdatesStatus
} from "./pluginsMap";
import { rpc } from "./rpc";
import type {
	AuthState,
	PluginInfo,
	PluginLogLine,
	PluginMethod,
	PluginPreview,
	PluginRegistry,
	PluginRegistryEntry,
	PluginSettingsResponse,
	PluginUpdatePreview,
	PluginUpdatesStatus
} from "./types";
import type {
	WirePluginInfo,
	WirePluginLogLine,
	WirePluginPreview,
	WirePluginRegistry,
	WirePluginUpdatesStatus,
	WireRegistryIndexEntry,
	WireSettingsView,
	WireTrustedKey
} from "./wire";

/** How many recent log lines to ask for. */
const LOG_LINE_LIMIT = 200;

/** A full SHA256 digest, lowercase as the server sends it. */
const REGISTRY_SHA256 = /^[0-9a-f]{64}$/;

/** One ed25519 public key that plugin signatures are verified against. */
export interface TrustedKey {
	/** Standard base64 of the 32-byte public key. */
	base64: string;
	/** Short SHA256 fingerprint, 16 hex characters. */
	fingerprint: string;
}

function toTrustedKey(wire: WireTrustedKey): TrustedKey {
	return { base64: wire.base64, fingerprint: wire.fingerprint };
}

function toTrustedKeys(keys: WireTrustedKey[] | null): TrustedKey[] {
	return (keys ?? []).map(toTrustedKey);
}

function toRegistryEntry(wire: WireRegistryIndexEntry): PluginRegistryEntry {
	return {
		id: wire.id,
		name: wire.name,
		description: wire.description,
		author: wire.author,
		repository: wire.repository || null,
		version: wire.version,
		downloadUrl: wire.downloadUrl,
		sha256: wire.sha256,
		homepage: wire.homepage || null
	};
}

/** Reads the PluginRegistry reply. A failed load stays in error so the dialog can offer a retry. */
function toPluginRegistry(wire: WirePluginRegistry): PluginRegistry {
	return {
		configured: wire.configured,
		url: wire.url ?? "",
		entries: (wire.entries ?? []).map(toRegistryEntry),
		error: wire.error || null
	};
}

/** A community entry can be installed only from an https package with a full SHA256. */
function isInstallableRegistryEntry(entry: PluginRegistryEntry): boolean {
	return entry.downloadUrl.startsWith("https://") && REGISTRY_SHA256.test(entry.sha256);
}

/** The community entries to list: installable ones that are not installed already. */
export function communityEntries(
	entries: PluginRegistryEntry[],
	installedIds: readonly string[]
): PluginRegistryEntry[] {
	const installed = new Set(installedIds);
	return entries.filter((entry) => !installed.has(entry.id) && isInstallableRegistryEntry(entry));
}

export const pluginsApi = {
	list: async (): Promise<PluginInfo[]> => {
		const plugins = await rpc<WirePluginInfo[] | null>("PluginsList");
		return (plugins ?? []).map(toPluginInfo);
	},

	/**
	 * The community index from PluginRegistry. An unconfigured URL returns an empty list and the
	 * server makes no request. A load failure is in error, not a rejected call.
	 */
	registry: async (): Promise<PluginRegistry> =>
		toPluginRegistry(await rpc<WirePluginRegistry>("PluginRegistry")),

	/** The keys a plugin SIGNATURE may match. Empty when no key is trusted. */
	trustedKeys: async (): Promise<TrustedKey[]> =>
		toTrustedKeys(await rpc<WireTrustedKey[] | null>("PluginTrustedKeys")),

	/** Adds one base64 ed25519 public key and returns the refreshed list. */
	addTrustedKey: async (key: string): Promise<TrustedKey[]> =>
		toTrustedKeys(await rpc<WireTrustedKey[]>("PluginTrustedKeyAdd", { key })),

	/** Removes one base64 ed25519 public key and returns the refreshed list. */
	removeTrustedKey: async (key: string): Promise<TrustedKey[]> =>
		toTrustedKeys(await rpc<WireTrustedKey[]>("PluginTrustedKeyRemove", { key })),

	previewPermissions: async (source: string, sha256 = ""): Promise<PluginPreview> =>
		toPluginPreview(
			source,
			await rpc<WirePluginPreview>("PluginPermissionsPreview", { source, sha256 })
		),

	/**
	 * Pass the consent list the user saw in the preview. The server rejects any other list.
	 * Unsigned packages also need allowUnsigned.
	 * This is a strict new install; use installOrUpdate for a dropped or picked package.
	 */
	install: async (
		source: string,
		allowUnsigned: boolean,
		acceptedPermissions: string[]
	): Promise<PluginInfo> =>
		toPluginInfo(
			await rpc<WirePluginInfo>("PluginInstall", {
				source,
				allowUnsigned,
				acceptedPermissions
			})
		),

	/**
	 * Installs a package, or updates the installed plugin with the same id in place, keeping its
	 * data folder and secrets. A package for the version already installed fails with ErrNoUpdate
	 * instead of replacing it.
	 */
	installOrUpdate: async (
		source: string,
		allowUnsigned: boolean,
		acceptedPermissions: string[]
	): Promise<PluginInfo> =>
		toPluginInfo(
			await rpc<WirePluginInfo>("PluginInstallOrUpdate", {
				source,
				allowUnsigned,
				acceptedPermissions
			})
		),

	/**
	 * Previews a newer version, with the permissions it adds. The release source and checksum come from
	 * the last update check; without them the plugin's recorded install source is used.
	 */
	previewUpdate: async (
		plugin: PluginInfo,
		source = "",
		sha256 = ""
	): Promise<PluginUpdatePreview> => {
		const target = source || plugin.source;
		if (!target) throw new Error("This plugin has no install source. Install it again.");
		const preview = await pluginsApi.previewPermissions(target, sha256);
		return {
			preview,
			newPermissions: newPermissionLines(plugin.consent, preview.consent),
			source: target,
			sha256: sha256 || null
		};
	},

	/** Pass the consent list from previewUpdate, so a version with new permissions is accepted. */
	update: async (
		id: string,
		allowUnsigned: boolean,
		acceptedPermissions: string[],
		source = "",
		sha256 = ""
	): Promise<PluginInfo> =>
		toPluginInfo(
			await rpc<WirePluginInfo>("PluginUpdate", {
				id,
				allowUnsigned,
				acceptedPermissions,
				source,
				sha256
			})
		),

	/** Removal keeps the plugin's data folder. */
	remove: (id: string) => rpc<void>("PluginRemove", { id, deleteData: false }),

	rollback: async (id: string): Promise<PluginInfo> =>
		toPluginInfo(await rpc<WirePluginInfo>("PluginRollback", { id })),

	setEnabled: async (id: string, enabled: boolean): Promise<PluginInfo> =>
		toPluginInfo(await rpc<WirePluginInfo>("PluginSetEnabled", { id, enabled })),

	settingsGet: async (id: string): Promise<PluginSettingsResponse> =>
		toPluginSettings(await rpc<WireSettingsView>("PluginSettingsGet", { id })),

	settingsSet: (id: string, values: Record<string, unknown>) =>
		rpc<void>("PluginSettingsSet", { id, settings: values }),

	call: <T = unknown>(plugin: string, method: PluginMethod, params: object) =>
		rpc<T>("PluginCall", { plugin, method, params }),

	/** Runs one command the plugin declares in its manifest. */
	command: (plugin: string, command: string, params: object = {}) =>
		rpc<unknown>("PluginCall", { plugin, command, params }),

	authStatus: (plugin: string, source: string) =>
		rpc<AuthState>("PluginCall", {
			plugin,
			method: "source.authStatus",
			params: { source }
		}),

	logs: async (id: string): Promise<PluginLogLine[]> => {
		const lines = await rpc<WirePluginLogLine[] | null>("PluginLogs", {
			id,
			lines: LOG_LINE_LIMIT
		});
		return (lines ?? []).map(toLogLine);
	},

	/** Asks GitHub for newer releases. With no id, checks every plugin that has a repository. */
	checkUpdates: async (id?: string): Promise<PluginUpdatesStatus> =>
		toPluginUpdatesStatus(
			await rpc<WirePluginUpdatesStatus>("PluginUpdatesCheck", id ? { id } : {})
		),

	/** The cached results of the last check. */
	updatesStatus: async (): Promise<PluginUpdatesStatus> =>
		toPluginUpdatesStatus(await rpc<WirePluginUpdatesStatus>("PluginUpdatesStatus"))
};
