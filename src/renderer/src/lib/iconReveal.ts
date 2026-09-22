/** Adds `loaded` when the image is ready, so CSS can fade it in. */
export function iconFade(node: HTMLImageElement, _src?: string) {
	const reveal = () => node.classList.add("loaded");
	const reset = () => node.classList.remove("loaded");
	const check = () => {
		if (node.complete && node.naturalWidth > 0) reveal();
	};

	reset();
	node.addEventListener("load", reveal);
	node.addEventListener("error", reveal);
	check();

	return {
		update(_src?: string) {
			reset();
			check();
		},
		destroy() {
			node.removeEventListener("load", reveal);
			node.removeEventListener("error", reveal);
		}
	};
}
