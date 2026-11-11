import {
	toGame,
	toGameDetail,
	toGameUpdate,
	toIncomingList,
	toLibraryListing,
	toPreview,
	toSearchResponse,
	toUpdateHistory,
	toUpdatesSummary
} from "./libraryMap";
import { rpc } from "./rpc";
import type {
	Game,
	GamePreview,
	IncomingListResponse,
	LibraryEditableFields,
	LibraryGetResponse,
	LibraryListResponse,
	SearchResponse,
	UpdateHistoryEntry,
	UpdatesResponse
} from "./types";
import type {
	WireCheckResult,
	WireGame,
	WireGameDetail,
	WireIncomingReport,
	WireLibraryPage,
	WirePreview,
	WireSearchResult,
	WireUpdateCheckRequest,
	WireUpdateHistoryEntry,
	WireUpdatesSummary
} from "./wire";

/** The Go side returns at most this many games per LibraryList call. */
const PAGE_SIZE = 1000;

/** Loads every game. Each page also carries the library counts, which are the same on every page. */
async function fetchLibrary(): Promise<{ games: Game[]; counts: WireLibraryPage }> {
	const games: Game[] = [];
	for (let offset = 0; ; offset += PAGE_SIZE) {
		const page = await rpc<WireLibraryPage>("LibraryList", { limit: PAGE_SIZE, offset });
		const items = (page.items ?? []).map(toGame);
		games.push(...items);
		if (items.length === 0 || games.length >= page.total) return { games, counts: page };
	}
}

async function fetchUpdates(): Promise<UpdatesResponse> {
	const summary = await rpc<WireUpdatesSummary>("LibraryUpdatesSummary");
	const { games } = await fetchLibrary();
	return toUpdatesSummary(summary, games);
}

export const libraryApi = {
	list: async (): Promise<LibraryListResponse> => {
		const { games, counts } = await fetchLibrary();
		return toLibraryListing(games, counts);
	},

	get: async (id: number): Promise<LibraryGetResponse> =>
		toGameDetail(await rpc<WireGameDetail>("LibraryGet", { id })),

	update: async (id: number, fields: LibraryEditableFields): Promise<Game> =>
		toGame(await rpc<WireGame>("LibraryUpdate", toGameUpdate(id, fields))),

	/** Removes the game from the library. Files on disk are kept. */
	remove: (id: number) => rpc<void>("LibraryRemove", { id }),

	search: async (query: string): Promise<SearchResponse> =>
		toSearchResponse(await rpc<WireSearchResult[]>("LibrarySearch", { query })),

	preview: async (url: string): Promise<GamePreview> =>
		toPreview(await rpc<WirePreview>("LibraryPreview", { url })),

	addFromUrl: async (url: string): Promise<Game> =>
		toGame(await rpc<WireGame>("LibraryAddFromUrl", { url })),

	/** Links a local game to a source game page. */
	linkSource: async (id: number, url: string): Promise<Game> =>
		toGame(await rpc<WireGame>("LibraryLinkLocal", { id, url })),

	/** Points a game at an existing install folder. */
	linkFolder: async (id: number, path: string): Promise<Game> =>
		toGame(await rpc<WireGame>("LibraryLinkFolder", { id, path })),

	/** Checks every game for a newer version, ignoring the recheck interval, then returns the updates. */
	checkUpdates: async (): Promise<UpdatesResponse> => {
		const request: WireUpdateCheckRequest = { force: true };
		await rpc<WireCheckResult[]>("LibraryCheckUpdates", request);
		return fetchUpdates();
	},

	/** The unseen update count, the games with a newer version, and the last check. */
	updates: fetchUpdates,

	/** Recent checks, newest first. Pass before as the last seen id to get the next page. */
	updateHistory: async (params: { gameId?: number; limit?: number; before?: number } = {}) =>
		toUpdateHistory(
			await rpc<WireUpdateHistoryEntry[]>("LibraryUpdateHistory", {
				gameId: params.gameId ?? 0,
				limit: params.limit ?? 50,
				before: params.before ?? 0
			})
		),

	setFavorite: async (id: number, favorite: boolean): Promise<Game> =>
		toGame(await rpc<WireGame>("LibrarySetFavorite", { id, favorite })),

	launch: (id: number) => rpc<void>("LibraryLaunch", { id }),

	incomingList: async (): Promise<IncomingListResponse> =>
		toIncomingList(await rpc<WireIncomingReport>("IncomingList")),

	incomingLink: (name: string, gameId: number) => rpc<void>("IncomingLink", { name, gameId }),

	incomingIgnore: (name: string) => rpc<void>("IncomingIgnore", { name }),

	incomingCreateLocal: async (name: string): Promise<Game> =>
		toGame(await rpc<WireGame>("IncomingCreateLocal", { name }))
};
