const FNV_OFFSET_BASIS = 0x811c9dc5;
const FNV_PRIME = 0x01000193;
const HUE_COUNT = 360;

/** 32-bit FNV-1a hash of the text's UTF-16 code units. */
export function fnv1a(text: string): number {
	let hash = FNV_OFFSET_BASIS;
	for (let index = 0; index < text.length; index += 1) {
		hash ^= text.charCodeAt(index);
		hash = Math.imul(hash, FNV_PRIME) >>> 0;
	}
	return hash;
}

/** Hue in degrees for an identity tile. The same id always gets the same hue. */
export function identityHue(id: string): number {
	return fnv1a(id) % HUE_COUNT;
}

/** The first letter of a name, upper case, for the identity tile. */
export function identityInitial(name: string): string {
	return Array.from(name.trim())[0]?.toUpperCase() ?? "?";
}
