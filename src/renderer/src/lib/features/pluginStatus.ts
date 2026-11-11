import type { PluginInfo } from "./types";

export type PluginStatusKind = "error" | "disabled" | "sleeping" | "unsigned" | "enabled";

/** The one status a plugin row shows. The first match wins. */
export function pluginStatusKind(
	plugin: Pick<PluginInfo, "status" | "enabled" | "idle" | "signed">
): PluginStatusKind {
	if (plugin.status === "broken") return "error";
	if (!plugin.enabled || plugin.status === "disabled") return "disabled";
	if (plugin.idle) return "sleeping";
	if (!plugin.signed) return "unsigned";
	return "enabled";
}
