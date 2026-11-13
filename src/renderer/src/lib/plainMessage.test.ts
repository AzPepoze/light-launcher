import { describe, expect, it } from "vitest";
import { plainMessage } from "./plainMessage";

describe("plainMessage", () => {
	it("strips RPC error codes", () => {
		expect(plainMessage("API key is not configured (code -32603)")).toBe(
			"API key is not configured",
		);
		expect(plainMessage("Host unreachable (error 12)")).toBe("Host unreachable");
	});

	it("reads the message of an Error", () => {
		expect(plainMessage(new Error("Disk is full (code -1)"))).toBe("Disk is full");
	});

	it("reads the message of an object with a message field", () => {
		expect(plainMessage({ message: "Prefix is busy" })).toBe("Prefix is busy");
	});

	it("removes the wrapper text of a remote call error", () => {
		expect(
			plainMessage("Error invoking remote method 'LaunchGame': Error: Missing exe (code -32603)"),
		).toBe("Missing exe");
	});

	it("drops raw URLs", () => {
		expect(plainMessage("Download failed from https://example.com/a/b.zip")).toBe(
			"Download failed from",
		);
	});

	it("keeps only the first line of a stack trace", () => {
		expect(plainMessage("Could not save\n    at save (app.js:10:3)")).toBe("Could not save");
	});

	it("uses the fallback when nothing is left", () => {
		expect(plainMessage("(code -32603)")).toBe("Something went wrong");
		expect(plainMessage(null, "Save failed")).toBe("Save failed");
		expect(plainMessage(undefined)).toBe("Something went wrong");
	});
});
