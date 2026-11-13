import type { TabItem } from "../../components/shared/tabs";
import { ago } from "../../lib/features/format";
import type { CoverSource } from "../../lib/features/mediaCache";
import type {
	Game,
	GameChange,
	GameFile,
	LibrarySource,
	UpdateHistoryEntry
} from "../../lib/features/types";

/** The source tab that keeps every source. */
export const ALL_SOURCES = "";

/* Local shapes of the library.updateCheck event and history cover fields. The page status is the canonical type in lib/features/types.ts. */
export const UPDATE_CHECK_EVENT = "library.updateCheck";

export type UpdateCheckState = "running" | "done";

/** What the page shows on open. Reading it never starts a check. */
export type { UpdateStatus } from "../../lib/features/types";

export type UpdateCheckOutcome = "update" | "current" | "error" | "skipped";

export interface UpdateCheckResult {
	gameId: number;
	title: string;
	from: string | null;
	to: string | null;
	result: UpdateCheckOutcome;
	error: string | null;
	coverPath: string | null;
	coverUrl: string | null;
}

/** A source that could not be checked as a whole, such as a missing API key. */
export interface UpdateIssue {
	source: string;
	label: string;
	kind: "auth" | "error";
	message: string;
	games: number;
}

export interface UpdateCheckEvent {
	state: UpdateCheckState;
	done: number;
	total: number;
	skipped: number;
	issues: UpdateIssue[];
	current: { gameId: number; title: string } | null;
	result: UpdateCheckResult | null;
}

export interface UpdateHistoryItem extends Omit<UpdateHistoryEntry, "result"> {
	result: UpdateCheckOutcome;
	coverPath: string | null;
	coverUrl: string | null;
}

export type HistoryFilter = "attention" | "all" | "update" | "error" | "current";

export interface RunCounts {
	updates: number;
	failed: number;
	upToDate: number;
	skipped: number;
}

export interface UpdateGroup {
	source: string;
	label: string;
	games: Game[];
}

export interface VersionPair {
	from: string;
	to: string;
}

export interface HistoryStep {
	kind: "checked" | "updated" | "failed" | "skipped";
	icon: string;
	tone: "success" | "accent" | "danger" | "muted";
	title: string;
}

export type UpdateFileChoice =
	| { kind: "single"; file: GameFile }
	| { kind: "none" }
	| { kind: "many" };

function count(value: number, one: string, many: string): string {
	return `${value} ${value === 1 ? one : many}`;
}

export function gamesWithUpdates(total: number): string {
	return `${count(total, "game has", "games have")} a newer release`;
}

/** The time of the last saved check, or a note that none ran yet. */
export function lastCheckedLabel(lastCheckedAt: string | null, now: number = Date.now()): string {
	if (!lastCheckedAt) return "Never checked";
	return `Last checked ${ago(lastCheckedAt, now)}`;
}

/**
 * When the next background check is due. Empty when none is scheduled or the time cannot be read.
 * A time at or before now means a scheduler tick is due, so the wait is shown in whole units and rounded up.
 */
export function nextCheckLabel(nextRunAt: string | null, now: number = Date.now()): string {
	if (!nextRunAt) return "";
	const time = Date.parse(nextRunAt);
	if (Number.isNaN(time)) return "";
	const wait = time - now;
	if (wait <= 0) return "Next check soon";
	if (wait < 60_000) return "Next check in under a minute";
	const minutes = Math.ceil(wait / 60_000);
	if (minutes < 60) return `Next check in ${minutes}m`;
	const hours = Math.ceil(wait / 3_600_000);
	if (hours < 24) return `Next check in ${hours}h`;
	return `Next check in ${Math.ceil(wait / 86_400_000)}d`;
}

/** The id to pass as `before` to load the next page: the oldest saved row. Zero starts from the top. */
export function pagingCursor(entries: UpdateHistoryItem[]): number {
	const saved = entries.filter((entry) => entry.id > 0).map((entry) => entry.id);
	return saved.length > 0 ? Math.min(...saved) : 0;
}

/** Progress as 0 to 1, or null before the first event says how many games there are. */
export function progressValue(done: number, total: number): number | null {
	return total > 0 ? done / total : null;
}

