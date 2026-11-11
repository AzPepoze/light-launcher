import { createLogger } from "../logger";

const log = createLogger("Media");

/** The game fields that decide which cover to show. */
export interface CoverSource {
	id: number;
	cover: string | null;
	coverLocal: string | null;
}

/** Downloads a file and turns it into a URL the page can show, then releases that URL. */
export interface BlobLoader {
	fetchBlob(path: string): Promise<Blob>;
	createUrl(blob: Blob): string;
	revokeUrl(url: string): void;
}

/**
 * Keeps blob URLs for cached images. A URL is revoked as soon as no mounted scope wants it.
 */
export interface MediaCache {
	/** Takes a reference for scope and resolves to a blob URL, or null when there is none or it fails. */
	load(scope: string, key: string | null, route: string): Promise<string | null>;
	/** Drops the reference of a scope, revoking its URL when no other scope wants it. */
	release(scope: string): void;
	/** Revokes every URL. */
	dispose(): void;
}

/** The server route that serves the cached cover of a game. */
export function coverRoute(id: number): string {
	return `/api/media/cover/${id}`;
}

/** The server route that serves one cached screenshot of a game. */
export function screenshotRoute(id: number, index: number): string {
	return `/api/media/screenshot/${id}/${index}`;
}

/** Identifies one cached cover file. Null when the game has no local cover. */
export function coverKey(game: CoverSource): string | null {
	return game.coverLocal ? `cover:${game.id}:${game.coverLocal}` : null;
}

/** Identifies one cached screenshot file. Null when the screenshot has no local copy. */
export function screenshotKey(gameId: number, index: number, local: string | null): string | null {
	return local ? `screenshot:${gameId}:${index}:${local}` : null;
}

export function createMediaCache(loader: BlobLoader): MediaCache {
	const shown = new Map<string, string>();
	/** The file each mounted scope shows. */
	const wanted = new Map<string, string>();
	const wantedCount = new Map<string, number>();
	const downloads = new Map<string, Promise<string | null>>();

	function drop(key: string): void {
		const url = shown.get(key);
		if (url === undefined) return;
		loader.revokeUrl(url);
		shown.delete(key);
	}

	function releaseKey(key: string): void {
		const count = (wantedCount.get(key) ?? 1) - 1;
		if (count > 0) {
			wantedCount.set(key, count);
			return;
		}
		wantedCount.delete(key);
		drop(key);
	}

	function setWanted(scope: string, key: string | null): void {
		const previous = wanted.get(scope);
		if (previous === key) return;
		if (previous !== undefined) releaseKey(previous);
		if (key === null) {
			wanted.delete(scope);
			return;
		}
		wanted.set(scope, key);
		wantedCount.set(key, (wantedCount.get(key) ?? 0) + 1);
	}

	function download(key: string, route: string): Promise<string | null> {
		const running = downloads.get(key);
		if (running) return running;

		const next = loader
			.fetchBlob(route)
			.then((blob) => {
				const url = loader.createUrl(blob);
				if ((wantedCount.get(key) ?? 0) === 0) {
					loader.revokeUrl(url);
					return null;
				}
				const current = shown.get(key);
				if (current !== undefined) {
					loader.revokeUrl(url);
					return current;
				}
				shown.set(key, url);
				return url;
			})
			.catch((error: unknown) => {
				log.warn("Image could not be loaded", error);
				return null;
			})
			.finally(() => downloads.delete(key));
		downloads.set(key, next);
		return next;
	}

	async function load(scope: string, key: string | null, route: string): Promise<string | null> {
		setWanted(scope, key);
		if (key === null) return null;
		const current = shown.get(key);
		if (current !== undefined) return current;
		return download(key, route);
	}

	function release(scope: string): void {
		const key = wanted.get(scope);
		if (key === undefined) return;
		wanted.delete(scope);
		releaseKey(key);
	}

	function dispose(): void {
		for (const url of shown.values()) {
			loader.revokeUrl(url);
		}
		shown.clear();
		wanted.clear();
		wantedCount.clear();
		downloads.clear();
	}

	return { load, release, dispose };
}
