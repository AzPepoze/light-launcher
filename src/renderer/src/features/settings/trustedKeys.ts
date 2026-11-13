/** The size of an ed25519 public key, as the host enforces it. */
export const ED25519_PUBLIC_KEY_BYTES = 32;

/** The label shown for one trusted key in the Settings list. */
export function fingerprintLabel(fingerprint: string): string {
	const value = fingerprint.trim();
	return value ? `SHA256 ${value}` : "SHA256 fingerprint unavailable";
}

/**
 * Checks a pasted key before the call, so obvious mistakes get a local message.
 * The host still validates the key; this only catches what a person can see at a glance.
 */
export function trustedKeyInputError(input: string): string | null {
	const key = input.trim();
	if (!key) return "Paste a base64 ed25519 public key.";
	let decoded: string;
	try {
		decoded = atob(key);
	} catch {
		return "That is not valid base64.";
	}
	if (decoded.length !== ED25519_PUBLIC_KEY_BYTES) {
		return `An ed25519 public key is 32 bytes; that one is ${decoded.length}.`;
	}
	return null;
}
