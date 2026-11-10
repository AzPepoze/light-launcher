import { errorMessage } from "../../lib/features/rpc";
import { libraryApi } from "../../lib/features/libraryApi";
import type { Game, LibraryStats } from "../../lib/features/types";
import {
	DEFAULT_FILTERS,
	hasActiveFilters,
	visibleGames,
	type LibraryFilters
} from "./libraryFilters";
import { loadSavedFilters, storeSavedFilters, type SavedFilter } from "./savedFilters";

const EMPTY_STATS: LibraryStats = { total: 0, installed: 0, updates: 0, bytes: 0, playtime: 0 };

export class LibraryState {
	games = $state<Game[]>([]);
	stats = $state<LibraryStats>(EMPTY_STATS);
	sources = $state<{ source: string; count: number }[]>([]);
	tags = $state<{ tag: string; count: number }[]>([]);
	loading = $state(true);
	error = $state("");
	filters = $state<LibraryFilters>({ ...DEFAULT_FILTERS, tags: [] });
	savedFilters = $state<SavedFilter[]>(loadSavedFilters());

	visible = $derived(visibleGames(this.games, this.filters));
	filtersActive = $derived(hasActiveFilters(this.filters));

	async load() {
		this.loading = true;
		this.error = "";
		try {
			const response = await libraryApi.list();
			this.games = response.items;
			this.stats = response.stats;
			this.sources = response.sources;
			this.tags = response.tags;
		} catch (err) {
			this.error = errorMessage(err);
		} finally {
			this.loading = false;
		}
	}

	gameById(id: number): Game | undefined {
		return this.games.find((game) => game.id === id);
	}

	replaceGame(game: Game) {
		this.games = this.games.map((existing) => (existing.id === game.id ? game : existing));
	}

	addGame(game: Game) {
		this.games = [game, ...this.games.filter((existing) => existing.id !== game.id)];
	}

	removeGame(id: number) {
		this.games = this.games.filter((game) => game.id !== id);
	}

	resetFilters() {
		this.filters = { ...DEFAULT_FILTERS, tags: [], sort: this.filters.sort };
	}

	saveCurrentFilter(name: string) {
		const trimmed = name.trim();
		if (!trimmed) return;
		const entry: SavedFilter = {
			id: `${Date.now()}`,
			name: trimmed,
			filters: { ...this.filters, tags: [...this.filters.tags] }
		};
		this.savedFilters = [...this.savedFilters, entry];
		storeSavedFilters(this.savedFilters);
	}

	applySavedFilter(entry: SavedFilter) {
		this.filters = { ...entry.filters, tags: [...entry.filters.tags] };
	}

	deleteSavedFilter(id: string) {
		this.savedFilters = this.savedFilters.filter((entry) => entry.id !== id);
		storeSavedFilters(this.savedFilters);
	}
}