export function progressLabel(done: number, total: number): string {
	return total > 0 ? `Checking ${done} of ${total}` : "Checking for updates";
}

export function checkedSummary(total: number): string {
	return `Checked ${count(total, "game", "games")}`;
}

/** A quiet line for games the check left out. Empty when none were skipped. */
export function skippedNote(skipped: number, issues: UpdateIssue[]): string {
	if (skipped === 0) return "";
	const needsKey = issues.find((issue) => issue.kind === "auth");
	const reason = needsKey ? ` (${needsKey.label} needs an API key)` : "";
	return `Skipped ${count(skipped, "game", "games")}${reason}`;
}

/** The run result for one notification. */
export function runMessage(total: number, counts: RunCounts): string {
	const updates = count(counts.updates, "update", "updates");
	const failed = counts.failed > 0 ? `, ${counts.failed} failed` : "";
	return `Checked ${count(total, "game", "games")}, ${updates}${failed}`;
}

/** One notification per source issue. Setup problems are warnings, other failures are errors. */
export function issueMessage(issue: UpdateIssue): string {
	const skipped = `${count(issue.games, "game was", "games were")} skipped`;
	if (issue.kind === "auth") return `${issue.label} needs an API key. ${skipped}.`;
	return `${issue.label}: ${plainMessage(issue.message) ?? "check failed"}. ${skipped}.`;
}

export function issueLevel(issue: UpdateIssue): "warning" | "error" {
	return issue.kind === "auth" ? "warning" : "error";
}

/** Removes error codes such as "(code -32603)" so a message reads in plain words. Null when nothing is left. */
export function plainMessage(message: string | null): string | null {
	if (!message) return null;
	const plain = message.replace(/\s*\((?:code|error)\s*-?\d+\)/gi, "").trim();
	return plain || null;
}

export function emptyCounts(): RunCounts {
	return { updates: 0, failed: 0, upToDate: 0, skipped: 0 };
}

/** Adds one checked game to the counts of a run. */
export function tally(counts: RunCounts, result: UpdateCheckResult): RunCounts {
	if (result.result === "update") return { ...counts, updates: counts.updates + 1 };
	if (result.result === "error") return { ...counts, failed: counts.failed + 1 };
	if (result.result === "skipped") return { ...counts, skipped: counts.skipped + 1 };
	return { ...counts, upToDate: counts.upToDate + 1 };
}

/** The result the backend history filter takes. Empty means every check. */
export function backendResult(filter: HistoryFilter): "" | "update" | "current" | "error" {
	if (filter === "update" || filter === "error" || filter === "current") return filter;
	return "";
}

/** Skipped games are not failures, so they only show under All. */
export function matchesFilter(filter: HistoryFilter, result: string): boolean {
	if (filter === "all") return true;
	if (filter === "attention") return result === "update" || result === "error";
	return result === filter;
}

/** Keeps only the newest up-to-date check of each game. Entries stay in their order. */
export function collapseUpToDate(entries: UpdateHistoryItem[]): UpdateHistoryItem[] {
	const seen = new Set<number>();
	return entries.filter((entry) => {
		if (entry.result !== "current") return true;
		if (seen.has(entry.gameId)) return false;
		seen.add(entry.gameId);
		return true;
	});
}

/** The rows the history shows for a filter. The All filter folds repeated up-to-date checks. */
export function visibleHistory(
	entries: UpdateHistoryItem[],
	filter: HistoryFilter
): UpdateHistoryItem[] {
	const matching = entries.filter((entry) => matchesFilter(filter, entry.result));
	return filter === "all" ? collapseUpToDate(matching) : matching;
}

/** The newest check time of each game, from the history rows. */
export function latestCheckByGame(entries: UpdateHistoryItem[]): Map<number, string> {
	const latest = new Map<number, string>();
	for (const entry of entries) {
		const known = latest.get(entry.gameId);
		if (!known || checkedTime(entry.checkedAt) > checkedTime(known))
			latest.set(entry.gameId, entry.checkedAt);
	}
	return latest;
}

function checkedTime(iso: string): number {
	return Date.parse(iso) || 0;
}

