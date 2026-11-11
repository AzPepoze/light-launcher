// Payloads and replies of the Go API, as they arrive on the wire.
// Field names match the json tags in core/internal. Mappers turn these into the UI types in types.ts.

import type { JsonSchema } from "./types";

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
	coverLocal: string | null;
	screenshots: string[] | null;
	screenshotsLocal: (string | null)[] | null;
	description: string | null;
	descriptionText: string | null;
	changelog: string | null;
	rating: number | null;
	status: string;
	running: boolean;
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

export interface WireGameFile {
	id: number;
	gameId: number;
	label: string;
	url: string;
	filename: string | null;
	size: number | null;
	host: string | null;
	kind: string;
	platform: string;
	rangeHint: number | null;
	requiresAuth: number;
}

export interface WireCheck {
	id: number;
	gameId: number;
	checkedAt: string;
	oldVersion: string | null;
	newVersion: string | null;
	changelog: string | null;
	seen: number;
	applied: number;
}

export interface WireGameDetail extends WireGame {
	files: WireGameFile[] | null;
	changes: WireCheck[] | null;
}

export interface WireLibraryStats {
	total: number;
	installed: number;
	updates: number;
	bytes: number;
	playtime: number;
}

export interface WireLibraryPage {
	items: WireGame[] | null;
	total: number;
	sources: { id: string; label: string; count: number }[] | null;
	tags: { tag: string; count: number }[] | null;
	stats: WireLibraryStats;
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
	installDir?: string;
}

export interface WireUpdateCheckRequest {
	ids?: number[];
	force?: boolean;
}

export interface WireUpdatesSummary {
	unseen: number;
	last: { checkedAt: string; results: WireCheckResult[] | null } | null;
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

export interface WireSearchGame extends WireGameDTO {
	score: number;
}

export interface WireSearchResult {
	plugin: string;
	source: string;
	games: WireSearchGame[] | null;
	error?: string;
}

export interface WireFileCandidate {
	id: string;
	label: string;
	url: string;
	filename?: string;
	size?: number;
	host?: string;
	kind: string;
	platform?: string;
	requiresAuth?: boolean;
}

export interface WirePreview {
	game: WireGameDTO;
	existing: { id: number; title: string; installed: boolean } | null;
	files: WireFileCandidate[] | null;
	sourceLabel: string;
}

export interface WireUpdateHistoryEntry {
	id: number;
	gameId: number;
	title: string;
	checkedAt: string;
	from: string | null;
	to: string | null;
	result: string;
	error: string | null;
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
	rateLimited: boolean;
	createdAt: string;
	finishedAt?: string;
}

export interface WireDownloadsStatus {
	installs: WireInstallProgress[] | null;
	sevenZip: boolean;
	speedLimitKBps: number;
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
	title?: string;
	state: string;
	phase: string;
	done: number;
	total: number;
	progress: number | null;
	speed: number;
	eta: number | null;
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
	author: string;
	description: string;
	homepage: string;
	hasSettings: boolean;
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
	consent: string[] | null;
}

export interface WirePluginNotify {
	plugin: string;
	level: string;
	message: string;
}

export interface WireSettingsView {
	schema: JsonSchema | null;
	values: Record<string, unknown> | null;
}

export interface WirePluginLogLine {
	time: string;
	level?: string;
	text: string;
}
