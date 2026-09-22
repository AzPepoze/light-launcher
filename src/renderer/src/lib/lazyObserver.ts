type LazyCallback = () => void;

const callbacks = new WeakMap<Element, LazyCallback>();
let observer: IntersectionObserver | null = null;

function getObserver(): IntersectionObserver | null {
	if (typeof IntersectionObserver === "undefined") return null;
	if (!observer) {
		observer = new IntersectionObserver(
			(entries) => {
				for (const entry of entries) {
					if (!entry.isIntersecting) continue;
					const cb = callbacks.get(entry.target);
					if (cb) cb();
				}
			},
			{ rootMargin: "250px" }
		);
	}
	return observer;
}

/** Observe until the element first intersects; invokes cb once, then unobserves. */
export function observeOnce(el: Element, cb: LazyCallback): () => void {
	const obs = getObserver();
	if (!obs) {
		cb();
		return () => {};
	}
	callbacks.set(el, cb);
	obs.observe(el);
	return () => {
		callbacks.delete(el);
		obs.unobserve(el);
	};
}
