const URI_LIST_TYPE = "text/uri-list";
const FILE_URI_PREFIX = "file://";
const WINDOWS_DRIVE_PATH = /^\/[A-Za-z]:/;

/** Reads the local paths out of a `text/uri-list` drag payload. Only file:// lines are kept. */
export function pathsFromUriList(uriList: string): string[] {
	return uriList
		.split(/\r?\n/)
		.map((line) => line.trim())
		.filter((line) => line.startsWith(FILE_URI_PREFIX))
		.map(pathFromFileUri);
}

function pathFromFileUri(uri: string): string {
	const pathname = decodeURIComponent(new URL(uri).pathname);
	return WINDOWS_DRIVE_PATH.test(pathname) ? pathname.slice(1) : pathname;
}

/** True when a drag carries file paths, so the page can read them. */
export function carriesUriList(event: DragEvent): boolean {
	return event.dataTransfer?.types.includes(URI_LIST_TYPE) ?? false;
}

export function droppedPaths(dataTransfer: DataTransfer): string[] {
	return pathsFromUriList(dataTransfer.getData(URI_LIST_TYPE));
}
