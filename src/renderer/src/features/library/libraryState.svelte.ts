import { createLogger } from "../../lib/logger";
import {
	LIBRARY_CHANGED_EVENT,
	LIBRARY_GAME_EVENT,
	LIBRARY_REMOVED_EVENT
} from "../../lib/features/events";
import { errorMessage, listen } from "../../lib/features/rpc";
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
const log = createLogger("Library");

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

	/** Keeps the list in step with library events. Returns a function that stops them. */
	subscribe(): () => void {
		const stops = [
			listen<{ id: number }>(LIBRARY_GAME_EVENT, (event) => this.refreshGame(event.id)),
			listen<{ id: number }>(LIBRARY_REMOVED_EVENT, (event) => this.removeGame(event.id)),
			listen<unknown>(LIBRARY_CHANGED_EVENT, () => this.load())
		];
		return () => stops.forEach((stop) => stop());
	}

	/** Reloads one game. A failed refresh keeps the old row; the next full load fixes it. */
	private async refreshGame(id: number) {
		try {
			const { game } = await libraryApi.get(id);
			if (this.gameById(id)) {
				this.replaceGame(game);
			} else {
				this.addGame(game);
			}
		} catch (err) {
			log.warn(`could not refresh game ${id}`, errorMessage(err));
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
