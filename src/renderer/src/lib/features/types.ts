// View types for the library, downloads, updates and plugin pages.
// The Go payloads are in wire.ts. The mappers in *Map.ts turn them into these types.

import type { LaunchOptions } from "@shared";

export interface Game {
	id: number;
	source: string;
	url: string;
	title: string;
	version: string | null;
	developer: string | null;
	engine: string | null;
	tags: string[];
	/** Remote cover image. */
	cover: string | null;
	/** Cover file cached by the launcher, relative to the media folder. */
	coverLocal: string | null;
	screenshots: string[];
	/** Cached screenshot files, relative to the media folder. Null where a screenshot is not cached. */
	screenshotLocals: (string | null)[];
	/** Plain text, with tags and scripts removed. */
	description: string | null;
	/** True while the game is running. */
	running: boolean;
	favorite: boolean;
	notes: string | null;
	installed: boolean;
	installedVersion: string | null;
	installDir: string | null;
	installSize: number | null;
	exePath: string | null;
	exeArgs: string | null;
	playtimeS: number;
	lastPlayedAt: string | null;
	addedAt: string;
	updateAvailable: boolean;
	/** How the row was found: plugin, hand, scan, watch or linked. */
	origin: string;
	/** local for games from folders and hand adding, plugin for imported source games. */
	sourceKind: "local" | "plugin";
	/** The game folder of a local row. */
	folderPath: string | null;
	/** The watched folder a scanned row came from. */
	scanFolder: string | null;
	/** The launch settings of a local row. Plugin rows have null. */
	localConfig: LaunchOptions | null;
	/** True when the row must survive scans and removals. */
	pinned: boolean;
}

export interface LibraryStats {
	total: number;
	installed: number;
	updates: number;
	bytes: number;
	playtime: number;
}

/** The sort names the library list accepts. The server names are in libraryMap.ts. */
export type LibrarySort = "title" | "added" | "lastPlayed";

/** The shape of a source's covers: 3:4, 16:9 or 1:1. */
export type SourceTile = "poster" | "wide" | "square";

/** The filter groups a plugin source can put in the Library sidebar. */
export type SidebarGroupType = "search" | "toggles" | "select" | "multi" | "sort";

export interface SidebarSort {
	value: string;
	order: "asc" | "desc";
}

export interface SidebarOption {
	value: string;
	label: string;
	icon: string | null;
	/** The number of games the source counts for this value, when it says. */
	count: number | null;
}

export interface SidebarGroup {
	type: SidebarGroupType;
	id: string;
	label: string;
	placeholder: string;
	options: SidebarOption[];
	/** True when the options come from LibraryFilters instead of the manifest. */
	dynamic: boolean;
	/** True when a dynamic group can be narrowed with its own search box. */
	searchable: boolean;
	/** True when a sort group can be reversed. */
	orderable: boolean;
	default: SidebarSort | null;
}

/** A button a source adds to the Library. Library actions show in the sidebar, game actions in the game detail. */
export interface SourceAction {
	id: string;
	label: string;
	/** Material icon name. */
	icon?: string;
	method: string;
	scope: "library" | "game";
}

export type SourceBadge = "favorite" | "updateAvailable" | "installed";

/** The game fields a card shows. Missing fields use the defaults: title and version, no badges. */
export interface SourceCard {
	title?: string;
	subtitle?: string;
	badges?: SourceBadge[];
}

export interface LibrarySource {
	source: string;
	label: string;
	kind: "local" | "plugin";
	plugin: string | null;
	/** Source features the UI may show: browse, sync:*, updateCheck, auth and so on. */
	capabilities: string[];
	/** True when the source can import or sync the user's games. */
	sync: boolean;
	/** True when the source can look for newer game versions. */
	updateCheck: boolean;
	count: number;
	/** Older card shape name, kept for consumers that predate cover. */
	tile: SourceTile;
	/** Card box as "W:H", such as "3:4". */
	cover: string;
	actions: SourceAction[];
	/** Null when the source names no card fields. */
	card: SourceCard | null;
	/** Empty for local games and sources that declare no filters. */
	sidebar: SidebarGroup[];
}

