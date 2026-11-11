import { postForm } from "@lib/transport";

const UPLOAD_PATH = "/api/plugin-upload";
const ARCHIVE_EXTENSIONS = [".llplugin", ".zip"];
const MANIFEST_IN_FOLDER = /^[^/]+\/plugin\.json$/;

export interface UploadFile {
	/** Relative path with '/' separators, such as "myplugin/plugin.json". */
	path: string;
	file: File;
}

export type UploadProblem = "not-a-plugin" | "several-plugins";

export type UploadPlan = { ok: true; files: UploadFile[] } | { ok: false; problem: UploadProblem };

export function isArchivePath(path: string): boolean {
	const lower = path.toLowerCase();
	return ARCHIVE_EXTENSIONS.some((extension) => lower.endsWith(extension));
}

/** The first folder of each path. Files at the top level have no folder. */
function topFolders(files: UploadFile[]): Set<string> {
	const folders = new Set<string>();
	for (const { path } of files) {
		const [first, ...rest] = path.split("/");
		if (rest.length > 0) folders.add(first);
	}
	return folders;
}

function isSingleArchive(files: UploadFile[]): boolean {
	return files.length === 1 && isArchivePath(files[0].path);
}

function hasManifestInFolder(files: UploadFile[]): boolean {
	return files.some(({ path }) => MANIFEST_IN_FOLDER.test(path));
}

/** Accepts one .llplugin or .zip file, or one folder that has plugin.json at its top level. */
export function planPluginUpload(files: UploadFile[]): UploadPlan {
	if (isSingleArchive(files)) return { ok: true, files };
	if (!hasManifestInFolder(files)) return { ok: false, problem: "not-a-plugin" };
	if (topFolders(files).size > 1) return { ok: false, problem: "several-plugins" };
	return { ok: true, files };
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
