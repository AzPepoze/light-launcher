import { writable } from "svelte/store";

interface NavigationCommand {
	page: string;
	gamePath?: string;
	section?: string;
}

export const navigationCommand = writable<NavigationCommand | null>(null);

/** The Settings section to scroll to once the Settings page is shown. */
export const pendingSettingsSection = writable<string | null>(null);

export function navigateTo(page: string) {
	navigationCommand.set({
		page
	});
}

export function navigateToSettingsSection(section: string) {
	pendingSettingsSection.set(section);
	navigationCommand.set({
		page: "settings",
		section
	});
}
