import {
	newPermissionLines,
	toLogLine,
	toPluginInfo,
	toPluginPreview,
	toPluginSettings
} from "./pluginsMap";
import { rpc } from "./rpc";
import type {
	AuthState,
	PluginInfo,
	PluginLogLine,
	PluginMethod,
	PluginPreview,
	PluginSettingsResponse,
	PluginUpdatePreview
} from "./types";
import type {
	WirePluginInfo,
	WirePluginLogLine,
	WirePluginPreview,
	WireSettingsView
} from "./wire";

/** How many recent log lines to ask for. */
const LOG_LINE_LIMIT = 200;

export const pluginsApi = {
	list: async (): Promise<PluginInfo[]> => {
		const plugins = await rpc<WirePluginInfo[] | null>("PluginsList");
		return (plugins ?? []).map(toPluginInfo);
	},

	previewPermissions: async (source: string): Promise<PluginPreview> =>
		toPluginPreview(source, await rpc<WirePluginPreview>("PluginPermissionsPreview", { source })),

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

	/** Previews the newest version at the plugin's install source, with the permissions it adds. */
	previewUpdate: async (plugin: PluginInfo): Promise<PluginUpdatePreview> => {
		if (!plugin.source) throw new Error("This plugin has no install source. Install it again.");
		const preview = await pluginsApi.previewPermissions(plugin.source);
		return { preview, newPermissions: newPermissionLines(plugin.consent, preview.consent) };
	},

	/** Pass the consent list from previewUpdate, so a version with new permissions is accepted. */
	update: async (
		id: string,
		allowUnsigned: boolean,
		acceptedPermissions: string[]
	): Promise<PluginInfo> =>
		toPluginInfo(
			await rpc<WirePluginInfo>("PluginUpdate", { id, allowUnsigned, acceptedPermissions })
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
	}
};
