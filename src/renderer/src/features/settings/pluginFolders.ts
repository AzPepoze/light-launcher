import type { LibrarySource } from "@lib/features/types";
import { folderKey } from "@lib/features/scanFolders";

/** How two folders relate: the same folder, one inside the other, or one containing the other. */
export type FolderRelation = "same" | "inside" | "contains";

/** One plugin source whose games folder overlaps a scan folder. */
export interface PluginFolderMatch {
	source: string;
	label: string;
	/** The resolved games folder of the plugin, as the overview sent it. */
	folder: string;
	relation: FolderRelation;
}

/** Whether a folder is the parent folder or sits inside it. Compares the way folderKey does: case and trailing slashes do not matter. */
export function isInsideFolder(folder: string, parent: string): boolean {
	const relation = folderRelation(folder, parent);
	return relation === "same" || relation === "inside";
}

/**
 * How a folder and another folder relate. Both paths are compared the way folderKey does, so a
 * trailing slash or another case does not hide an overlap. Unrelated paths do not match.
 */
export function folderRelation(folder: string, other: string): FolderRelation | null {
	const child = comparableKey(folder);
	const root = comparableKey(other);
	if (child === "" || root === "") return null;
	if (child === root) return "same";
	if (isUnder(child, root)) return "inside";
	if (isUnder(root, child)) return "contains";
	return null;
}

// comparableKey returns the folderKey of a path, keeping a bare root separator so the root can be
// compared as a parent. folderKey on its own drops the root and returns an empty key for it.
function comparableKey(path: string): string {
	const key = folderKey(path);
	if (key !== "") return key;
	const root = path.trim()[0];
	return root === "/" || root === "\\" ? root : "";
}

// isUnder reports whether child sits below parent. The root folder is the parent of every path.
function isUnder(child: string, parent: string): boolean {
	if (child === parent) return true;
	if (parent === "/" || parent === "\\") return child.startsWith(parent);
	return child.startsWith(`${parent}/`) || child.startsWith(`${parent}\\`);
}

/**
 * The plugin sources that use a scan folder: the same folder, a folder inside it, or a folder that
 * contains it. Sources without a games folder, such as local games, never match.
 */
export function pluginFolderMatches(folder: string, sources: LibrarySource[]): PluginFolderMatch[] {
	const matches: PluginFolderMatch[] = [];
	for (const source of sources) {
		if (!source.gamesFolder) continue;
		const relation = folderRelation(folder, source.gamesFolder);
		if (!relation) continue;
		matches.push({
			source: source.source,
			label: source.label.trim() || source.source,
			folder: source.gamesFolder,
			relation
		});
	}
	return matches;
}

/** The sentence shown in the scan-folder row for one plugin games folder it overlaps. */
export function pluginFolderWarning(match: PluginFolderMatch): string {
	switch (match.relation) {
		case "same":
			return `${match.label} already uses this folder for its games. Scanning it can add duplicates.`;
		case "inside":
			return `This folder is inside ${match.label}'s games folder (${match.folder}). Scanning it can add duplicates.`;
		case "contains":
			return `This folder contains ${match.label}'s games folder (${match.folder}). Scanning it can add duplicates.`;
	}
}
