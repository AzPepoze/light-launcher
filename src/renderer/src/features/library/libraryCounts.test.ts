import { describe, expect, it } from "vitest";
import type { LibraryStats } from "../../lib/features/types";
import { librarySubtitle, updateCountOf } from "./libraryCounts";

const STATS: LibraryStats = { total: 40, installed: 12, updates: 0, bytes: 0, playtime: 0 };

describe("library counts", () => {
	it("leaves the updates part out when no game is behind", () => {
		expect(librarySubtitle(STATS)).toBe("40 games · 12 installed");
	});

	it("names the games behind their installed version", () => {
		expect(librarySubtitle({ ...STATS, updates: 23 })).toBe("40 games · 12 installed · 23 updates");
	});

	it("gives the Updates badge the same count as the header", () => {
		const stats = { ...STATS, updates: 23 };
		expect(updateCountOf(stats)).toBe(23);
		expect(librarySubtitle(stats)).toContain(`${updateCountOf(stats)} updates`);
	});
});
