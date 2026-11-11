import { postForm } from "@lib/transport";

const UPLOAD_PATH = "/api/plugin-upload";
const ARCHIVE_EXTENSIONS = [".llplugin", ".zip"];
const MANIFEST = "plugin.json";
/** A manifest one folder below a dropped folder, such as "bundle/one/plugin.json". */
const NESTED_MANIFEST = /^[^/]+\/[^/]+\/plugin\.json$/;

export interface UploadFile {
	/** Relative path with '/' separators, such as "myplugin/plugin.json". */
	path: string;
	file: File;
}

/** "not-a-plugin": nothing usable. "nested-plugins": a folder that holds plugin folders. */
export type SkipReason = "not-a-plugin" | "nested-plugins";

export interface SkippedUpload {
	/** The top-level name of the skipped file or folder. */
	name: string;
	reason: SkipReason;
}

export interface PluginUploadPlan {
	/** The files of each plugin, in drop order. */
	plugins: UploadFile[][];
	skipped: SkippedUpload[];
}

export function isArchivePath(path: string): boolean {
	const lower = path.toLowerCase();
	return ARCHIVE_EXTENSIONS.some((extension) => lower.endsWith(extension));
}

function topEntryName(path: string): string {
	return path.split("/")[0];
}

/** Groups the files by their top-level name, in the order the names first appear. */
function groupByTopEntry(files: UploadFile[]): Map<string, UploadFile[]> {
	const groups = new Map<string, UploadFile[]>();
	for (const file of files) {
		const name = topEntryName(file.path);
		const group = groups.get(name) ?? [];
		group.push(file);
		groups.set(name, group);
	}
	return groups;
}

function isArchiveFile(name: string, group: UploadFile[]): boolean {
	return group.length === 1 && group[0].path === name && isArchivePath(name);
}

function hasManifest(name: string, group: UploadFile[]): boolean {
	return group.some(({ path }) => path === `${name}/${MANIFEST}`);
}

function hasNestedManifest(group: UploadFile[]): boolean {
	return group.some(({ path }) => NESTED_MANIFEST.test(path));
}

function skipReason(group: UploadFile[]): SkipReason {
	return hasNestedManifest(group) ? "nested-plugins" : "not-a-plugin";
}

/**
 * Splits a drop into plugins. Each top-level archive is one plugin, and each top-level
 * folder with plugin.json directly inside is one plugin. Everything else is skipped by name.
 */
export function planPluginUploads(files: UploadFile[]): PluginUploadPlan {
	const plugins: UploadFile[][] = [];
	const skipped: SkippedUpload[] = [];
	for (const [name, group] of groupByTopEntry(files)) {
		if (isArchiveFile(name, group) || hasManifest(name, group)) {
			plugins.push(group);
		} else {
			skipped.push({ name, reason: skipReason(group) });
		}
	}
	return { plugins, skipped };
}

/** Uploads the plugin files and returns the server path to use as an install source. */
export async function uploadPlugin(files: UploadFile[]): Promise<string> {
	const form = new FormData();
	for (const { path, file } of files) {
		form.append(path, file, path);
	}
	const result = await postForm<{ path: string }>(UPLOAD_PATH, form);
	return result.path;
}
