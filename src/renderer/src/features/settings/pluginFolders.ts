import { folderKey } from "@lib/features/scanFolders";

/** Whether a folder is the parent folder or sits inside it. Compares the way folderKey does: case and trailing slashes do not matter. */
export function isInsideFolder(folder: string, parent: string): boolean {
	const child = folderKey(folder);
	const root = folderKey(parent);
	if (child === "" || root === "") return false;
	return child === root || child.startsWith(`${root}/`) || child.startsWith(`${root}\\`);
}
