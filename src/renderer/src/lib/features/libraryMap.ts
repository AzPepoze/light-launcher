import type {
	Game,
	GamePreview,
	IncomingItem,
	IncomingListResponse,
	IncomingSuggestion,
	LibraryEditableFields,
	LibraryListResponse,
	SearchResponse,
	UpdatesResponse
} from "./types";
import type {
	WireCheckResult,
	WireGame,
	WireGameDTO,
	WireGameUpdate,
	WireIncomingItem,
	WireIncomingReport,
	WirePreview,
	WireSearchResult,
	WireSuggestion
} from "./wire";

const HTML_ENTITIES: Record<string, string> = {
	"&amp;": "&",
	"&lt;": "<",
	"&gt;": ">",
	"&quot;": '"',
	"&#39;": "'",
	"&nbsp;": " "
};

/** Turns source descriptions (HTML) into plain text with line breaks. */
export function htmlToText(html: string | null | undefined): string | null {
	if (!html) return null;
	const text = html
		.replace(/<br\s*\/?>/gi, "\n")
		.replace(/<\/p>/gi, "\n\n")
		.replace(/<[^>]*>/g, "")
		.replace(/&(amp|lt|gt|quot|#39|nbsp);/g, (entity) => HTML_ENTITIES[entity])
		.trim();
	return text || null;
}

function orNull(value: string | null | undefined): string | null {
	return value || null;
}

/** Same rule as the Go "updates" filter: a stored version that differs from the installed one. */
export function hasNewVersion(version: string | null, installedVersion: string | null): boolean {
	return Boolean(installedVersion) && Boolean(version) && version !== installedVersion;
}

export function toGame(wire: WireGame): Game {
	return {
		id: wire.id,
		source: wire.source,
		url: wire.url,
		title: wire.title,
		version: wire.version,
		developer: wire.developer,
		engine: wire.engine,
		tags: wire.tags ?? [],
		cover: wire.coverUrl,
		screenshots: wire.screenshots ?? [],
		description: htmlToText(wire.description),
		favorite: wire.favorite === 1,
		notes: wire.notes,
		installed: Boolean(wire.installDir),
		installedVersion: wire.installedVersion,
		installDir: wire.installDir,
		installSize: wire.installSize,
		exePath: wire.exePath,
		exeArgs: wire.exeArgs,
		playtimeS: wire.playtimeS,
		lastPlayedAt: wire.lastPlayedAt,
		addedAt: wire.addedAt,
		updateAvailable: hasNewVersion(wire.version, wire.installedVersion)
	};
}

function countBy<T>(items: T[], keyOf: (item: T) => string): Map<string, number> {
	const counts = new Map<string, number>();
	for (const item of items) {
		const key = keyOf(item);
		counts.set(key, (counts.get(key) ?? 0) + 1);
	}
	return counts;
}

/** Builds the header stats, source filter and tag filter from the full game list. */
export function buildLibrarySnapshot(games: Game[]): LibraryListResponse {
	const sourceCounts = countBy(games, (game) => game.source);
	const tagCounts = countBy(
		games.flatMap((game) => game.tags),
		(tag) => tag
	);
	return {
		items: games,
		stats: {
			total: games.length,
			installed: games.filter((game) => game.installed).length,
			updates: games.filter((game) => game.updateAvailable).length,
			bytes: games.reduce((sum, game) => sum + (game.installSize ?? 0), 0),
			playtime: games.reduce((sum, game) => sum + game.playtimeS, 0)
		},
		sources: [...sourceCounts]
			.map(([source, count]) => ({ source, count }))
			.sort((a, b) => a.source.localeCompare(b.source)),
		tags: [...tagCounts]
			.map(([tag, count]) => ({ tag, count }))
			.sort((a, b) => b.count - a.count || a.tag.localeCompare(b.tag))
	};
}

export function toGameUpdate(id: number, fields: LibraryEditableFields): WireGameUpdate {
	return { id, ...fields };
}

function toSearchErrors(results: WireSearchResult[]): string[] {
	return results
		.filter((result) => result.error)
		.map((result) => `${result.source}: ${result.error}`);
}

export function toSearchResponse(results: WireSearchResult[]): SearchResponse {
	const hits = results.flatMap((result) =>
		(result.games ?? []).map((game) => ({
			url: game.url,
			title: game.title,
			source: result.source
		}))
	);
	return { hits, errors: toSearchErrors(results) };
}

export function toPreview(wire: WirePreview): GamePreview {
	const game: WireGameDTO = wire.game;
	return {
		source: game.source,
		title: game.title,
		version: orNull(game.version),
		developer: orNull(game.developer),
		engine: orNull(game.engine),
		tags: game.tags ?? [],
		description: htmlToText(game.descriptionHtml),
		cover: orNull(game.coverUrl),
		url: game.url,
		existing: wire.existing
	};
}

export function toUpdatesResponse(results: WireCheckResult[], games: Game[]): UpdatesResponse {
	const updatable = games.filter((game) => game.updateAvailable);
	return {
		count: updatable.length,
		games: updatable,
		history: [],
		errors: results
			.filter((result) => result.error)
			.map((result) => `${result.title}: ${result.error}`)
	};
}

function toSuggestion(wire: WireSuggestion): IncomingSuggestion {
	return {
		gameId: wire.gameId,
		title: wire.title,
		cover: wire.cover,
		version: wire.version,
		installed: wire.installed,
		confidence: wire.confidence
	};
}

function toIncomingItem(wire: WireIncomingItem): IncomingItem {
	return {
		name: wire.name,
		size: wire.size,
		mtime: wire.mtime,
		ready: wire.ready,
		ignored: wire.ignored,
		auto: wire.auto,
		suggestions: (wire.suggestions ?? []).map(toSuggestion)
	};
}

export function toIncomingList(wire: WireIncomingReport): IncomingListResponse {
	return {
		configured: wire.configured,
		folder: wire.folder,
		items: (wire.items ?? []).map(toIncomingItem),
		count: wire.count,
		warning: wire.warning
	};
}