/** Counts and the tag list for the whole library. Loaded once, not per page. */
export interface LibraryOverview {
	sources: LibrarySource[];
	tags: { tag: string; count: number }[];
	stats: LibraryStats;
}

/** What one list shows. An empty source means every source. */
export interface LibraryQuery {
	source: string;
	search: string;
	tags: string[];
	favorite: boolean;
	installed: boolean;
	updates: boolean;
	sort: LibrarySort;
}

export interface LibraryPageRequest {
	offset: number;
	limit: number;
	wantTotal: boolean;
}

/** A filter value the plugin protocol accepts: one choice, several, or a toggle. */
export type BrowseFilterValue = string | string[] | boolean;

/** One page request to a plugin source's browse method. Nothing is saved. */
export interface BrowseQuery {
	source: string;
	query: string;
	filters: Record<string, BrowseFilterValue>;
	sort: string;
	order: "asc" | "desc" | "";
	cursor: string;
	limit: number;
}

export interface BrowsePage {
	games: Game[];
	nextCursor: string | null;
	total: number | null;
}

/** One choice of a dynamic sidebar group, with the source's count when it sends one. */
export interface FilterOption {
	value: string;
	label: string;
	count: number | null;
}

export interface LibraryListPage {
	items: Game[];
	/** Sent for the first page or when asked for. Undefined otherwise. */
	total?: number;
}

export type LibrarySyncPhase = "running" | "done" | "error";

export interface LibrarySyncProgress {
	source: string;
	phase: LibrarySyncPhase;
	imported: number;
	skipped: number;
	failed: number;
	message?: string;
}

export interface GameFile {
	id: number;
	label: string;
	url: string;
	filename: string | null;
	size: number | null;
	host: string | null;
	kind: string;
	platform: string;
	requiresAuth: boolean;
}

export interface GameChange {
	id: number;
	checkedAt: string;
	from: string | null;
	to: string | null;
	changelog: string | null;
	seen: boolean;
	applied: boolean;
}

export interface LibraryGetResponse {
	game: Game;
	files: GameFile[];
	changes: GameChange[];
}

export type LibraryEditableFields = Partial<
	Pick<Game, "title" | "notes" | "exePath" | "exeArgs" | "favorite">
>;

export interface SearchHit {
	url: string;
	title: string;
	source: string;
	/** How well the title matches the query, about 0 to 2.5. */
	score: number;
}

export interface SearchResponse {
	hits: SearchHit[];
	/** One message per source that failed to search. */
	errors: string[];
}

export interface GamePreview {
	source: string;
	title: string;
	version: string | null;
	developer: string | null;
	engine: string | null;
	tags: string[];
	description: string | null;
	cover: string | null;
	url: string;
	existing: { id: number; title: string; installed: boolean } | null;
	/** Download files the source lists for the game. Empty when the source cannot list them. */
	files: PreviewFile[];
	sourceLabel: string;
}

export interface PreviewFile {
	id: string;
	label: string;
	url: string;
	filename: string | null;
	size: number | null;
	host: string | null;
	kind: string;
	platform: string;
	requiresAuth: boolean;
}

export interface IncomingSuggestion {
	gameId: number;
	title: string;
	cover: string | null;
	version: string | null;
	installed: boolean;
	confidence: number;
}

export interface IncomingItem {
	name: string;
	size: number;
	mtime: string;
	ready: boolean;
	ignored: boolean;
	/** The watcher has an obvious match and imports the file on its own. */
	auto: boolean;
	suggestions: IncomingSuggestion[];
}

export interface IncomingListResponse {
	configured: boolean;
	folder: string;
	items: IncomingItem[];
	count: number;
	warning?: string;
}

/** What one update check found for a game. */
export type UpdateResult = "update" | "current" | "error";

