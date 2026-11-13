import { rpc } from "@lib/features/rpc";
import type { WirePluginInfo } from "@lib/features/wire";
import { updateCheckSources, type UpdateSource } from "./updateChecks";

/** The sources of the installed plugins that can check for updates. */
export async function loadUpdateSources(): Promise<UpdateSource[]> {
	const plugins = await rpc<WirePluginInfo[] | null>("PluginsList");
	return updateCheckSources(plugins ?? []);
}
