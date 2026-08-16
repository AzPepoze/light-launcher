import { writable } from "svelte/store";

export interface UserSettings {
	theme: "light" | "dark";
	transparency: number; // 0.0 to 1.0
	backgroundImagePath: string;
}

const defaultSettings: UserSettings = {
	theme: (localStorage.getItem("theme") as "light" | "dark") || "dark",
	transparency: parseFloat(localStorage.getItem("transparency") || "0.85"),
	backgroundImagePath: localStorage.getItem("backgroundImagePath") || "",
};

export const settingsStore = writable<UserSettings>(defaultSettings);

settingsStore.subscribe((val) => {
	localStorage.setItem("theme", val.theme);
	localStorage.setItem("transparency", val.transparency.toString());
	localStorage.setItem("backgroundImagePath", val.backgroundImagePath);
	document.documentElement.dataset.theme = val.theme;
	document.documentElement.style.setProperty('--app-opacity', val.transparency.toString());
});
