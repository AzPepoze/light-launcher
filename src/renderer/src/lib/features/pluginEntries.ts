import type { UploadFile } from "./pluginUpload";

const SKIPPED_FOLDERS = new Set(["node_modules", ".git"]);

/** True when any folder in the relative path is node_modules or .git. */
export function isSkippedPath(path: string): boolean {
	return path.split("/").some((part) => SKIPPED_FOLDERS.has(part));
}

/** Turns a dropped entry path ("/folder/file.json") into a relative one ("folder/file.json"). */
export function toRelativePath(fullPath: string): string {
	return fullPath.replace(/^\/+/, "");
}

function readFile(entry: FileSystemFileEntry): Promise<File> {
	return new Promise((resolve, reject) => entry.file(resolve, reject));
}

/** readEntries returns the children in batches, so it is called until a batch is empty. */
async function readAllEntries(reader: FileSystemDirectoryReader): Promise<FileSystemEntry[]> {
	const entries: FileSystemEntry[] = [];
	for (;;) {
		const batch = await new Promise<FileSystemEntry[]>((resolve, reject) =>
			reader.readEntries(resolve, reject)
		);
		if (batch.length === 0) return entries;
		entries.push(...batch);
	}
}

/** Lists the files under an entry, skipping node_modules and .git. */
export async function walkEntry(entry: FileSystemEntry): Promise<UploadFile[]> {
	const path = toRelativePath(entry.fullPath);
	if (isSkippedPath(path)) return [];

	if (!entry.isDirectory) {
		const file = await readFile(entry as FileSystemFileEntry);
		return [{ path, file }];
	}

	const reader = (entry as FileSystemDirectoryEntry).createReader();
	const children = await readAllEntries(reader);
	const nested = await Promise.all(children.map(walkEntry));
	return nested.flat();
}

/**
 * Reads the dropped items. The entries are taken before the first await,
 * because the browser clears DataTransfer once the drop event returns.
 */
export async function collectDroppedFiles(dataTransfer: DataTransfer): Promise<UploadFile[]> {
	const entries = droppedEntries(dataTransfer.items);
	const nested = await Promise.all(entries.map(walkEntry));
	return nested.flat();
}

function droppedEntries(items: DataTransferItemList): FileSystemEntry[] {
	const entries: FileSystemEntry[] = [];
	for (const item of Array.from(items)) {
		if (item.kind !== "file") continue;
		const entry = item.webkitGetAsEntry();
		if (entry) entries.push(entry);
	}
	return entries;
}

/** Converts files from a file input. Folder picks carry their relative path in webkitRelativePath. */
export function filesFromInput(files: File[]): UploadFile[] {
	return files
		.map((file) => ({ path: file.webkitRelativePath || file.name, file }))
		.filter(({ path }) => !isSkippedPath(path));
}
