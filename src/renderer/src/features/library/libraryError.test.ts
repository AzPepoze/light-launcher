import { describe, expect, it } from "vitest";
import { libraryErrorMessage } from "./libraryError";

describe("libraryErrorMessage", () => {
	it("replaces a missing plugin method with plain words", () => {
		expect(libraryErrorMessage(new Error("method not found: LibraryBrowse"))).toBe(
			"This source does not support this yet."
		);
	});

	it("keeps other error messages", () => {
		expect(libraryErrorMessage(new Error("network down"))).toBe("network down");
	});
});
