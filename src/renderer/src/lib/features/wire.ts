// Payloads and replies of the Go API, as they arrive on the wire.
// Field names match the json tags in core/internal. Mappers turn these into the UI types in types.ts.

// Library

export interface WireGame {
	id: number;
	source: string;
	sourceId: string;
	url: string;
	title: string;
	version: string | null;
	developer: string | null;
	engine: string | null;
	tags: string[] | null;
	coverPath: string | null;
	coverUrl: string | null;
	screenshots: string[] | null;
	description: string | null;
	changelog: string | null;
	rating: number | null;
	status: string;
	favorite: number;
	hidden: number;
	notes: string | null;
	installedVersion: string | null;
	installDir: string | null;
	installSize: number | null;
	exePath: string | null;
	exeArgs: string | null;
	lastPlayedAt: string | null;
	playtimeS: number;
	addedAt: string;
	updatedAt: string;
}

export interface WireLibraryPage {
	items: WireGame[] | null;
	total: number;
}

export interface WireLibraryFilter {
	search?: string;
	source?: string;
	tags?: string[];
	status?: string[];
	favorite?: boolean;
	installed?: boolean;
	updates?: boolean;
	hidden?: boolean;
	sort?: string;
	order?: string;
	limit?: number;
	offset?: number;
}

export interface WireGameUpdate {
	id: number;
	title?: string;
	notes?: string;
	tags?: string[];
	hidden?: boolean;
	favorite?: boolean;
	exePath?: string;
	exeArgs?: string;
}

export interface WireGameDTO {
	source: string;
	sourceId: string;
	url: string;
	title: string;
	version?: string;
	developer?: string;
	engine?: string;
	tags?: string[];
	coverUrl?: string;
	screenshots?: string[];
	descriptionHtml?: string;
	changelog?: string;
	status?: string;
	updatedAt?: string;
}

export interface WireSearchResult {
	plugin: string;
	source: string;
	games: WireGameDTO[] | null;
	error?: string;
}

export interface WirePreview {
	game: WireGameDTO;
	existing: { id: number; title: string; installed: boolean } | null;
}

export interface WireCheckResult {
	id: number;
	title: string;
	from: string | null;
	to: string | null;
	hasUpdate: boolean;
	error?: string;
}

export interface WireSuggestion {
	gameId: number;
	title: string;
	cover: string | null;
	version: string | null;
	installed: boolean;
	confidence: number;
}

export interface WireIncomingItem {
	name: string;
	size: number;
	mtime: string;
	ready: boolean;
	ignored: boolean;
	auto: boolean;
	suggestions: WireSuggestion[] | null;
}

export interface WireIncomingReport {
	configured: boolean;
	folder: string;
	items: WireIncomingItem[] | null;
	count: number;
	warning?: string;
}

// Downloads

export interface WireDownloadView {
	id: number;
	gameId: number;
	url: string;
	filename: string;
	dest: string;
	state: string;
	error?: string;
	received: number;
	total: number;
	speed: number;
	connections: number;
	install: boolean;
	createdAt: string;
	finishedAt?: string;
}

export interface WireDownloadSettings {
	connections: number;
	maxActive: number;
	hostConnections: number;
	retries: number;
	speedLimitKBps: number;
	keepArchives: boolean;
}

export interface WireDownloadProgress {
	id: number;
	state: string;
	received: number;
	total: number;
	speed: number;
	error?: string;
}

export interface WireInstallProgress {
	downloadId: number;
	gameId: number;
	state: string;
	phase: string;
	done: number;
	total: number;
	progress: number | null;
	speed: number;
	engine?: string;
	installDir?: string;
	exePath?: string;
	error?: string;
}

export interface WireToast {
	level: string;
	message: string;
}

// Plugins

export interface WirePermissions {
	network: string[] | null;
	fs: string[] | null;
	secrets: boolean;
	browser: boolean;
	exec: boolean;
}

export interface WireSource {
	plugin: string;
	sourceId: string;
	label: string;
	capabilities: string[] | null;
}

export interface WirePluginInfo {
	id: string;
	name: string;
	version: string;
	previous?: string;
	status: string;
	enabled: boolean;
	running: boolean;
	error?: string;
	disabledReason?: string;
	source?: string;
	unsigned: boolean;
	permissions: WirePermissions;
	sources: WireSource[] | null;
}

export interface WireRemovedPlugin {
	id: string;
	removed: true;
}

export interface WirePluginPreview {
	manifest: {
		id: string;
		name: string;
		version: string;
		permissions: WirePermissions;
	};
	unsigned: boolean;
}

export interface WirePluginNotify {
	plugin: string;
	level: string;
	message: string;
}
