const cache = new Map<string, [number, number, number]>();

function rgbToKey(r: number, g: number, b: number): string {
	return `${r},${g},${b}`;
}

function getVibrantScore(r: number, g: number, b: number): number {
	const max = Math.max(r, g, b);
	const min = Math.min(r, g, b);
	const delta = max - min;
	if (max === 0) return 0;
	const saturation = delta / max;
	const value = max / 255;
	// Filter out near-gray and very dark
	if (saturation < 0.15) return 0;
	if (value < 0.2) return 0;
	// Prefer vibrant + bright
	return saturation * 0.7 + value * 0.3 + saturation * value * 0.5;
}

export async function getDominantColor(src: string): Promise<[number, number, number] | null> {
	if (!src) return null;
	if (cache.has(src)) return cache.get(src)!;

	return new Promise((resolve) => {
		const img = new Image();
		// base64 data urls don't need crossOrigin, but set anyway for http
		if (!src.startsWith("data:")) img.crossOrigin = "anonymous";

		const timeout = setTimeout(() => {
			resolve(null);
		}, 3000);

		img.onload = () => {
			clearTimeout(timeout);
			try {
				const canvas = document.createElement("canvas");
				const size = 32;
				canvas.width = size;
				canvas.height = size;
				const ctx = canvas.getContext("2d", { willReadFrequently: true });
				if (!ctx) {
					resolve(null);
					return;
				}
				// Draw scaled image — this acts like ColorThief's quantization via downscale
				ctx.drawImage(img, 0, 0, size, size);
				const data = ctx.getImageData(0, 0, size, size).data;

				let bestScore = -1;
				let best: [number, number, number] = [255, 255, 255];
				let rSum = 0, gSum = 0, bSum = 0, count = 0;

				for (let i = 0; i < data.length; i += 4) {
					const r = data[i];
					const g = data[i + 1];
					const b = data[i + 2];
					const a = data[i + 3];
					if (a < 128) continue;
					// Skip near black/white extremes that make muddy spotlight
					const max = Math.max(r, g, b);
					const min = Math.min(r, g, b);
					if (max < 20 && min < 20) continue; // near black
					if (max > 250 && min > 230) continue; // near white

					const score = getVibrantScore(r, g, b);
					if (score > bestScore) {
						bestScore = score;
						best = [r, g, b];
					}
					// Also accumulate for fallback average
					rSum += r;
					gSum += g;
					bSum += b;
					count++;
				}

				let result: [number, number, number];
				if (bestScore > 0) {
					result = best;
				} else if (count > 0) {
					result = [Math.round(rSum / count), Math.round(gSum / count), Math.round(bSum / count)];
				} else {
					result = [255, 255, 255];
				}

				cache.set(src, result);
				resolve(result);
			} catch {
				resolve(null);
			}
		};

		img.onerror = () => {
			clearTimeout(timeout);
			resolve(null);
		};

		img.src = src;
	});
}

export function clearDominantColorCache() {
	cache.clear();
}
