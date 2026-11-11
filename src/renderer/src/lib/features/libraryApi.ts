import {
	toGame,
	toGameDetail,
	toGameUpdate,
	toIncomingList,
	toLibraryListPage,
	toLibraryOverview,
	toPreview,
	toSearchResponse,
	toUpdateHistory,
	toUpdatesSummary,
	toWireListRequest
} from "./libraryMap";
import { rpc } from "./rpc";
import type {
	Game,
	GamePreview,
	IncomingListResponse,
	LibraryEditableFields,
	LibraryGetResponse,
	LibraryListPage,
	LibraryOverview,
	LibraryPageRequest,
	LibraryQuery,
	SearchResponse,
	UpdatesResponse
} from "./types";
import type {
	WireCheckResult,
	WireGame,
	WireGameDetail,
	WireIncomingReport,
	WireLibraryList,
	WireLibraryOverview,
	WirePreview,
	WireSearchResult,
	WireUpdateCheckRequest,
	WireUpdateHistoryEntry,
	WireUpdatesSummary
} from "./wire";

/** Games per update page. The server caps one list at 1000. */
const UPDATES_PAGE_SIZE = 200;

/** Every game, with the updates filter on. */
const UPDATES_QUERY: LibraryQuery = {
	source: "",
	search: "",
	tags: [],
	favorite: false,
	installed: false,
	updates: true,
	sort: "added"
};

async function listPage(query: LibraryQuery, page: LibraryPageRequest): Promise<LibraryListPage> {
	const wire = await rpc<WireLibraryList>("LibraryList", toWireListRequest(query, page));
	return toLibraryListPage(wire);
}

/** Reads every game with a newer version, one page at a time. */
async function fetchUpdateGames(): Promise<Game[]> {
	const games: Game[] = [];
	for (let offset = 0; ; offset += UPDATES_PAGE_SIZE) {
		const page = await listPage(UPDATES_QUERY, {
			offset,
			limit: UPDATES_PAGE_SIZE,
			wantTotal: false
		});
		games.push(...page.items);
		if (page.items.length < UPDATES_PAGE_SIZE) return games;
	}
}

async function fetchUpdates(): Promise<UpdatesResponse> {
	const summary = await rpc<WireUpdatesSummary>("LibraryUpdatesSummary");
	return toUpdatesSummary(summary, await fetchUpdateGames());
}

export const libraryApi = {
	/** One page of games for a source, search and filters. Ask for the total on the first page. */
	list: listPage,

	/** Counts, tags and stats for the whole library. Not sent per page. */
	overview: async (): Promise<LibraryOverview> =>
		toLibraryOverview(await rpc<WireLibraryOverview>("LibraryOverview")),

	/** Starts an import from a plugin source. Progress arrives as library.sync events. */
	syncSource: async (source: string): Promise<void> => {
		await rpc<{ started: boolean }>("LibrarySyncSource", { source });
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
