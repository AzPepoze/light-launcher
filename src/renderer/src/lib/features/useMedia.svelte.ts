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

function useMediaImage(request: () => MediaRequest): { readonly url: string | null } {
	const scope = `media-${++nextScope}`;
	let url = $state<string | null>(null);

	$effect(() => {
		const wanted = request();
		let active = true;
		url = null;
		if (wanted.key === null) {
			media.release(scope);
			url = wanted.fallback;
			return;
		}
		void media.load(scope, wanted.key, wanted.route).then((local) => {
			if (active) url = local ?? wanted.fallback;
		});
		return () => {
			active = false;
		};
	});

	onDestroy(() => media.release(scope));

	return {
		get url() {
			return url;
		}
	};
}

/**
 * The image URL to show for a game's cover. It is the cached copy when the launcher has one,
 * otherwise the remote cover. Call it during component setup.
 */
export function useCover(getGame: () => CoverSource): { readonly url: string | null } {
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
export function useScreenshot(getShot: () => ScreenshotSource): { readonly url: string | null } {
	return useMediaImage(() => {
		const shot = getShot();
		return {
			key: screenshotKey(shot.gameId, shot.index, shot.local),
			route: screenshotRoute(shot.gameId, shot.index),
			fallback: shot.remote
		};
	});
}
