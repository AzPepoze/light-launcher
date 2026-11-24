import {
	CreatePrefix,
	GetDriveSetup,
	GetPrefixBaseDir,
	ListPartitions,
	ListPrefixDrives,
	ListPrefixes,
	LoadPrefixConfigWithProton,
	RemovePrefix,
	RunPrefixTool,
	SaveDriveSetup,
	SavePrefixConfig
} from "@lib/api";
import * as core from "@shared";
import { notifications } from "@stores/notificationStore";
import { plainMessage } from "./plainMessage";
import { createLogger } from "./logger";

const log = createLogger("prefixService");

export interface PrefixData {
	availablePrefixes: string[];
	baseDir: string;
}

/**
 * Resolves the default prefix name from a list of prefixes
 */
export function getDefaultPrefixName(prefixes: string[]): string {
	if (!prefixes || prefixes.length === 0) return "Default";
	if (prefixes.includes("Default")) return "Default";
	if (prefixes.includes("default")) return "default";
	return prefixes[0];
}

/**
 * Fetches available prefixes and the base directory
 */
export async function getPrefixData(): Promise<PrefixData> {
	try {
		const [list, base] = await Promise.all([ListPrefixes(), GetPrefixBaseDir()]);
		log.debug("Fetched prefix data", {
			count: list?.length || 0,
			baseDir: base
		});
		return {
			availablePrefixes: Array.isArray(list) ? list : [],
			baseDir: base || ""
		};
	} catch (err) {
		log.error("Failed to fetch prefix data", err);
		return { availablePrefixes: [], baseDir: "" };
	}
}

/**
 * Loads configuration for a specific prefix
 * Backend handles Proton matching and resolution
 */
export async function getPrefixConfig(
	name: string,
	baseDir: string
): Promise<{
	path: string;
	options: core.LaunchOptions | null;
	selectedProton: string;
}> {
	const path = `${baseDir}/${name}`;
	try {
		log.info("getPrefixConfig for prefix", { name });
		const result = await LoadPrefixConfigWithProton(name);
		log.debug("Backend returned config", result);
		if (!result) {
			log.info("No config found");
			return { path, options: null, selectedProton: "" };
		}

		log.info("Returning config with proton", {
			proton: result.protonDisplayName
		});
		return {
			path,
			options: result.config,
			selectedProton: result.protonDisplayName
		};
	} catch (e) {
		log.error("Failed to load config", { prefix: name, error: e });
		return { path, options: null, selectedProton: "" };
	}
}

/**
 * Saves configuration for a prefix
 */
export async function savePrefixDefaults(
	prefixPath: string,
	options: core.LaunchOptions,
	protonTool: core.ProtonTool | null
): Promise<void> {
	if (!prefixPath) return;
	const name = prefixPath.split("/").pop() || "Default";

	log.info("Saving prefix defaults", { name, protonTool });

	// Set ProtonPath from the selected tool
	if (protonTool) {
		options.ProtonPath = protonTool.Path;
		log.debug("Set proton path", { path: protonTool.Path });
	} else {
		options.ProtonPath = "";
		log.debug("Cleared proton path");
	}

	log.debug("Calling SavePrefixConfig", {
		name,
		protonPath: options.ProtonPath
	});
	await notifications.withNotification(SavePrefixConfig(name, options), {
		success: "Prefix defaults saved!",
		error: "Failed to save configuration"
	});
	log.info("Prefix defaults saved");
}

/**
 * Creates a new prefix and returns updated data
 */
export async function createNewPrefix(name: string): Promise<void> {
	if (!name) return;
	await notifications.withNotification(CreatePrefix(name), {
		success: `Created prefix "${name}"`,
		error: "Failed to create prefix"
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
		error: "Failed to delete prefix"
	});
}

/**
 * Executes a prefix tool (Winecfg, etc.)
 */
export async function executePrefixTool(
	prefixPath: string,
	toolName: string,
	protonPath: string
): Promise<void> {
	if (!prefixPath) {
		notifications.error("Please select or create a prefix first.");
		return;
	}

	try {
		await RunPrefixTool(prefixPath, toolName, protonPath);
	} catch (err) {
		notifications.error(`Failed to run ${toolName}: ${plainMessage(err)}`);
		throw err;
	}
}

/** Returns the disks the launcher can bind. An empty list is returned when the read fails. */
export async function loadPartitions(): Promise<core.Partition[]> {
	try {
		return await ListPartitions();
	} catch (err) {
		notifications.error(`Failed to list disks: ${plainMessage(err)}`);
		return [];
	}
}

/** Returns the drive letters bound in a prefix. Shows a message and returns an empty list on failure. */
export async function loadPrefixDrives(prefixPath: string): Promise<core.PrefixDrive[]> {
	try {
		return await ListPrefixDrives(prefixPath);
	} catch (err) {
		notifications.error(`Failed to list drives: ${plainMessage(err)}`);
		return [];
	}
}

/** Returns the global drive setup applied to every prefix. Returns null on failure, so callers never mistake it for an empty setup. */
export async function loadDriveSetup(): Promise<core.SetupEntry[] | null> {
	try {
		return await GetDriveSetup();
	} catch (err) {
		notifications.error(`Failed to load drive setup: ${plainMessage(err)}`);
		return null;
	}
}

/** Saves the global drive setup. Returns false when the backend rejects it. */
export async function saveDriveSetup(entries: core.SetupEntry[]): Promise<boolean> {
	try {
		await SaveDriveSetup(entries);
		notifications.success("Drive setup saved");
		return true;
	} catch (err) {
		notifications.error(`Could not save drive setup: ${plainMessage(err)}`);
		return false;
	}
}
