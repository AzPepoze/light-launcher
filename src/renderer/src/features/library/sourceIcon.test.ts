import { describe, expect, it } from "vitest";
import { LOCAL_SOURCE } from "./librarySections";
import { sourceIcon } from "./sourceIcon";

describe("sourceIcon", () => {
	it("uses a folder icon for local games", () => {
		expect(sourceIcon(LOCAL_SOURCE)).toBe("folder_open");
	});

	it("uses the generic icon for any plugin source", () => {
		expect(sourceIcon("example")).toBe("store");
		expect(sourceIcon("constructor")).toBe("store");
	});
});
