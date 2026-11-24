import { get, writable } from "svelte/store";

export type LeaveChoice = "stay" | "discard" | "save";

export interface LeaveGuard {
	isDirty: () => boolean;
	/** Resolves false when the save failed, so the page stays open. */
	save: () => Promise<boolean>;
}

interface LeavePrompt {
	answer: (choice: LeaveChoice) => void;
}

/** The page that currently holds unsaved changes, or null. */
export const leaveGuard = writable<LeaveGuard | null>(null);

/** The open "unsaved changes" question, or null. Read by the dialog in App. */
export const leavePrompt = writable<LeavePrompt | null>(null);

/** Asks before leaving a page with unsaved changes. Returns true when the page may be left. */
export async function confirmLeave(): Promise<boolean> {
	const guard = get(leaveGuard);
	if (!guard || !guard.isDirty()) return true;

	const choice = await askLeaveChoice();
	if (choice === "stay") return false;
	if (choice === "save") return guard.save();
	return true;
}

function askLeaveChoice(): Promise<LeaveChoice> {
	return new Promise((resolve) => {
		leavePrompt.set({
			answer: (choice) => {
				leavePrompt.set(null);
				resolve(choice);
			}
		});
	});
}
