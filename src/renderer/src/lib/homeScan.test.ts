import { describe, expect, it } from "vitest";
import { toHomeScan } from "./homeScan";

describe("toHomeScan", () => {
	it("maps a running payload to progress", () => {
		expect(toHomeScan({ state: "running", done: 2, total: 5 })).toEqual({
			state: "running",
			done: 2,
			total: 5,
			current: ""
		});
	});

	it("keeps the current folder name", () => {
		expect(toHomeScan({ state: "running", done: 1, total: 3, current: "Games" })).toEqual({
			state: "running",
			done: 1,
			total: 3,
			current: "Games"
		});
	});

	it("treats a blank current as empty", () => {
		expect(toHomeScan({ state: "running", done: 1, total: 3, current: "   " })).toEqual({
			state: "running",
			done: 1,
			total: 3,
			current: ""
		});
		expect(toHomeScan({ state: "running", done: 1, total: 3, current: 7 })).toEqual({
			state: "running",
			done: 1,
			total: 3,
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
