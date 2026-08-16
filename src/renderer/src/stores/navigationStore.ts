import { writable } from "svelte/store";

interface NavigationCommand {
	page: string;
	gamePath?: string;
}

export const navigationCommand = writable<NavigationCommand | null>(null);

export function navigateToEditLsfg(gamePath: string) {
	navigationCommand.set({
		page: "editlsfg",
		gamePath,
	});
}

export function navigateTo(page: string) {
	navigationCommand.set({
		page,
	});
}
