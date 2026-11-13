import { describe, expect, it } from "vitest";
import { isDirectLink } from "./addLink";

describe("isDirectLink", () => {
	it("accepts http and https links, ignoring spaces", () => {
		expect(isDirectLink("https://example.test/example.zip")).toBe(true);
		expect(isDirectLink("  http://example.test/a  ")).toBe(true);
	});

	it("rejects other text and schemes", () => {
		expect(isDirectLink("")).toBe(false);
		expect(isDirectLink("example.zip")).toBe(false);
		expect(isDirectLink("ftp://example.test/a")).toBe(false);
	});
});
