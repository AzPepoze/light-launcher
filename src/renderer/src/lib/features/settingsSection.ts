/** The element id of the "Game folders" section on the Settings page. */
export const GAME_FOLDERS_SECTION = "game-folders";

export interface SettingsNavItem {
	id: string;
	label: string;
	/** Platform feature the section needs. The section is hidden when it is missing. */
	requires?: "linux" | "prefix";
	/** Extra words the search box matches, such as words used in the section copy. */
	keywords?: string[];
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
			{
				id: "theme-opacity",
				label: "Theme & opacity",
				keywords: ["theme", "dark", "light", "opacity", "transparent", "background"]
			},
			{ id: "wallpaper", label: "Wallpaper background", keywords: ["image", "background"] },
			{
				id: "library-covers",
				label: "Library covers",
				keywords: ["card size", "cover", "blur", "grid"]
			}
		]
	},
	{
		label: "Library",
		items: [
			{
				id: GAME_FOLDERS_SECTION,
				label: "Game folders",
				keywords: ["scan", "folder", "path", "exe"]
			},
			{
				id: "hidden-games",
				label: "Hidden games",
				keywords: ["blacklist", "hide", "hidden", "restore"]
			},
			{
				id: "game-activity",
				label: "Game activity & integrations",
				keywords: ["playtime", "discord", "rich presence"]
			}
		]
	},
	{
		label: "Downloads & updates",
		items: [
			{ id: "downloads", label: "Downloads", keywords: ["download", "speed", "archive"] },
			{
				id: "updates",
				label: "Updates",
				keywords: ["check", "version", "auto", "background"]
			}
		]
	},
	{
		label: "Plugins",
		items: [
			{
				id: "trusted-keys",
				label: "Trusted plugin keys",
				keywords: ["signature", "signed", "key", "plugin", "ed25519"]
			},
			{
				id: "community-plugins",
				label: "Community plugins",
				keywords: ["registry", "plugin", "index", "url"]
			}
		]
	},
	{
		label: "Wine & display",
		items: [
			{
				id: "prefix-storage",
				label: "Wine prefix storage",
				requires: "prefix",
				keywords: ["wine", "prefix", "storage", "location", "drive"]
			},
			{ id: "window-type", label: "Window type", keywords: ["browser", "window", "open"] },
			{
				id: "display-platform",
				label: "Display & window platform",
				requires: "linux",
				keywords: ["wayland", "x11", "display", "xwayland", "ozone"]
			}
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

/** Keeps the items whose label or keywords contain the query. Groups left empty are dropped. */
export function settingsNavFilter(groups: SettingsNavGroup[], query: string): SettingsNavGroup[] {
	const needle = query.trim().toLowerCase();
	if (needle === "") return groups;
	return groups
		.map((group) => ({
			label: group.label,
			items: group.items.filter((item) => navItemMatches(item, needle))
		}))
		.filter((group) => group.items.length > 0);
}

function navItemMatches(item: SettingsNavItem, needle: string): boolean {
	const terms = [item.label, ...(item.keywords ?? [])];
	return terms.some((term) => term.toLowerCase().includes(needle));
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
