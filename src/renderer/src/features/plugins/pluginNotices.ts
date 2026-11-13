import type { PluginUpdateStatus } from "../../lib/features/types";
import type { SkippedUpload } from "../../lib/features/pluginUpload";
import { SKIPPED_HEADING, skippedText } from "./pluginCopy";

export type NoticeTone = "success" | "info" | "error";

export interface CheckOutcome {
	tone: NoticeTone;
	message: string;
}

/** Update-check errors that only mean the repository has nothing to check yet. */
const NOTHING_TO_CHECK_ERRORS = [
	"returned HTTP 404",
	"has no published release",
	"latest release is a draft"
];

export const UNAVAILABLE_CHECK_TEXT = "Update check unavailable";

export function isNothingToCheck(value: string | PluginUpdateStatus): boolean {
	if (typeof value !== "string" && value.code === "not_found") return true;
	const text = typeof value === "string" ? value : value.error ?? "";
	return NOTHING_TO_CHECK_ERRORS.some((part) => text.includes(part));
}

/** True when a plugin's update check failed in a way the person should hear about. */
export function isCheckFailure(item: PluginUpdateStatus): boolean {
	return item.error !== null && !isNothingToCheck(item);
}

/** The short tooltip for a failed check. It never shows the raw error. */
export function checkFailureHint(value: string | PluginUpdateStatus): string {
	const code = typeof value === "string" ? "" : value.code;
	const text = typeof value === "string" ? value : value.error ?? "";
	if (code === "rate_limited" || text.includes("rate limit")) {
		return "GitHub limits update checks for a while. Try again later.";
	}
	if (code === "untrusted_host") return "The update source pointed at an untrusted host.";
	return "The update source did not answer. Try again later.";
}

function countText(count: number, one: string, many: string): string {
	return count === 1 ? `1 ${one}` : `${count} ${many}`;
}

/** One message for a whole check run, such as "2 updates available". */
export function checkOutcome(updates: PluginUpdateStatus[]): CheckOutcome {
	const available = updates.filter((item) => item.available).length;
	const failed = updates.filter(isCheckFailure).length;
	const parts: string[] = [];
	if (available > 0) parts.push(countText(available, "update available", "updates available"));
	if (failed > 0) {
		parts.push(countText(failed, "plugin could not be checked", "plugins could not be checked"));
	}
	if (parts.length === 0) return { tone: "success", message: "All plugins are up to date" };
	return { tone: failed > 0 ? "error" : "info", message: parts.join(", ") };
}

/** One message for plugins that were dropped but cannot be installed. */
export function skippedSummary(skipped: SkippedUpload[]): string {
	return `${SKIPPED_HEADING} ${skipped.map(skippedText).join("; ")}`;
}
