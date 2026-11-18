/** The element id of the "Game folders" section on the Settings page. */
export const GAME_FOLDERS_SECTION = "game-folders";

export interface SettingsNavItem {
	id: string;
	label: string;
	/** Platform feature the section needs. The section is hidden when it is missing. */
	requires?: "linux" | "prefix";
}

export interface SettingsNavGroup {
	label: string;
	items: SettingsNavItem[];
}

/** Every Settings section, in page order. Ids match the wrapper element on the page. */
export const SETTINGS_NAV_GROUPS: SettingsNavGroup[] = [
	{
		label: "Appearance",
		items: [
			{ id: "theme-opacity", label: "Theme & opacity" },
			{ id: "wallpaper", label: "Wallpaper background" },
			{ id: "library-covers", label: "Library covers" }
		]
	},
	{
		label: "Library",
		items: [
			{ id: GAME_FOLDERS_SECTION, label: "Game folders" },
			{ id: "hidden-games", label: "Hidden games" },
			{ id: "game-activity", label: "Game activity & integrations" }
		]
	},
	{
		label: "Downloads & updates",
		items: [
			{ id: "downloads", label: "Downloads" },
			{ id: "updates", label: "Updates" }
		]
	},
	{
		label: "Plugins",
		items: [
			{ id: "trusted-keys", label: "Trusted plugin keys" },
			{ id: "community-plugins", label: "Community plugins" }
		]
	},
	{
		label: "Wine & display",
		items: [
			{ id: "prefix-storage", label: "Wine prefix storage", requires: "prefix" },
			{ id: "window-type", label: "Window type" },
			{ id: "display-platform", label: "Display & window platform", requires: "linux" }
		]
	}
];

export interface SettingsFeatures {
	linux: boolean;
	prefix: boolean;
}

/** The nav groups for this platform. Groups left with no items are dropped. */
export function settingsNavGroups(
	features: SettingsFeatures,
	groups: SettingsNavGroup[] = SETTINGS_NAV_GROUPS
): SettingsNavGroup[] {
	return groups
		.map((group) => ({
			label: group.label,
			items: group.items.filter((item) => !item.requires || features[item.requires])
		}))
		.filter((group) => group.items.length > 0);
}

export interface SectionScroll {
	id: string;
	behavior: ScrollBehavior;
}

/** Turns a pending section request into a scroll. Reduced motion jumps instead of animating. */
export function sectionScroll(section: string | null, reducedMotion: boolean): SectionScroll | null {
	if (!section) return null;
	return { id: section, behavior: reducedMotion ? "auto" : "smooth" };
}
