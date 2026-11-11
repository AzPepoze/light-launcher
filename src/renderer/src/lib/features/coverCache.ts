import { createLogger } from "../logger";

const log = createLogger("Covers");

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

export interface CoverCache {
	/** Resolves to a blob URL of the cached cover, or null when there is none or it fails to load. */
	load(game: CoverSource): Promise<string | null>;
}

interface ShownCover {
	key: string;
	url: string;
}

/** The server route that serves the cached cover of a game. */
export function coverRoute(id: number): string {
	return `/api/cover/${id}`;
}

/** Identifies one cached cover file. Null when the game has no local cover. */
export function coverKey(game: CoverSource): string | null {
	return game.coverLocal ? `${game.id}:${game.coverLocal}` : null;
}

/**
 * Keeps one blob URL per game, so each cover is downloaded once.
 * A game's old URL is revoked when its cover file changes or the cover goes away.
 */
export function createCoverCache(loader: BlobLoader): CoverCache {
	const shown = new Map<number, ShownCover>();
	const wantedKey = new Map<number, string>();
	const downloads = new Map<string, Promise<string | null>>();

	function forget(id: number): void {
		const previous = shown.get(id);
		if (previous) {
			loader.revokeUrl(previous.url);
			shown.delete(id);
		}
		wantedKey.delete(id);
	}

	/** Keeps a downloaded cover only while its game still asks for that file. */
	function keep(id: number, key: string, url: string): string | null {
		if (wantedKey.get(id) !== key) {
			loader.revokeUrl(url);
			return null;
		}
		const previous = shown.get(id);
		if (previous && previous.key !== key) {
			loader.revokeUrl(previous.url);
		}
		shown.set(id, { key, url });
		return url;
	}

	function download(id: number, key: string): Promise<string | null> {
		const running = downloads.get(key);
		if (running) return running;

		const next = loader
			.fetchBlob(coverRoute(id))
			.then((blob) => keep(id, key, loader.createUrl(blob)))
			.catch((error: unknown) => {
				log.warn("Cover could not be loaded", error);
				return null;
			})
			.finally(() => downloads.delete(key));
		downloads.set(key, next);
		return next;
	}

	async function load(game: CoverSource): Promise<string | null> {
		const key = coverKey(game);
		if (key === null) {
			forget(game.id);
			return null;
		}
		wantedKey.set(game.id, key);
		const current = shown.get(game.id);
		if (current?.key === key) return current.url;
		return download(game.id, key);
	}

	return { load };
}
