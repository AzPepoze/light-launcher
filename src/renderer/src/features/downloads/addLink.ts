/** True for a web link that starts with http:// or https://. */
export function isDirectLink(text: string): boolean {
	try {
		const url = new URL(text.trim());
		return url.protocol === "http:" || url.protocol === "https:";
	} catch {
		return false;
	}
}
