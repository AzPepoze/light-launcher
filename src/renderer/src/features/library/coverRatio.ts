/** The cover box used when a source sends no usable ratio. */
export const DEFAULT_COVER_RATIO = 3 / 4;

const RATIO_PATTERN = /^(\d+):(\d+)$/;
const MAX_TERM = 32;
const MIN_RATIO = 0.25;
const MAX_RATIO = 4;

/** Width divided by height for a "W:H" string. Null when the text is not a valid ratio. */
export function parseCoverRatio(text: string | null | undefined): number | null {
	const match = RATIO_PATTERN.exec(text?.trim() ?? "");
	if (!match) return null;
	const width = Number(match[1]);
	const height = Number(match[2]);
	if (width < 1 || height < 1 || width > MAX_TERM || height > MAX_TERM) return null;
	const ratio = width / height;
	return ratio >= MIN_RATIO && ratio <= MAX_RATIO ? ratio : null;
}

/** The ratio to draw a cover with. Missing or invalid text uses the default. */
export function coverRatioOf(text: string | null | undefined): number {
	return parseCoverRatio(text) ?? DEFAULT_COVER_RATIO;
}

/** Wide covers get this many times the width of a portrait card, so they read as large as the rest. */
const LANDSCAPE_WIDTH_FACTOR = 1.6;

/** The smallest column width for a cover. Landscape covers are scaled up so they do not shrink. */
export function tileMinWidth(cardSize: number, ratio: number): number {
	return ratio > 1 ? Math.round(cardSize * LANDSCAPE_WIDTH_FACTOR) : cardSize;
}
