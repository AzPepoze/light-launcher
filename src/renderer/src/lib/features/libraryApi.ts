import {
	buildLibrarySnapshot,
	toGame,
	toGameUpdate,
	toIncomingList,
	toPreview,
	toSearchResponse,
	toUpdatesResponse
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
	UpdatesResponse
} from "./types";
import type {
	WireCheckResult,
	WireGame,
	WireIncomingReport,
	WireLibraryPage,
	WirePreview,
	WireSearchResult
} from "./wire";

/** The Go side returns at most this many games per LibraryList call. */
const PAGE_SIZE = 1000;

async function fetchAllGames(): Promise<Game[]> {
	const games: Game[] = [];
	for (let offset = 0; ; offset += PAGE_SIZE) {
		const page = await rpc<WireLibraryPage>("LibraryList", { limit: PAGE_SIZE, offset });
		const items = (page.items ?? []).map(toGame);
		games.push(...items);
		if (items.length === 0 || games.length >= page.total) return games;
	}
}

export const libraryApi = {
	/** Loads every game, then derives stats, sources and tags from them. */
	list: async (): Promise<LibraryListResponse> => buildLibrarySnapshot(await fetchAllGames()),

	/** Files and changes have no API yet, so they come back empty. */
	get: async (id: number): Promise<LibraryGetResponse> => ({
		game: toGame(await rpc<WireGame>("LibraryGet", { id })),
		files: [],
		changes: []
	}),

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

	/** Checks every game for a newer version, then returns the games that have one. */
	checkUpdates: async (): Promise<UpdatesResponse> => {
		const results = await rpc<WireCheckResult[]>("LibraryCheckUpdates", {});
		return toUpdatesResponse(results, await fetchAllGames());
	},

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
