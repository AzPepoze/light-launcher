import {
	toBrowsePage,
	toFilterOptions,
	toGame,
	toGameDetail,
	toGameUpdate,
	toIncomingList,
	toLibraryListPage,
	toLibraryOverview,
	toPreview,
	toSearchResponse,
	toUpdateHistory,
	toUpdateStatus,
	toUpdatesSummary,
	toWireBrowseRequest,
	toWireListRequest
} from "./libraryMap";
import { rpc } from "./rpc";
import type {
	BrowsePage,
	BrowseQuery,
	FilterOption,
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
	UpdateResult,
	UpdateStatus,
	UpdatesResponse
} from "./types";
import type {
	WireActionResult,
	WireBrowsePage,
	WireCheckResult,
	WireFiltersPage,
	WireGame,
	WireGameDetail,
	WireIncomingReport,
	WireLibraryList,
	WireLibraryOverview,
	WirePreview,
	WireSearchResult,
	WireUpdateCheckRequest,
	WireUpdateHistoryEntry,
	WireUpdateStatus,
	WireUpdatesSummary
} from "./wire";

/** Games per list page. The server caps one list at 1000. */
const GAMES_PAGE_SIZE = 200;

/** The host caps a filter list at 200. */
const FILTERS_LIMIT = 200;

/** Every game, with no filter. */
const ALL_GAMES_QUERY: LibraryQuery = {
	source: "",
	search: "",
	tags: [],
	favorite: false,
	installed: false,
	updates: false,
	sort: "added"
};

async function listPage(query: LibraryQuery, page: LibraryPageRequest): Promise<LibraryListPage> {
	const wire = await rpc<WireLibraryList>("LibraryList", toWireListRequest(query, page));
	return toLibraryListPage(wire);
}

/** Reads every game the query matches, one page at a time. */
async function fetchAllGames(query: LibraryQuery): Promise<Game[]> {
	const games: Game[] = [];
	for (let offset = 0; ; offset += GAMES_PAGE_SIZE) {
		const page = await listPage(query, {
			offset,
			limit: GAMES_PAGE_SIZE,
			wantTotal: false
		});
		games.push(...page.items);
		if (page.items.length < GAMES_PAGE_SIZE) return games;
	}
}

/** Reads every game with a newer version. */
function fetchUpdateGames(): Promise<Game[]> {
	return fetchAllGames({ ...ALL_GAMES_QUERY, updates: true });
}

async function fetchUpdates(): Promise<UpdatesResponse> {
	const summary = await rpc<WireUpdatesSummary>("LibraryUpdatesSummary");
	return toUpdatesSummary(summary, await fetchUpdateGames());
}

export const libraryApi = {
	/** One page of games for a source, search and filters. Ask for the total on the first page. */
	list: listPage,

	/** Every game of one source, read one page at a time. */
	sourceGames: (source: string): Promise<Game[]> => fetchAllGames({ ...ALL_GAMES_QUERY, source }),

	/** Counts, tags and stats for the whole library. Not sent per page. */
	overview: async (): Promise<LibraryOverview> =>
		toLibraryOverview(await rpc<WireLibraryOverview>("LibraryOverview")),

	/** One page of a source's games with its sidebar filters. Browse results are not saved. */
	browse: async (query: BrowseQuery): Promise<BrowsePage> =>
		toBrowsePage(await rpc<WireBrowsePage>("LibraryBrowse", toWireBrowseRequest(query))),

	/** The values of one dynamic sidebar group, such as the source's tags. */
	filters: async (source: string, group: string, query = ""): Promise<FilterOption[]> => {
		const page = await rpc<WireFiltersPage>("LibraryFilters", {
			source,
			group,
			query,
			limit: FILTERS_LIMIT
		});
		return toFilterOptions(page.options);
	},

	/** Starts an import from a plugin source. Progress arrives as library.sync events. */
	syncSource: async (source: string): Promise<void> => {
		await rpc<{ started: boolean }>("LibrarySyncSource", { source });
	},

	/** Runs one action a source declares. Omit gameId for a library action. Refresh the list when refresh is true. */
	runAction: (source: string, action: string, gameId: number | null = null): Promise<WireActionResult> =>
		rpc<WireActionResult>("LibraryRunAction", { source, action, gameId }),

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

	/** Recent checks, newest first. Pass before as the last seen id to get the next page. Pass result to keep one result. */
	updateHistory: async (
		params: {
			gameId?: number;
			limit?: number;
			before?: number;
			result?: UpdateResult | "";
			source?: string;
		} = {}
	) =>
		toUpdateHistory(
			await rpc<WireUpdateHistoryEntry[]>("LibraryUpdateHistory", {
				gameId: params.gameId ?? 0,
				limit: params.limit ?? 50,
				before: params.before ?? 0,
				result: params.result ?? "",
				source: params.source ?? ""
			})
		),

	/** The Updates page state. Starts no check: the last run, whether one runs, and the games with updates. */
	updateStatus: async (): Promise<UpdateStatus> =>
		toUpdateStatus(await rpc<WireUpdateStatus>("LibraryUpdateStatus")),

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
