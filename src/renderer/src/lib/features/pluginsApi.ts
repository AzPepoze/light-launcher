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
	PluginSettingsResponse,
	PluginUpdatePreview,
	PluginUpdatesStatus
} from "./types";
import type {
	WirePluginInfo,
	WirePluginLogLine,
	WirePluginPreview,
	WirePluginUpdatesStatus,
	WireSettingsView
} from "./wire";

/** How many recent log lines to ask for. */
const LOG_LINE_LIMIT = 200;

export const pluginsApi = {
	list: async (): Promise<PluginInfo[]> => {
		const plugins = await rpc<WirePluginInfo[] | null>("PluginsList");
		return (plugins ?? []).map(toPluginInfo);
	},

	previewPermissions: async (source: string, sha256 = ""): Promise<PluginPreview> =>
		toPluginPreview(
			source,
			await rpc<WirePluginPreview>("PluginPermissionsPreview", { source, sha256 })
		),

	/**
	 * Pass the consent list the user saw in the preview. The server rejects any other list.
	 * Unsigned packages also need allowUnsigned.
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
