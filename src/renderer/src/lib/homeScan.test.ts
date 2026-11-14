import { describe, expect, it } from "vitest";
import { toHomeScan, toScanGroup } from "./homeScan";

describe("toHomeScan", () => {
	it("maps a running payload to progress", () => {
		expect(toHomeScan({ state: "running", done: 2, total: 5, found: 7 })).toEqual({
			state: "running",
			done: 2,
			total: 5,
			found: 7,
			current: ""
		});
	});

	it("defaults a missing found count to zero", () => {
		expect(toHomeScan({ state: "running", done: 2, total: 5 })).toEqual({
			state: "running",
			done: 2,
			total: 5,
			found: 0,
			current: ""
		});
		expect(toHomeScan({ state: "running", done: 2, total: 5, found: "7" })).toEqual({
			state: "running",
			done: 2,
			total: 5,
			found: 0,
			current: ""
		});
	});

	it("keeps the current folder name", () => {
		expect(toHomeScan({ state: "running", done: 1, total: 3, found: 4, current: "Games" })).toEqual(
			{
				state: "running",
				done: 1,
				total: 3,
				found: 4,
				current: "Games"
			}
		);
	});

	it("treats a blank current as empty", () => {
		expect(toHomeScan({ state: "running", done: 1, total: 3, found: 0, current: "   " })).toEqual({
			state: "running",
			done: 1,
			total: 3,
			found: 0,
			current: ""
		});
		expect(toHomeScan({ state: "running", done: 1, total: 3, found: 0, current: 7 })).toEqual({
			state: "running",
			done: 1,
			total: 3,
			found: 0,
			current: ""
		});
	});

	it("maps a done payload", () => {
		expect(toHomeScan({ state: "done", done: 5, total: 5, current: "Games" })).toEqual({
			state: "done"
		});
	});

	it("rejects payloads that are not objects", () => {
		expect(toHomeScan(null)).toBeNull();
		expect(toHomeScan("running")).toBeNull();
	});

	it("rejects an unknown state", () => {
		expect(toHomeScan({ state: "paused", done: 1, total: 2 })).toBeNull();
	});

	it("rejects a running payload without numeric counts", () => {
		expect(toHomeScan({ state: "running", done: "1", total: 2 })).toBeNull();
		expect(toHomeScan({ state: "running" })).toBeNull();
	});
});

describe("toScanGroup", () => {
	it("reads the finished folder from the event", () => {
		const game = { name: "Alpha", path: "C:/Games/Alpha/alpha.exe" };
		expect(
			toScanGroup({
				state: "running",
				done: 1,
				total: 2,
				found: 1,
				group: { folderPath: "C:/Games", folderName: "Games", games: [game] }
			})
		).toEqual({ folderPath: "C:/Games", folderName: "Games", games: [game] });
	});

	it("treats a null games list as empty", () => {
		expect(
			toScanGroup({ state: "running", group: { folderPath: "D:/Lib", folderName: "Lib", games: null } })
		).toEqual({ folderPath: "D:/Lib", folderName: "Lib", games: [] });
	});

	it("returns null when the event has no group", () => {
		expect(toScanGroup({ state: "running", done: 1, total: 2, found: 0 })).toBeNull();
		expect(toScanGroup({ state: "running", group: null })).toBeNull();
	});

	it("ignores a group without string folder fields", () => {
		expect(toScanGroup({ group: { folderName: "Lib", games: [] } })).toBeNull();
		expect(toScanGroup({ group: { folderPath: 4, folderName: "Lib", games: [] } })).toBeNull();
		expect(toScanGroup({ group: { folderPath: "D:/Lib", games: [] } })).toBeNull();
	});

	it("ignores payloads that are not objects", () => {
		expect(toScanGroup(null)).toBeNull();
		expect(toScanGroup("group")).toBeNull();
	});
});
