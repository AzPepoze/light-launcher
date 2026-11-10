import type { Game } from "../../lib/features/types";

export type LibrarySort = "title" | "added" | "played";

export interface LibraryFilters {
	search: string;
	source: string;
	tags: string[];
	installed: boolean;
	updates: boolean;
	favorite: boolean;
	sort: LibrarySort;
}

export const DEFAULT_FILTERS: LibraryFilters = {
	search: "",
	source: "",
	tags: [],
	installed: false,
	updates: false,
	favorite: false,
	sort: "added"
};

export const SORT_LABELS: Record<LibrarySort, string> = {
	title: "Title",
	added: "Recently added",
	played: "Recently played"
};

function matchesSearch(game: Game, search: string): boolean {
	const query = search.trim().toLowerCase();
	if (!query) return true;
	const haystack = [game.title, game.developer ?? "", ...game.tags].join(" ").toLowerCase();
	return query.split(/\s+/).every((word) => haystack.includes(word));
}

function matchesTags(game: Game, tags: string[]): boolean {
	return tags.every((tag) => game.tags.includes(tag));
}

export function matchesFilters(game: Game, filters: LibraryFilters): boolean {
	if (!matchesSearch(game, filters.search)) return false;
	if (filters.source && game.source !== filters.source) return false;
	if (!matchesTags(game, filters.tags)) return false;
	if (filters.installed && !game.installed) return false;
	if (filters.updates && !game.updateAvailable) return false;
	if (filters.favorite && !game.favorite) return false;
	return true;
}

export function filterGames(games: Game[], filters: LibraryFilters): Game[] {
	return games.filter((game) => matchesFilters(game, filters));
}

function compareTitles(a: Game, b: Game): number {
	return a.title.localeCompare(b.title, undefined, { sensitivity: "base" });
}

function compareAdded(a: Game, b: Game): number {
	return b.addedAt.localeCompare(a.addedAt);
}

function comparePlayed(a: Game, b: Game): number {
	if (a.lastPlayedAt === b.lastPlayedAt) return 0;
	if (!a.lastPlayedAt) return 1;
	if (!b.lastPlayedAt) return -1;
	return b.lastPlayedAt.localeCompare(a.lastPlayedAt);
}

const COMPARATORS: Record<LibrarySort, (a: Game, b: Game) => number> = {
	title: compareTitles,
	added: compareAdded,
	played: comparePlayed
};

export function sortGames(games: Game[], sort: LibrarySort): Game[] {
	return [...games].sort(COMPARATORS[sort]);
}

export function visibleGames(games: Game[], filters: LibraryFilters): Game[] {
	return sortGames(filterGames(games, filters), filters.sort);
}

export function hasActiveFilters(filters: LibraryFilters): boolean {
	return (
		filters.search.trim() !== "" ||
		filters.source !== "" ||
		filters.tags.length > 0 ||
		filters.installed ||
		filters.updates ||
		filters.favorite
	);
}
