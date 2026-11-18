import { onDestroy } from "svelte";
import { fetchBlob } from "../transport";
import {
	coverKey,
	coverRoute,
	createMediaCache,
	screenshotKey,
	screenshotRoute,
	type CoverSource
} from "./mediaCache";

const media = createMediaCache({
	fetchBlob,
	createUrl: (blob) => URL.createObjectURL(blob),
	revokeUrl: (url) => URL.revokeObjectURL(url)
});

let nextScope = 0;

interface MediaRequest {
	key: string | null;
	route: string;
	fallback: string | null;
}

export interface ScreenshotSource {
	gameId: number;
	index: number;
	local: string | null;
	remote: string;
}

export interface MediaImage {
	/** The URL to show, or null when there is nothing to show. */
	readonly url: string | null;
	/** True until the URL is decided, so the page can show a placeholder. */
	readonly pending: boolean;
}

function useMediaImage(request: () => MediaRequest): MediaImage {
	const scope = `media-${++nextScope}`;
	let url = $state<string | null>(null);
	let pending = $state(true);

	$effect(() => {
		const wanted = request();
		let active = true;
		url = null;
		pending = true;
		if (wanted.key === null) {
			media.release(scope);
			url = wanted.fallback;
			pending = false;
			return;
		}
		void media.load(scope, wanted.key, wanted.route).then((local) => {
			if (!active) return;
			url = local ?? wanted.fallback;
			pending = false;
		});
		return () => {
			active = false;
		};
	});

	onDestroy(() => media.release(scope));

	return {
		get url() {
			return url;
		},
		get pending() {
			return pending;
		}
	};
}

/**
 * The image URL to show for a game's cover. It is the cached copy when the launcher has one,
 * otherwise the remote cover. Call it during component setup.
 */
export function useCover(getGame: () => CoverSource): MediaImage {
	return useMediaImage(() => {
		const game = getGame();
		return {
			key: coverKey(game),
			route: coverRoute(game.id),
			fallback: game.cover
		};
	});
}

/**
 * The image URL to show for one screenshot. It is the cached copy when the launcher has one,
 * otherwise the remote link. Call it during component setup.
 */
export function useScreenshot(getShot: () => ScreenshotSource): MediaImage {
	return useMediaImage(() => {
		const shot = getShot();
		return {
			key: screenshotKey(shot.gameId, shot.index, shot.local),
			route: screenshotRoute(shot.gameId, shot.index),
			fallback: shot.remote
		};
	});
}

/**
 * The URLs of a list of screenshots, the same ones useScreenshot shows: the cached copy when
 * there is one, otherwise the remote link. Each entry starts as the remote link and switches
 * to the cached copy when it loads. Call it during component setup.
 */
export function useScreenshotUrls(getShots: () => ScreenshotSource[]): { readonly urls: string[] } {
	const scope = `shots-${++nextScope}`;
	let urls = $state<string[]>([]);

	$effect(() => {
		const shots = getShots();
		let active = true;
		urls = shots.map((shot) => shot.remote);
		shots.forEach((shot, position) => {
			const itemScope = `${scope}-${position}`;
			const key = screenshotKey(shot.gameId, shot.index, shot.local);
			const route = screenshotRoute(shot.gameId, shot.index);
			void media.load(itemScope, key, route).then((local) => {
				if (active && local) urls[position] = local;
			});
		});
		return () => {
			active = false;
			shots.forEach((_, position) => media.release(`${scope}-${position}`));
		};
	});

	return {
		get urls() {
			return urls;
		}
	};
}
