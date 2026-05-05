import {
	CreatePrefix,
	GetPrefixBaseDir,
	ListPrefixes,
	LoadPrefixConfig,
	RunPrefixTool,
	SavePrefixConfig,
	RemovePrefix,
} from "@bindings/light-launcher/internal/app/app";
import * as core from "@bindings/light-launcher/internal/types/models";
import { notifications } from "@stores/notificationStore";

export interface PrefixData {
	availablePrefixes: string[];
	baseDir: string;
}

/**
 * Fetches available prefixes and the base directory
 */
export async function getPrefixData(): Promise<PrefixData> {
	try {
		const [list, base] = await Promise.all([
			ListPrefixes(),
			GetPrefixBaseDir(),
		]);
		return {
			availablePrefixes: Array.isArray(list) ? list : [],
			baseDir: base || "",
		};
	} catch (err) {
		console.error("Failed to fetch prefix data:", err);
		return { availablePrefixes: [], baseDir: "" };
	}
}

/**
 * Loads configuration for a specific prefix and handles Proton matching
 */
export async function getPrefixConfig(
	name: string,
	baseDir: string,
	protonVersions: core.ProtonTool[]
): Promise<{ 
	path: string; 
	options: core.LaunchOptions | null; 
	selectedProton: string 
}> {
	const path = `${baseDir}/${name}`;
	try {
		const config = await LoadPrefixConfig(name);
		let selectedProton = "";

		if (config) {
			const match = protonVersions.find((p) => p.Path === config.ProtonPath);
			if (match) {
				selectedProton = match.DisplayName;
			} else if (config.ProtonPath) {
				selectedProton = config.ProtonPath;
			}
			return { path, options: config, selectedProton };
		}
	} catch (e) {
		console.error(`Failed to load config for prefix ${name}:`, e);
	}
	return { path, options: null, selectedProton: "" };
}

/**
 * Saves configuration for a prefix
 */
export async function savePrefixDefaults(
	prefixPath: string,
	options: core.LaunchOptions,
	selectedProton: string,
	protonVersions: core.ProtonTool[]
): Promise<void> {
	if (!prefixPath) return;
	const name = prefixPath.split("/").pop() || "Default";

	const tool = protonVersions.find((p) => p.DisplayName === selectedProton);
	options.ProtonPath = tool ? tool.Path : (selectedProton.includes("/") ? selectedProton : "");

	await notifications.withNotification(SavePrefixConfig(name, options), {
		success: "Prefix defaults saved!",
		error: "Failed to save configuration",
	});
}

/**
 * Creates a new prefix and returns updated data
 */
export async function createNewPrefix(name: string): Promise<void> {
	if (!name) return;
	await notifications.withNotification(CreatePrefix(name), {
		success: `Created prefix "${name}"`,
		error: "Failed to create prefix",
	});
}

/**
 * Removes a prefix
 */
export async function deletePrefix(name: string): Promise<void> {
	if (name === "Default") {
		notifications.add("Cannot delete Default prefix", "error");
		return;
	}
	await notifications.withNotification(RemovePrefix(name), {
		success: `Deleted prefix "${name}"`,
		error: "Failed to delete prefix",
	});
}

/**
 * Executes a prefix tool (Winecfg, etc.)
 */
export async function executePrefixTool(
	prefixPath: string,
	toolName: string,
	selectedProton: string,
	protonVersions: core.ProtonTool[]
): Promise<void> {
	if (!prefixPath) {
		notifications.error("Please select or create a prefix first.");
		return;
	}

	const selectedTool = protonVersions.find((p) => p.DisplayName === selectedProton);
	const protonPath = selectedTool ? selectedTool.Path : (selectedProton.includes("/") ? selectedProton : "");

	try {
		await RunPrefixTool(prefixPath, toolName, protonPath);
	} catch (err) {
		notifications.error(`Failed to run ${toolName}: ${err}`);
		throw err;
	}
}