export interface UpdateHistoryEntry {
	id: number;
	gameId: number;
	title: string;
	checkedAt: string;
	from: string | null;
	to: string | null;
	result: UpdateResult;
	error: string | null;
	/** Cover file cached by the launcher, relative to the media folder. */
	coverPath: string | null;
	/** Remote cover image. */
	coverUrl: string | null;
	/** Source id of the game, so the history can be paged by source. */
	source: string;
}

export interface UpdatesResponse {
	count: number;
	games: Game[];
	/** Titles of games whose check failed, with the reason. */
	errors: string[];
	/** When the last check ran, or null when none ran. */
	lastCheckedAt: string | null;
}

/** The Updates page state, read without starting a check. */
export interface UpdateStatus {
	/** When the last saved check ran, or null when none ran. */
	lastCheckedAt: string | null;
	/** True while a check runs. */
	running: boolean;
	progress: UpdateProgress | null;
	/** Unseen updates. */
	count: number;
	/** When the next background check starts, or null when none is scheduled. */
	nextRunAt: string | null;
	/** Games with a newer version. */
	games: Game[];
}

export interface UpdateProgress {
	done: number;
	total: number;
}

/** The result of one game in a check run. Skipped games were not asked: their source needs a sign-in. */
export interface UpdateCheckOutcome {
	gameId: number;
	title: string;
	from: string | null;
	to: string | null;
	result: UpdateResult | "skipped";
	error: string | null;
	coverPath: string | null;
	coverUrl: string | null;
}

/** One source that could not be checked. There is one per source. */
export interface UpdateCheckIssue {
	source: string;
	label: string;
	kind: "auth" | "error";
	/** A plain sentence for the user. */
	message: string;
	/** Games of this source that were not checked. */
	games: number;
}

/** A progress push of a check run. The issues come with the final "done" event. */
export interface UpdateCheckEvent {
	state: "running" | "done";
	done: number;
	total: number;
	skipped: number;
	current: { gameId: number; title: string } | null;
	result: UpdateCheckOutcome | null;
	issues: UpdateCheckIssue[];
}

export type DownloadStatus = "queued" | "downloading" | "paused" | "done" | "error" | "canceled";

export interface DownloadItem {
	id: number;
	gameId: number | null;
	url: string;
	filename: string;
	size: number;
	bytesDone: number;
	status: DownloadStatus;
	error: string | null;
	speed: number;
	eta: number | null;
	connections: number;
	/** True while a total speed limit applies to this download. */
	rateLimited: boolean;
	createdAt: string;
	finishedAt: string | null;
}

export type InstallPhase = "extracting" | "unpacking" | "finalizing" | "done";

export interface InstallProgress {
	downloadId: number;
	gameId: number;
	title: string;
	phase: InstallPhase;
	done: number;
	total: number;
	progress: number | null;
	speed: number;
	eta: number | null;
	/** Name of the extraction tool, such as "7-Zip". */
	engine: string | null;
	state: "active" | "done" | "error";
	error: string | null;
}

export interface Toast {
	level: string;
	message: string;
}

export interface DownloadsListResponse {
	items: DownloadItem[];
	/** Installs running now. */
	installs: InstallProgress[];
	/** 0 means no limit. */
	speedLimitKBps: number;
	/** True when 7-Zip is installed. Without it, some archives unpack more slowly or not at all. */
	sevenZip: boolean;
}

export type DownloadAction = "pause" | "resume" | "cancel" | "retry" | "remove";

export type PluginStatus = "broken" | "disabled" | "running" | "stopped";

export interface PluginInfo {
	id: string;
	name: string;
	version: string;
	enabled: boolean;
	signed: boolean;
	status: PluginStatus;
	/** True when the launcher stopped the plugin to save memory. */
	idle: boolean;
	/** When the launcher last called the plugin, or null before its first call. */
	lastActiveAt: string | null;
	error: string | null;
	/** Why the plugin is switched off, when the launcher turned it off. */
	disabledReason: string | null;
	author: string;
	description: string;
	homepage: string;
	repository: string | null;
	/** True when the plugin has a settings form. */
	hasSettings: boolean;
	/** Source ids this plugin contributes. */
	sources: string[];
	/** Source ids that can check for updates. */
	updateSources: string[];
	/** Where the plugin was installed from. Updates download from here. */
	source: string | null;
	/** The permissions the installed version has, as consent entries. */
	consent: string[];
	/** Command palette actions the plugin declares. */
	commands: PluginCommand[];
}

