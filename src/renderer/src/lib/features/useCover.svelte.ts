import { fetchBlob } from "../transport";
import { createCoverCache, type CoverSource } from "./coverCache";

const covers = createCoverCache({
	fetchBlob,
	createUrl: (blob) => URL.createObjectURL(blob),
	revokeUrl: (url) => URL.revokeObjectURL(url)
});

/**
 * The image URL to show for a game's cover. It is the cached copy when the launcher has one,
 * otherwise the remote cover. Call it during component setup.
 */
export function useCover(getGame: () => CoverSource): { readonly url: string | null } {
	let url = $state<string | null>(null);

	$effect(() => {
		const game = getGame();
		let active = true;
		url = null;
		void covers.load(game).then((local) => {
			if (active) url = local ?? game.cover;
		});
		return () => {
			active = false;
		};
	});

	return {
		get url() {
			return url;
		}
	};
}