/** Groups games by source, in label order. Inside a group the most recent check comes first. */
export function groupBySource(
	games: Game[],
	labels: Record<string, string>,
	lastChecked: Map<number, string>
): UpdateGroup[] {
	const bySource = new Map<string, Game[]>();
	for (const game of games) {
		const list = bySource.get(game.source) ?? [];
		list.push(game);
		bySource.set(game.source, list);
	}
	return [...bySource.entries()]
		.map(([source, list]) => ({
			source,
			label: labels[source] ?? source,
			games: [...list].sort(
				(a, b) =>
					checkedTime(lastChecked.get(b.id) ?? "") - checkedTime(lastChecked.get(a.id) ?? "") ||
					a.title.localeCompare(b.title)
			)
		}))
		.sort((a, b) => a.label.localeCompare(b.label));
}

/** The games of one source. The all-sources tab keeps every game. */
export function gamesOfSource(games: Game[], source: string): Game[] {
	return source === ALL_SOURCES ? games : games.filter((game) => game.source === source);
}

/** The source tabs: All first, then each source that can check for updates, by label. Badges count the games with updates. */
export function updateSourceTabs(
	sources: LibrarySource[],
	updateSourceIds: string[],
	games: Game[]
): TabItem[] {
	const checkable = new Set(updateSourceIds);
	const sourceTabs = sources
		.filter((source) => checkable.has(source.source))
		.map((source) => ({
			id: source.source,
			label: source.label,
			badge: gamesOfSource(games, source.source).length
		}))
		.sort((a, b) => a.label.localeCompare(b.label));
	return [{ id: ALL_SOURCES, label: "All", badge: games.length }, ...sourceTabs];
}

/** The history rows of the games in a source. Null keeps every row. */
export function historyOfSource(
	entries: UpdateHistoryItem[],
	sourceGameIds: ReadonlySet<number> | null
): UpdateHistoryItem[] {
	if (!sourceGameIds) return entries;
	return entries.filter((entry) => sourceGameIds.has(entry.gameId));
}

export function versionPair(game: Game): VersionPair {
	return { from: game.installedVersion ?? "unknown", to: game.version ?? "new" };
}

/** The changelog of the newest check that has one, or null. */
export function newestChangelog(changes: GameChange[]): string | null {
	let newest: GameChange | null = null;
	for (const change of changes) {
		if (!change.changelog?.trim()) continue;
		if (!newest || checkedTime(change.checkedAt) > checkedTime(newest.checkedAt)) newest = change;
	}
	return newest?.changelog ?? null;
}

export function historyStep(entry: UpdateHistoryItem): HistoryStep {
	if (entry.result === "error") {
		return {
			kind: "failed",
			icon: "error",
			tone: "danger",
			title: plainMessage(entry.error) ?? "Check failed"
		};
	}
	if (entry.result === "skipped") {
		return {
			kind: "skipped",
			icon: "skip_next",
			tone: "muted",
			title: plainMessage(entry.error) ?? "Skipped"
		};
	}
	if (entry.result === "update") {
		return { kind: "updated", icon: "upgrade", tone: "accent", title: "Update found" };
	}
	return { kind: "checked", icon: "check_circle", tone: "success", title: "Up to date" };
}

/** A history row's cover. Only the remote URL is used: the on-disk path is not a media route. */
export function historyCover(entry: UpdateHistoryItem): CoverSource {
	return { id: entry.gameId, cover: entry.coverUrl, coverLocal: null };
}

/** A row for a check that just finished, shown before the history is reloaded. */
export function liveHistoryItem(
	result: UpdateCheckResult,
	id: number,
	checkedAt: string
): UpdateHistoryItem {
	return {
		id,
		gameId: result.gameId,
		title: result.title,
		checkedAt,
		from: result.from,
		to: result.to,
		result: result.result,
		error: result.error,
		coverPath: result.coverPath,
		coverUrl: result.coverUrl,
		source: ""
	};
}

/** A game with one saved file can be downloaded straight away. Otherwise the person picks it in Library. */
export function pickUpdateFile(files: GameFile[]): UpdateFileChoice {
	if (files.length === 0) return { kind: "none" };
	if (files.length === 1) return { kind: "single", file: files[0] };
	return { kind: "many" };
}
