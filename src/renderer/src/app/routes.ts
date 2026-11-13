import type { PlatformFeatures } from "@shared";
import protonIcon from "../icons/protron_forked.png";

export interface AppRoute {
	/** Stable navigation id, also used by navigation commands. */
	id: string;
	/** Navbar tooltip and command palette name base. */
	label: string;
	icon: string;
	order: number;
	/** Platform feature that must be on for the route to be reachable. */
	feature?: keyof PlatformFeatures;
	/** Command palette wording when the navbar label would read badly. */
	paletteLabel?: string;
	isCustomIcon?: boolean;
	/** Reachable route that is never listed in navigation or the command palette. */
	hidden?: boolean;
	/** Listed in navigation at the bottom of the rail, below the other pages. */
	bottom?: boolean;
}

type RouteDefinition = Omit<AppRoute, "order">;

// Array order is the canonical page order used by the navbar and page transitions.
const routeDefinitions: RouteDefinition[] = [
	{ id: "home", label: "Home", icon: "home" },
	{ id: "library", label: "Library", icon: "collections_bookmark" },
	{ id: "manager", label: "Game Manager", icon: "monitor_heart" },
	{
		id: "run",
		label: "Run",
		icon: "rocket_launch",
		paletteLabel: "Go to Launch Configuration"
	},
	{
		id: "versions",
		label: "Versions",
		icon: protonIcon,
		paletteLabel: "Go to Proton Versions",
		isCustomIcon: true,
		feature: "proton"
	},
	{
		id: "prefix",
		label: "Prefix",
		icon: "folder_shared",
		paletteLabel: "Go to WINE Prefixes",
		feature: "prefix"
	},
	{
		id: "utils",
		label: "Utils",
		icon: "handyman",
		paletteLabel: "Go to Utilities"
	},
	{ id: "plugins", label: "Plugins", icon: "extension" },
	{
		id: "settings",
		label: "Settings",
		icon: "settings",
		paletteLabel: "Go to Appearance & Settings"
	},
	{
		id: "editlsfg",
		label: "Edit LSFG",
		icon: "tune",
		feature: "lsfg",
		hidden: true
	},
	{ id: "notifications", label: "Notifications", icon: "notifications", bottom: true }
];

/** Ordered registry of every page in the app. */
export const routes: readonly AppRoute[] = routeDefinitions.map((route, order) => ({
	...route,
	order
}));

const routesById = new Map(routes.map((route) => [route.id, route]));

export function getRoute(id: string): AppRoute | undefined {
	return routesById.get(id);
}

/** Sort position of a route; unknown pages sort first. */
export function routeOrder(id: string): number {
	return getRoute(id)?.order ?? 0;
}
