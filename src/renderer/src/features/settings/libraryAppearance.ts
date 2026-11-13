import type { AppSettings } from "@shared";

/** Library card width limits and default, in px. */
export const LIBRARY_CARD_SIZE_MIN = 100;
export const LIBRARY_CARD_SIZE_MAX = 260;
export const LIBRARY_CARD_SIZE_DEFAULT = 160;

/** Keeps a card width in range and rounds it to whole px. */
export function clampLibraryCardSize(size: number): number {
	if (!Number.isFinite(size)) return LIBRARY_CARD_SIZE_DEFAULT;
	return Math.min(LIBRARY_CARD_SIZE_MAX, Math.max(LIBRARY_CARD_SIZE_MIN, Math.round(size)));
}

/** The Library card width in px, or the default when the setting is missing. */
export function libraryCardSize(settings: AppSettings): number {
	const size = settings.LibraryCardSize;
	return typeof size === "number" ? clampLibraryCardSize(size) : LIBRARY_CARD_SIZE_DEFAULT;
}

/** Whether Library covers stay blurred until hovered or focused. Missing means off. */
export function blurCoversUntilHover(settings: AppSettings): boolean {
	return settings.BlurCoversUntilHover === true;
}
