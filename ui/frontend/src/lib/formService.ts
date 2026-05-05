import * as core from "@bindings/light-launcher/internal/types/models";
import { GetListGpus, DetectLosslessDll } from "@bindings/light-launcher/internal/app/app";
import { DEFAULT_LAUNCH_OPTIONS } from "@lib/constants";

export interface FormState<T> {
	data: T;
	isLoading: boolean;
	error: string;
	isSuccess: boolean;
}

/**
 * Creates a new form state with default values
 */
export function createFormState<T>(defaults: T): FormState<T> {
	return {
		data: structuredClone(defaults),
		isLoading: false,
		error: "",
		isSuccess: false,
	};
}

/**
 * Resets form state to defaults
 */
export function resetFormState<T>(state: FormState<T>, defaults: T): void {
	state.data = structuredClone(defaults);
	state.error = "";
	state.isSuccess = false;
}

/**
 * Merges loaded data with existing options while preserving defaults
 */
export function mergeOptions(existing: core.LaunchOptions, loaded: Partial<core.LaunchOptions>): core.LaunchOptions {
	return {
		...existing,
		...loaded,
	};
}

/**
 * Initializes LaunchOptions with defaults
 */
export function createLaunchOptions(overrides?: Partial<core.LaunchOptions>): core.LaunchOptions {
	return {
		...DEFAULT_LAUNCH_OPTIONS,
		...overrides,
	};
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
		console.error("Failed to load GPUs:", e);
	}

	try {
		const detected = await DetectLosslessDll();
		if (detected) {
			dll = detected;
		}
	} catch (e) {
		console.error("Failed to detect DLL:", e);
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

/**
 * Formats numeric GB value to string with unit
 */
export function formatMemoryValue(gb: number): string {
	if (gb < 1) {
		return Math.round(gb * 1024) + "M";
	}
	return Math.round(gb * 10) / 10 + "G";
}
