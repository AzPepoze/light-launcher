// Shapes returned by the Go API methods used by the library, downloads, updates and plugin pages.
// Assumed until the Go side lands: keep this file in sync with the method handlers.

export interface Game {
	id: number;
	source: string;
	url: string;
	title: string;
	version: string | null;
	developer: string | null;
	engine: string | null;
	tags: string[];
	cover: string | null;
	screenshots: string[];
	description: string | null;
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
	running: boolean;
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
	sources: { source: string; count: number }[];
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
	Pick<Game, "title" | "notes" | "exePath" | "exeArgs" | "installDir" | "favorite">
>;

export interface SearchHit {
	url: string;
	title: string;
	score: number;
}

export interface SearchResponse {
	hits: SearchHit[];
}

export interface PreviewFile {
	label: string;
	host: string | null;
	platform: string;
	size: number | null;
	kind: string;
	requiresAuth: boolean;
}

export interface GamePreview {
	source: string;
	sourceLabel: string;
	title: string;
	version: string | null;
	developer: string | null;
	engine: string | null;
	tags: string[];
	description: string | null;
	cover: string | null;
	url: string;
	existing: { id: number; title: string; installed: boolean } | null;
	files: PreviewFile[];
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

export interface UpdateCheckRow {
	id: number;
	gameId: number;
	checkedAt: string;
	from: string | null;
	to: string | null;
	changelog: string | null;
	seen: boolean;
	applied: boolean;
	title: string;
	cover: string | null;
}

export interface UpdatesResponse {
	count: number;
	unseen: number;
	games: Game[];
	history: UpdateCheckRow[];
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
	rateLimited?: boolean;
	connections: number;
	createdAt: string;
	finishedAt: string | null;
}

export interface InstallProgress {
	downloadId: number;
	gameId: number;
	title: string;
	phase: "extracting" | "unpacking" | "finalizing";
	done: number;
	total: number;
	progress: number | null;
	speed: number;
	eta: number | null;
	/** Name of the extraction tool, such as "7-Zip". */
	engine: string | null;
	state: "active" | "done" | "error";
	error?: string | null;
}

export interface DownloadsListResponse {
	items: DownloadItem[];
	installs: InstallProgress[];
	sevenZip: boolean;
	/** 0 means no limit. */
	speedLimitKBps: number;
}

export type DownloadAction = "pause" | "resume" | "cancel" | "retry" | "remove";

export const DOWNLOAD_PROGRESS_EVENT = "download-progress";
export const INSTALL_PROGRESS_EVENT = "install-progress";
export const SEVEN_ZIP_MISSING_EVENT = "sevenzip-missing";

export type PluginStatus = "ok" | "error";

export interface PluginInfo {
	id: string;
	name: string;
	version: string;
	author: string | null;
	enabled: boolean;
	signed: boolean;
	status: PluginStatus;
	error: string | null;
	sources: string[];
	hasSettings: boolean;
}

export interface PluginPreview {
	source: string;
	id: string;
	name: string;
	version: string;
	author: string | null;
	description: string | null;
	signed: boolean;
	permissions: string[];
}

export interface JsonSchemaProperty {
	type?: "string" | "number" | "integer" | "boolean" | "object" | "array";
	title?: string;
	description?: string;
	default?: string | number | boolean;
	enum?: (string | number)[];
	minimum?: number;
	maximum?: number;
}

export interface JsonSchema {
	type?: "object";
	properties?: Record<string, JsonSchemaProperty>;
	required?: string[];
}

export interface PluginSettingsResponse {
	schema: JsonSchema | null;
	values: Record<string, unknown>;
}

export interface PluginLogLine {
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
