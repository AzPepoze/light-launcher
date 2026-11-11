import { toLogLine, toPluginInfo, toPluginPreview, toPluginSettings } from "./pluginsMap";
import { rpc } from "./rpc";
import type {
	AuthState,
	PluginInfo,
	PluginLogLine,
	PluginMethod,
	PluginPreview,
	PluginSettingsResponse
} from "./types";
import type { WirePluginInfo, WirePluginLogLine, WirePluginPreview, WireSettingsView } from "./wire";

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
	 * Unsigned packages need allowUnsigned, which the user confirms in the install dialog.
	 * Without acceptedPermissions, the permissions from a fresh preview are accepted.
	 */
	install: async (
		source: string,
		allowUnsigned: boolean,
		acceptedPermissions?: string[]
	): Promise<PluginInfo> => {
		const accepted = acceptedPermissions ?? (await pluginsApi.previewPermissions(source)).consent;
		return toPluginInfo(
			await rpc<WirePluginInfo>("PluginInstall", {
				source,
				allowUnsigned,
				acceptedPermissions: accepted
			})
		);
	},

	/** A version that asks for new permissions needs acceptedPermissions that list all of them. */
	update: async (
		id: string,
		allowUnsigned: boolean,
		acceptedPermissions: string[] = []
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
		const lines = await rpc<WirePluginLogLine[] | null>("PluginLogs", { id, lines: LOG_LINE_LIMIT });
		return (lines ?? []).map(toLogLine);
	}
};
