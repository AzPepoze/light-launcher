import { beforeEach, describe, expect, it, vi } from "vitest";
import { libraryApi } from "./libraryApi";
import { rpc } from "./rpc";
import type { LibraryQuery } from "./types";
import type { WireGame, WireLibraryList, WireLibraryOverview, WireUpdatesSummary } from "./wire";

vi.mock("./rpc", () => ({ rpc: vi.fn() }));

const rpcMock = vi.mocked(rpc);

const QUERY: LibraryQuery = {
	source: "f95",
	search: "",
	tags: [],
	favorite: false,
	installed: false,
	updates: false,
	sort: "title"
};

/** A game row as the server sends it. Only the fields the mapper reads are filled in. */
function wireGame(id: number, updateAvailable = false): WireGame {
	return {
		id,
		source: "f95",
		url: `https://example.test/${id}`,
		title: `Game ${id}`,
		version: updateAvailable ? "2" : "1",
		installedVersion: "1",
		installDir: null,
		favorite: 0
	} as unknown as WireGame;
}

/** Answers each method the test calls, by name. */
function answerWith(answers: Record<string, unknown>) {
	rpcMock.mockImplementation(async (method: string) => {
		if (!(method in answers)) throw new Error(`unexpected call ${method}`);
		return answers[method] as never;
	});
}

describe("libraryApi.list", () => {
	beforeEach(() => {
		rpcMock.mockReset();
	});

	it("asks for one page with the query, and maps the games and the total", async () => {
		answerWith({ LibraryList: { items: [wireGame(4)], total: 61 } satisfies WireLibraryList });

		const page = await libraryApi.list(QUERY, { offset: 0, limit: 60, wantTotal: true });

		expect(rpcMock).toHaveBeenCalledWith(
			"LibraryList",
			expect.objectContaining({
				source: "f95",
				sort: "title",
				order: "asc",
				offset: 0,
				limit: 60,
				wantTotal: true
			})
		);
		expect(page.items.map((game) => game.id)).toEqual([4]);
		expect(page.total).toBe(61);
	});

	it("does not send a total for a later page the server did not count", async () => {
		answerWith({ LibraryList: { items: [wireGame(70)] } satisfies WireLibraryList });

		const page = await libraryApi.list(QUERY, { offset: 60, limit: 60, wantTotal: false });

		expect(page.total).toBeUndefined();
	});
});

describe("libraryApi.overview", () => {
	beforeEach(() => {
		rpcMock.mockReset();
	});

	it("reads the overview once, with the source tile shapes", async () => {
		answerWith({
			LibraryOverview: {
				sources: [{ id: "f95", label: "F95 Zone", count: 5, tile: "wide" }],
				tags: [],
				stats: { total: 5, installed: 0, updates: 0, bytes: 0, playtime: 0 }
			} satisfies WireLibraryOverview
		});

		const overview = await libraryApi.overview();

		expect(rpcMock).toHaveBeenCalledTimes(1);
		expect(rpcMock).toHaveBeenCalledWith("LibraryOverview");
		expect(overview.sources).toEqual([
			{ source: "f95", label: "F95 Zone", count: 5, tile: "wide" }
		]);
	});
});

describe("libraryApi.syncSource", () => {
	beforeEach(() => {
		rpcMock.mockReset();
	});

	it("starts the import of one source", async () => {
		answerWith({ LibrarySyncSource: { started: true } });

		await libraryApi.syncSource("f95");

		expect(rpcMock).toHaveBeenCalledWith("LibrarySyncSource", { source: "f95" });
	});

	it("passes the server error on", async () => {
		rpcMock.mockRejectedValue(new Error("sync already running"));

		await expect(libraryApi.syncSource("f95")).rejects.toThrow("sync already running");
	});
});

describe("libraryApi.updates", () => {
	beforeEach(() => {
		rpcMock.mockReset();
	});

	it("reads the updates through paged lists with the updates filter on", async () => {
		const firstPage = Array.from({ length: 200 }, (_, index) => wireGame(index + 1, true));
		const secondPage = [wireGame(201, true), wireGame(202)];
		const pages = [{ items: firstPage }, { items: secondPage }];
		rpcMock.mockImplementation(async (method: string) => {
			if (method === "LibraryUpdatesSummary") {
				return { unseen: 2, last: null } satisfies WireUpdatesSummary;
			}
			return pages.shift() as never;
		});

		const updates = await libraryApi.updates();

		const listCalls = rpcMock.mock.calls.filter(([method]) => method === "LibraryList");
		expect(listCalls.map(([, payload]) => (payload as { offset: number }).offset)).toEqual([
			0, 200
		]);
		expect(listCalls.every(([, payload]) => (payload as { updates: boolean }).updates)).toBe(true);
		expect(updates.count).toBe(2);
		expect(updates.games).toHaveLength(201);
	});
});
