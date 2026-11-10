import { rpc } from "./rpc";
import type {
	AuthState,
	PluginInfo,
	PluginLogLine,
	PluginMethod,
	PluginPreview,
	PluginSettingsResponse
} from "./types";

export const pluginsApi = {
	list: () => rpc<{ items: PluginInfo[] }>("PluginsList"),
	previewPermissions: (source: string) =>
		rpc<PluginPreview>("PluginPermissionsPreview", { source }),
	install: (source: string, acceptedPermissions: string[]) =>
		rpc<PluginInfo>("PluginInstall", { source, acceptedPermissions }),
	update: (id: string) => rpc<PluginInfo>("PluginUpdate", { id }),
	remove: (id: string) => rpc<void>("PluginRemove", { id }),
	rollback: (id: string) => rpc<PluginInfo>("PluginRollback", { id }),
	setEnabled: (id: string, enabled: boolean) =>
		rpc<PluginInfo>("PluginSetEnabled", { id, enabled }),
	settingsGet: (id: string) => rpc<PluginSettingsResponse>("PluginSettingsGet", { id }),
	settingsSet: (id: string, values: Record<string, unknown>) =>
		rpc<void>("PluginSettingsSet", { id, values }),
	call: <T = unknown>(plugin: string, method: PluginMethod, params: object) =>
		rpc<T>("PluginCall", { plugin, method, params }),
	authStatus: (plugin: string, source: string) =>
		rpc<AuthState>("PluginCall", {
			plugin,
			method: "source.authStatus",
			params: { source }
		}),
	logs: (id: string) => rpc<{ lines: PluginLogLine[] }>("PluginLogs", { id })
};