export interface PluginCommand {
	id: string;
	label: string;
	/** Null for a command that only shows up; the host refuses to run it. */
	method: string | null;
}

export type PermissionKind = "site" | "files" | "sign-in" | "browser" | "program" | "other";

/** One permission on the consent screen: a plain sentence and how it is classed. */
export interface PermissionRow {
	text: string;
	kind: PermissionKind;
	sensitive: boolean;
}

/** Permission lines for the consent screen. The sites are grouped, the rest are one row each. */
export interface PermissionLines {
	/** One line per site, such as "example.com and its subdomains". Sorted. */
	sites: string[];
	other: PermissionRow[];
}

export interface PluginPreview {
	source: string;
	id: string;
	name: string;
	version: string;
	signed: boolean;
	/** Plain-language permission lines for the user. */
	permissions: PermissionLines;
	/** The entries to send back as accepted permissions when the user agrees. */
	consent: string[];
	/** True when the new version names a different repository than the installed one. */
	repositoryChanged: boolean;
	installedRepository: string | null;
}

export interface PluginUpdatePreview {
	/** The newest version at the plugin's install source. */
	preview: PluginPreview;
	/** Plain-language lines for the permissions the new version adds. */
	newPermissions: PermissionLines;
	/** Where the update package is downloaded from. */
	source: string;
	/** Release checksum, when the release publishes SHA256SUMS. */
	sha256: string | null;
}

export interface PluginUpdateStatus {
	id: string;
	installedVersion: string;
	available: boolean;
	version: string | null;
	assetUrl: string | null;
	sha256Url: string | null;
	sha256: string | null;
	/** Release notes as plain text. */
	notes: string | null;
	publishedAt: string | null;
	error: string | null;
	/** Stable failure reason: not_found, rate_limited, untrusted_host or check_failed. */
	code: string | null;
}

export interface PluginUpdatesStatus {
	/** When the last check ran, as an RFC 3339 time. */
	checkedAt: string | null;
	plugins: PluginUpdateStatus[];
}

/** The JSON Schema keywords the settings form reads, plus the x-ll-* extensions. See docs/writing-plugins.md. */
export interface JsonSchemaProperty {
	type?: "string" | "number" | "integer" | "boolean" | "object" | "array";
	title?: string;
	description?: string;
	default?: unknown;
	examples?: unknown[];
	enum?: (string | number)[];
	minimum?: number;
	maximum?: number;
	minLength?: number;
	maxLength?: number;
	minItems?: number;
	maxItems?: number;
	uniqueItems?: boolean;
	items?: JsonSchemaProperty;
	properties?: Record<string, JsonSchemaProperty>;
	required?: string[];
	"x-ll-group"?: string;
	"x-ll-widget"?: "folder" | "file";
	"x-ll-advanced"?: boolean;
}

export interface JsonSchema {
	type?: "object";
	properties?: Record<string, JsonSchemaProperty>;
	required?: string[];
}

export interface PluginSettingsResponse {
	/** The settings schema from plugin.json, or null when the plugin has no settings. */
	schema: JsonSchema | null;
	values: Record<string, unknown>;
}

export interface PluginLogLine {
	/** When the launcher received the line, as an RFC 3339 time. */
	time: string;
	level: "debug" | "info" | "warn" | "error";
	message: string;
}

export interface AuthState {
	status: "unconfigured" | "ok" | "expired" | "invalid" | "error";
	user?: string | null;
	message?: string | null;
}

export type PluginMethod = "source.authStatus" | "source.setCredential" | "source.clearCredential";
