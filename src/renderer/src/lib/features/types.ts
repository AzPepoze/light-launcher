// View types for the library, downloads, updates and plugin pages.
// The Go payloads are in wire.ts. The mappers in *Map.ts turn them into these types.

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
}

export interface LibraryStats {
	total: number;
	installed: number;
	updates: number;
	bytes: number;
	playtime: number;
}

export interface LibraryListResponse {
	items: Game[];
	stats: LibraryStats;
	sources: { source: string; label: string; count: number }[];
	tags: { tag: string; count: number }[];
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

export interface UpdateHistoryEntry {
	id: number;
	gameId: number;
	title: string;
	checkedAt: string;
	from: string | null;
	to: string | null;
	result: "update" | "current" | "error";
	error: string | null;
}

export interface UpdatesResponse {
	count: number;
	games: Game[];
	/** Titles of games whose check failed, with the reason. */
	errors: string[];
	/** When the last check ran, or null when none ran. */
	lastCheckedAt: string | null;
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

export type PluginStatus = "ok" | "error";

export interface PluginInfo {
	id: string;
	name: string;
	version: string;
	enabled: boolean;
	signed: boolean;
	status: PluginStatus;
	error: string | null;
	/** Why the plugin is switched off, when the launcher turned it off. */
	disabledReason: string | null;
	author: string;
	description: string;
	homepage: string;
	/** True when the plugin has a settings form. */
	hasSettings: boolean;
	/** Source ids this plugin contributes. */
	sources: string[];
	/** Where the plugin was installed from. Updates download from here. */
	source: string | null;
	/** The permissions the installed version has, as consent entries. */
	consent: string[];
}

export interface PluginPreview {
	source: string;
	id: string;
	name: string;
	version: string;
	signed: boolean;
	/** Plain-language permission lines for the user. */
	permissions: string[];
	/** The entries to send back as accepted permissions when the user agrees. */
	consent: string[];
}

export interface PluginUpdatePreview {
	/** The newest version at the plugin's install source. */
	preview: PluginPreview;
	/** Plain-language lines for the permissions the new version adds. */
	newPermissions: string[];
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
