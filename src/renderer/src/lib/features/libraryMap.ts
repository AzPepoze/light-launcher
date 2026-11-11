import type {
	Game,
	GameChange,
	GameFile,
	GamePreview,
	IncomingItem,
	IncomingListResponse,
	IncomingSuggestion,
	LibraryEditableFields,
	LibraryGetResponse,
	LibraryListResponse,
	LibraryStats,
	PreviewFile,
	SearchResponse,
	UpdatesResponse
} from "./types";
import type {
	WireCheck,
	WireFileCandidate,
	WireGame,
	WireGameDetail,
	WireGameDTO,
	WireGameFile,
	WireGameUpdate,
	WireIncomingItem,
	WireIncomingReport,
	WireLibraryPage,
	WireLibraryStats,
	WirePreview,
	WireSearchResult,
	WireSuggestion,
	WireUpdatesSummary
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
		coverLocal: wire.coverLocal,
		screenshots: wire.screenshots ?? [],
		screenshotLocals: wire.screenshotsLocal ?? [],
		description: wire.descriptionText,
		running: wire.running,
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

function toStats(wire: WireLibraryStats): LibraryStats {
	return {
		total: wire.total,
		installed: wire.installed,
		updates: wire.updates,
		bytes: wire.bytes,
		playtime: wire.playtime
	};
}

/** Builds the list view from the games of one or more pages. Counts come from the API, not from the games. */
export function toLibraryListing(games: Game[], counts: WireLibraryPage): LibraryListResponse {
	return {
		items: games,
		stats: toStats(counts.stats),
		sources: (counts.sources ?? []).map((source) => ({
			source: source.id,
			label: source.label,
			count: source.count
		})),
		tags: (counts.tags ?? []).map((tag) => ({ tag: tag.tag, count: tag.count }))
	};
}

function toGameFile(wire: WireGameFile): GameFile {
	return {
		id: wire.id,
		label: wire.label,
		url: wire.url,
		filename: wire.filename,
		size: wire.size,
		host: wire.host,
		kind: wire.kind,
		platform: wire.platform,
		requiresAuth: wire.requiresAuth === 1
	};
}

function toGameChange(wire: WireCheck): GameChange {
	return {
		id: wire.id,
		checkedAt: wire.checkedAt,
		from: wire.oldVersion,
		to: wire.newVersion,
		changelog: wire.changelog,
		seen: wire.seen === 1,
		applied: wire.applied === 1
	};
}

export function toGameDetail(wire: WireGameDetail): LibraryGetResponse {
	return {
		game: toGame(wire),
		files: (wire.files ?? []).map(toGameFile),
		changes: (wire.changes ?? []).map(toGameChange)
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
			source: result.source,
			score: game.score
		}))
	);
	return { hits, errors: toSearchErrors(results) };
}

function toPreviewFile(wire: WireFileCandidate): PreviewFile {
	return {
		id: wire.id,
		label: wire.label,
		url: wire.url,
		filename: orNull(wire.filename),
		size: wire.size || null,
		host: orNull(wire.host),
		kind: wire.kind,
		platform: wire.platform || "unknown",
		requiresAuth: wire.requiresAuth ?? false
	};
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
		existing: wire.existing,
		files: (wire.files ?? []).map(toPreviewFile),
		sourceLabel: wire.sourceLabel
	};
}

/** Turns the updates summary into the updates view. Games with a newer version come from the given list. */
export function toUpdatesSummary(summary: WireUpdatesSummary, games: Game[]): UpdatesResponse {
	const updatable = games.filter((game) => game.updateAvailable);
	const last = summary.last;
	return {
		count: summary.unseen,
		games: updatable,
		history: [],
		errors: (last?.results ?? [])
			.filter((result) => result.error)
			.map((result) => `${result.title}: ${result.error}`),
		lastCheckedAt: last?.checkedAt ?? null
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
