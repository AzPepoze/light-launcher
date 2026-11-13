import { describe, expect, it } from "vitest";
import { ED25519_PUBLIC_KEY_BYTES, fingerprintLabel, trustedKeyInputError } from "./trustedKeys";

// Standard base64 of the 32 bytes 0x00..0x1f.
const VALID_KEY = "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=";

describe("fingerprintLabel", () => {
	it("prefixes the short fingerprint with the hash name", () => {
		expect(fingerprintLabel("0123456789abcdef")).toBe("SHA256 0123456789abcdef");
	});

	it("trims the fingerprint it was given", () => {
		expect(fingerprintLabel("  0123456789abcdef ")).toBe("SHA256 0123456789abcdef");
	});

	it("says so when the host sends no fingerprint", () => {
		expect(fingerprintLabel("")).toBe("SHA256 fingerprint unavailable");
		expect(fingerprintLabel("   ")).toBe("SHA256 fingerprint unavailable");
	});
});

describe("trustedKeyInputError", () => {
	it("accepts a 32-byte base64 key", () => {
		expect(atob(VALID_KEY).length).toBe(ED25519_PUBLIC_KEY_BYTES);
		expect(trustedKeyInputError(VALID_KEY)).toBeNull();
		expect(trustedKeyInputError(`  ${VALID_KEY}\n`)).toBeNull();
	});

	it("asks for a key when the input is empty", () => {
		expect(trustedKeyInputError("")).toBe("Paste a base64 ed25519 public key.");
		expect(trustedKeyInputError("   \n\t")).toBe("Paste a base64 ed25519 public key.");
	});

	it("rejects text that is not base64", () => {
		expect(trustedKeyInputError("not a key!")).toBe("That is not valid base64.");
	});

	it("reports a key of the wrong size", () => {
		const short = btoa(String.fromCharCode(...new Uint8Array(ED25519_PUBLIC_KEY_BYTES - 1)));
		const long = btoa(String.fromCharCode(...new Uint8Array(ED25519_PUBLIC_KEY_BYTES + 1)));
		expect(trustedKeyInputError(short)).toBe("An ed25519 public key is 32 bytes; that one is 31.");
		expect(trustedKeyInputError(long)).toBe("An ed25519 public key is 32 bytes; that one is 33.");
	});
});
