import type { DropdownOption } from "@components/shared/Dropdown.svelte";
import type { WireSourceUpdateCheck, WireUpdateCheckSettings } from "@lib/features/wire";

export type SourceMode = "inherit" | "custom" | "off";

/** The global values with the defaults filled in. */
export interface GlobalUpdateChecks {
	enabled: boolean;
	onStartup: boolean;
	intervalHours: number;
	recheckHours: number;
}

/** What one source row shows: its mode and the skip window it uses. */
export interface SourceUpdateView {
	mode: SourceMode;
	recheckHours: number;
}

export interface UpdateSource {
	id: string;
	label: string;
}

/** Limits and defaults from core/internal/types/types.go. */
export const UPDATE_HOURS_MAX = 720;
export const DEFAULT_UPDATE_INTERVAL_HOURS = 24;
export const DEFAULT_UPDATE_RECHECK_HOURS = 24;
const UPDATE_CAPABILITY = "updateCheck";

export const INTERVAL_OPTIONS: DropdownOption[] = [
	{ value: "0", label: "Never" },
	{ value: "6", label: "6 hours" },
	{ value: "12", label: "12 hours" },
	{ value: "24", label: "24 hours" },
	{ value: "48", label: "2 days" },
	{ value: "168", label: "7 days" }
];

export const RECHECK_OPTIONS: DropdownOption[] = [
	{ value: "0", label: "Do not skip" },
	{ value: "1", label: "1 hour" },
	{ value: "6", label: "6 hours" },
	{ value: "24", label: "24 hours" },
	{ value: "72", label: "3 days" }
];

export const SOURCE_MODE_OPTIONS: DropdownOption[] = [
	{ value: "inherit", label: "Use global" },
	{ value: "custom", label: "Custom" },
	{ value: "off", label: "Off" }
];

/** Clamps an hour count to 0 to 720 and rounds it. A value that is not finite becomes the fallback. */
export function clampHours(value: number, fallback: number): number {
	if (!Number.isFinite(value)) return fallback;
	return Math.min(UPDATE_HOURS_MAX, Math.max(0, Math.round(value)));
}

/** Option text for an hour count. Whole days read as days. */
export function hoursLabel(hours: number): string {
	if (hours > 0 && hours % 24 === 0) {
		const days = hours / 24;
		return days === 1 ? "1 day" : `${days} days`;
	}
	return hours === 1 ? "1 hour" : `${hours} hours`;
}

/** The list with the current value added when it is missing, so a hand-edited value still shows. */
export function optionsWithHours(options: DropdownOption[], hours: number): DropdownOption[] {
	const value = String(hours);
	if (options.some((option) => option.value === value)) return options;
	const added = [...options, { value, label: hoursLabel(hours) }];
	return added.sort((a, b) => Number(a.value) - Number(b.value));
}

/** Reads a dropdown value as hours. */
export function hoursFromOption(value: string, fallback: number): number {
	return clampHours(Number(value), fallback);
}

/** Reads a source mode dropdown value. Unknown values mean inherit. */
export function sourceModeFromOption(value: string): SourceMode {
	return value === "custom" || value === "off" ? value : "inherit";
}

/** The global values with defaults filled in and hours clamped, the way the backend resolves them. */
export function globalUpdateChecks(settings: WireUpdateCheckSettings | undefined): GlobalUpdateChecks {
	return {
		enabled: settings?.enabled ?? true,
		onStartup: settings?.onStartup ?? true,
		intervalHours: clampHours(
			settings?.intervalHours ?? DEFAULT_UPDATE_INTERVAL_HOURS,
			DEFAULT_UPDATE_INTERVAL_HOURS
		),
		recheckHours: clampHours(
			settings?.recheckHours ?? DEFAULT_UPDATE_RECHECK_HOURS,
			DEFAULT_UPDATE_RECHECK_HOURS
		)
	};
}

/** The mode of a source. A missing or unknown mode means inherit. */
export function sourceMode(settings: WireUpdateCheckSettings | undefined, sourceId: string): SourceMode {
	return sourceModeFromOption(settings?.sources?.[sourceId]?.mode ?? "inherit");
}

/** The skip window a source uses: its own value when set, otherwise the global one. */
export function sourceView(settings: WireUpdateCheckSettings | undefined, sourceId: string): SourceUpdateView {
	const global = globalUpdateChecks(settings);
	const own = settings?.sources?.[sourceId]?.recheckHours;
	return {
		mode: sourceMode(settings, sourceId),
		recheckHours: clampHours(own ?? global.recheckHours, global.recheckHours)
	};
}

/** The help line under a source's name. */
export function sourceHelp(mode: SourceMode): string {
	if (mode === "custom") return "Uses its own skip window.";
	if (mode === "off") return "No background checks. Check now still works.";
	return "Uses the global settings.";
}

/** The settings with some global values changed. The four top-level values are always filled in. */
export function withGlobalUpdateChecks(
	settings: WireUpdateCheckSettings | undefined,
	patch: Partial<GlobalUpdateChecks>
): WireUpdateCheckSettings {
	return { ...settings, ...globalUpdateChecks({ ...settings, ...patch }) };
}

function withSource(
	settings: WireUpdateCheckSettings | undefined,
	sourceId: string,
	change: WireSourceUpdateCheck
): WireUpdateCheckSettings {
	const sources = settings?.sources ?? {};
	return { ...settings, sources: { ...sources, [sourceId]: { ...sources[sourceId], ...change } } };
}

export function withSourceMode(
	settings: WireUpdateCheckSettings | undefined,
	sourceId: string,
	mode: SourceMode
): WireUpdateCheckSettings {
	return withSource(settings, sourceId, { mode });
}

export function withSourceRecheck(
	settings: WireUpdateCheckSettings | undefined,
	sourceId: string,
	hours: number
): WireUpdateCheckSettings {
	return withSource(settings, sourceId, {
		recheckHours: clampHours(hours, DEFAULT_UPDATE_RECHECK_HOURS)
	});
}

interface PluginSourceLike {
	sourceId: string;
	label: string;
	capabilities: string[] | null;
}

interface PluginLike {
	enabled: boolean;
	sources: PluginSourceLike[] | null;
}

/** The sources of enabled plugins that can check for updates, one per source id, sorted by label. */
export function updateCheckSources(plugins: PluginLike[]): UpdateSource[] {
	const sources = plugins
		.filter((plugin) => plugin.enabled)
		.flatMap((plugin) => plugin.sources ?? [])
		.filter((source) => source.capabilities?.includes(UPDATE_CAPABILITY) ?? false);
	const byId = new Map<string, UpdateSource>();
	for (const source of sources) byId.set(source.sourceId, { id: source.sourceId, label: source.label });
	return [...byId.values()].sort((a, b) => a.label.localeCompare(b.label));
}
