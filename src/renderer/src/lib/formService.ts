import * as core from "@shared";
import { GetListGpus, DetectLosslessDll } from "@lib/api";
import { DEFAULT_LAUNCH_OPTIONS } from "@lib/constants";
import { createLogger } from "./logger";

const log = createLogger("formService");

/**
 * Merges loaded data with existing options while preserving defaults
 */
export function mergeOptions(
	existing: core.LaunchOptions,
	loaded: Partial<core.LaunchOptions>
): core.LaunchOptions {
	const merged = {
		...existing,
		...loaded
	};

	if (loaded.Extras) {
		merged.Extras = {
			...existing.Extras,
			...loaded.Extras
		};
		if (loaded.Extras.Lsfg) {
			merged.Extras.Lsfg = {
				...existing.Extras.Lsfg,
				...loaded.Extras.Lsfg
			};
		}
		if (loaded.Extras.Gamescope) {
			merged.Extras.Gamescope = {
				...existing.Extras.Gamescope,
				...loaded.Extras.Gamescope
			};
		}
		if (loaded.Extras.Memory) {
			merged.Extras.Memory = {
				...existing.Extras.Memory,
				...loaded.Extras.Memory
			};
		}
	}

	return merged;
}

/**
 * Initializes LaunchOptions with defaults
 */
export function createLaunchOptions(overrides?: Partial<core.LaunchOptions>): core.LaunchOptions {
	const base = structuredClone(DEFAULT_LAUNCH_OPTIONS);
	if (!overrides) return base;
	return mergeOptions(base, overrides);
}

/**
 * Loads GPU list and DLL path, commonly needed for LSFG configuration
 */
export async function loadLsfgResources(): Promise<{ gpus: string[]; dll: string }> {
	const gpus: string[] = [];
	let dll = "";

	try {
		const loadedGpus = await GetListGpus();
		if (loadedGpus && loadedGpus.length > 0) {
			gpus.push(...loadedGpus);
		}
	} catch (e) {
		log.error("Failed to load GPUs", e);
	}

	try {
		const detected = await DetectLosslessDll();
		if (detected) {
			dll = detected;
		}
	} catch (e) {
		log.error("Failed to detect DLL", e);
	}

	return { gpus, dll };
}

/**
 * Parses memory value string to numeric GB
 */
export function parseMemoryValue(value: string): number {
	const match = value.match(/(\d+(?:\.\d+)?)/);
	if (!match) return 4;

	const num = parseFloat(match[1]);
	if (value.toUpperCase().endsWith("M")) {
		return num / 1024; // Convert MB to GB
	}
	return num;
}
